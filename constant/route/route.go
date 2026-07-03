package route

const (

	// Authentication
	Login     = "login"
	LoginByQr = "loginbyqr"
	Refresh   = "refresh"

	// Company
	AddCompany        = "add.company"
	ViewCompany       = "view.company"
	EditCompany       = "edit.company/:id"
	EditTelegram      = "edit.telegram/:id"
	ViewManageCompany = "view.manage.company"
	ViewCompanyColor  = "view.company.color"

	// User
	AddUser          = "add.user"
	ViewUser         = "view.user"
	EditUser         = "edit.user/:id"
	ToggleUserStatus = "toggle.status.user/:id"
	ChangePassword   = "change.password"
	CountUser        = "count.user"
	DeleteUser       = "delete.user/:id"
	ViewUserData     = "view.user.data"
	ViewUserApprove  = "view.user.approve"

	// Shift
	EditShift = "edit.shift"
	AddShift  = "add.shift"

	// Attendance
	AddAttendance       = "add.attendance"
	ViewAttendance      = "view.attendance"
	ViewAttendanceDraft = "view.attendance.draft"
	EditAttendance      = "edit.attendance"

	// Payroll
	AddPayroll            = "add.payroll"
	ViewPayroll           = "view.payroll"
	ViewPayrollDraft      = "view.payroll.draft"
	EditPayroll           = "edit.payroll"
	DeletePayroll         = "delete.payroll"
	GenerateAttendancePDF = "generate.attendance.pdf"

	// Role
	ViewRole = "view.role"

	// BackUp
	CreateBackup   = "add.backup"
	ViewBackup     = "view.backup"
	DownloadBackup = "view.download.backup"
	DeleteBackup   = "delete.backup"

	// LeaveDeduction
	ViewLeaveDeductType = "view.leave.deduct.type"

	// LeaveType
	ViewLeave     = "view.leave.type"
	AddLeaveType  = "add.leave.type"
	EditLeaveType = "edit.leave.type/:id"

	// LeaveRequest
	ViewLeaveRequest       = "view.leave.request"
	AddLeaveRequest        = "add.leave.request"
	EditLeaveRequest       = "edit.leave.request/:id"
	EditStatusLeaveRequest = "edit.status.leave.request/:id"
)
