package permission

type Permission string

const (
	EmployeeView   Permission = "employee.view"
	EmployeeCreate Permission = "employee.create"
	EmployeeUpdate Permission = "employee.update"
	EmployeeDelete Permission = "employee.delete"

	ClientView   Permission = "client.view"
	ClientUpdate Permission = "client.update"
)

var All = []Permission{
	EmployeeView,
	EmployeeCreate,
	EmployeeUpdate,
	EmployeeDelete,
	ClientView,
	ClientUpdate,
}
