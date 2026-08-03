// api/client.ts

const BASE_URL = import.meta.env.VITE_API_BASE_URL;

if (!BASE_URL) {
  throw new Error("Thiếu VITE_API_BASE_URL trong file .env");
}

interface ApiErrorBody {
  error: string;
}

type GetTokenFn = () => string | null;
type SetTokenFn = (token: string) => void;

let getAccessToken: GetTokenFn = () => null;
let setAccessToken: SetTokenFn = () => {};

// AuthContext sẽ gọi hàm này 1 lần lúc khởi tạo, "cắm" 2 hàm thật vào đây.
// Nhờ vậy client.ts không cần import AuthContext ngược lại (tránh circular
// dependency giữa 2 module phụ thuộc lẫn nhau).
export function registerTokenHandlers(getter: GetTokenFn, setter: SetTokenFn) {
  getAccessToken = getter;
  setAccessToken = setter;
}

async function refreshAccessToken(): Promise<string> {
  const res = await fetch(`${BASE_URL}/auth/refresh`, {
    method: "POST",
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Refresh thất bại — cần đăng nhập lại");
  }

  const data: { access_token: string; expires_in: number } = await res.json();
  setAccessToken(data.access_token);
  return data.access_token;
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const doFetch = async (accessToken: string | null): Promise<Response> => {
    return fetch(`${BASE_URL}${path}`, {
      ...options,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
        ...options.headers,
      },
    });
  };

  let res = await doFetch(getAccessToken());

  // Access token hết hạn → refresh 1 lần → thử lại request gốc.
  // Không refresh nếu chính request gốc đã LÀ /auth/refresh hoặc /auth/login
  // — tránh vòng lặp vô hạn nếu bản thân refresh cũng trả 401.
  const isAuthEndpoint = path.startsWith("/auth/refresh") || path.startsWith("/auth/login");

  if (res.status === 401 && !isAuthEndpoint) {
    const newToken = await refreshAccessToken();
    res = await doFetch(newToken);
  }

  if (!res.ok) {
    const errBody: ApiErrorBody = await res
      .json()
      .catch(() => ({ error: "Lỗi không xác định" }));
    throw new Error(errBody.error);
  }

  // Một số endpoint (VD logout) có thể trả body rỗng — .json() sẽ lỗi nếu
  // parse chuỗi rỗng, nên cần bắt riêng trường hợp này.
  const text = await res.text();
  return (text ? JSON.parse(text) : undefined) as T;
}