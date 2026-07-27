package dto

type CloneVMRequest struct {
	TemplateVMID int    `json:"template_vmid" binding:"required,min=100"`
	NewVMID      int    `json:"new_vmid" binding:"required,min=100"`
	Name         string `json:"name" binding:"required,min=1,max=255"`
	FullClone    bool   `json:"full_clone"`
}
