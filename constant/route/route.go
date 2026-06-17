package route

const (

	// Authentication
	Login     = "login"
	LoginByQr = "loginbyqr"
	Refresh   = "refresh"

	// Company
	AddCompany   = "add.company"
	ViewCompany  = "view.company"
	EditCompany  = "edit.company/:id"
	EditTelegram = "edit.telegram/:id"

	// User
	AddUser          = "add.user"
	ViewUser         = "view.user"
	EditUser         = "edit.user/:id"
	ToggleUserStatus = "toggle.status.user/:id"
	ChangePassword   = "change.password"
	CountUser        = "count.user"

	// Shift
	EditShift = "edit.shift"
	AddShift  = "add.shift"

	// Attendance
	AddAttendance       = "add.attendance"
	ViewAttendance      = "view.attendance"
	ViewAttendanceDraft = "view.attendance.draft"
	EditAttendance      = "edit.attendance"

	// Payroll
	AddPayroll       = "add.payroll"
	ViewPayroll      = "view.payroll"
	ViewPayrollDraft = "view.payroll.draft"
	EditPayroll      = "edit.payroll"

	// Role
	ViewRole = "view.role"

	// BackUp
	CreateBackup   = "add.backup"
	ViewBackup     = "view.backup"
	DownloadBackup = "view.download.backup"
	DeleteBackup   = "delete.backup"
)
