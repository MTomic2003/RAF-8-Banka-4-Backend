package dto

type ListActuariesQuery struct {
	Email        string `form:"email"`
	FirstName    string `form:"first_name"`
	LastName     string `form:"last_name"`
	Position     string `form:"position"`
	Department   string `form:"department"`
	Type         string `form:"type" binding:"omitempty,oneof=agent supervisor"`
	Active       *bool  `form:"active"`
	NeedApproval *bool  `form:"need_approval"`
	Page         int    `form:"page" binding:"min=1"`
	PageSize     int    `form:"page_size" binding:"min=1,max=100"`
}
