package route

const (

	// Authentication
	Login     = "login"
	LoginByQr = "loginbyqr"
	Refresh   = "refresh"

	// Company
	AddCompany  = "add.company"
	ViewCompany = "view.company"
	EditCompany = "edit.company/:id"

	// User
	AddUser          = "add.user"
	ViewUser         = "view.user"
	EditUser         = "edit.user/:id"
	ToggleUserStatus = "toggle.status.user/:id"
	ChangePassword   = "change.password"

	// Shift
	EditShift = "edit.shift"

	// Attendance
	AddAttendance  = "add.attendance"
	ViewAttendance = "view.attendance"
	EditAttendance = "edit.attendance"

	// Payroll
	AddPayroll  = "add.payroll"
	ViewPayroll = "view.payroll"
	EditPayroll = "edit.payroll"
)
