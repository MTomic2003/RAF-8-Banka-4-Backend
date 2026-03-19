package dto

type ListClientsQuery struct {
	Email     string `form:"email"`
	FirstName string `form:"first_name"`
	LastName  string `form:"last_name"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}
