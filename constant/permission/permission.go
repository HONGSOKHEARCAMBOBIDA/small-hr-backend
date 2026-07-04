package permission

const (
	// Company
	AddCompany  = "add.company"
	ViewCompany = "view.company"
	EditCompany = "edit.company"

	// User
	AddUser        = "add.user"
	ViewUser       = "view.user"
	EditUser       = "edit.user"
	ChangePassword = "change.password"

	// Attendance
	AddAttendance  = "add.attendance"
	ViewAttendance = "view.attendance"
	EditAttendance = "edit.attendance"

	// Payroll
	AddPayroll  = "add.payroll"
	ViewPayroll = "view.payroll"
	EditPayroll = "edit.payroll"

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
	EditLeaveType = "edit.leave.type"

	// LeaveRequest
	ViewLeaveRequest       = "view.leave.request"
	AddLeaveRequest        = "add.leave.request"
	EditLeaveRequest       = "edit.leave.request"
	EditStatusLeaveRequest = "edit.status.leave.request"

	// RoleHasPermission
	ViewRoleHasPermission   = "view.role.has.permission"
	AddRoleHasPermission    = "add.role.has.permission"
	DeleteRoleHasPermission = "delete.role.has.permission"
)
