package routes

import (
	"mysql/constant/permission"
	"mysql/constant/route"
	"mysql/controller"
	"mysql/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	authcontroller := controller.NewAuthController()
	companycontroller := controller.NewCompanyController()
	shiftcontroller := controller.NewShiftController()
	attendancecontroller := controller.NewAttendanceController()
	payrollcontroller := controller.NewPayrollController()
	backupcontroller := controller.NewBackupController()
	leavededucttypecontroller := controller.NewLeaveDeductTypeController()
	leavetypecontroller := controller.NewLeaveTypeController()
	leaverequestcontroller := controller.NewLeaveRequestController()
	rolehaspermissioncontroller := controller.NewRoleHasPermissionController()
	r.Static("/clientimage", "./public/clientimage")
	public := r.Group("/")
	public.Use(middleware.APIKeyAuth())
	{
		public.POST(route.Login, authcontroller.Login)
		public.POST(route.LoginByQr, authcontroller.LoginByQr)
		public.POST(route.Refresh, authcontroller.Refresh)
	}
	auth := r.Group("/")
	auth.Use(middleware.APIKeyAuth())
	auth.Use(middleware.AuthMiddleware())
	{
		// Company
		auth.GET(route.ViewCompany, middleware.PermissionMiddleware(permission.ViewCompany), companycontroller.GetCompany)
		auth.POST(route.AddCompany, middleware.PermissionMiddleware(permission.AddCompany), companycontroller.CreateCompany)
		auth.PUT(route.EditCompany, middleware.PermissionMiddleware(permission.EditCompany), companycontroller.UpdateCompany)
		auth.PUT(route.EditTelegram, middleware.PermissionMiddleware(permission.EditCompany), companycontroller.UpdateTelegram)
		auth.GET(route.ViewManageCompany, middleware.PermissionMiddleware(permission.ViewCompany), companycontroller.ShowManageCompany)
		auth.GET(route.ViewCompanyColor, middleware.PermissionMiddleware(permission.ViewCompany), companycontroller.GetCompanyColor)

		// User
		auth.POST(route.AddUser, middleware.PermissionMiddleware(permission.AddUser), authcontroller.Register)
		auth.GET(route.ViewUser, middleware.PermissionMiddleware(permission.ViewUser), authcontroller.GetUser)
		auth.PUT(route.ToggleUserStatus, middleware.PermissionMiddleware(permission.EditUser), authcontroller.ToggleUserStatus)
		auth.PUT(route.ChangePassword, middleware.PermissionMiddleware(permission.ChangePassword), authcontroller.ChangePassword)
		auth.PUT(route.EditUser, middleware.PermissionMiddleware(permission.EditUser), authcontroller.UpdateUser)
		auth.GET(route.CountUser, middleware.PermissionMiddleware(permission.ViewUser), authcontroller.CountUser)
		auth.GET(route.ViewRole, middleware.PermissionMiddleware(permission.ViewUser), authcontroller.GetRole)
		//auth.DELETE(route.DeleteUser, middleware.PermissionMiddleware(permission.EditUser), authcontroller.DeleteUser)
		auth.GET(route.ViewUserData, middleware.PermissionMiddleware(permission.ViewUser), authcontroller.GetUserData)
		auth.GET(route.ViewUserApprove, middleware.PermissionMiddleware(permission.ViewUser), authcontroller.GetUserApprove)
		auth.PUT(route.VerifyUser, middleware.PermissionMiddleware(permission.EditUser), authcontroller.VerifyUser)

		// Shift
		auth.PUT(route.EditShift, middleware.PermissionMiddleware(permission.EditUser), shiftcontroller.UpdateShift)
		auth.POST(route.AddShift, middleware.PermissionMiddleware(permission.EditUser), shiftcontroller.CreateShift)

		// Attendance
		auth.POST(route.AddAttendance, middleware.PermissionMiddleware(permission.AddAttendance), attendancecontroller.CreateAttendance)
		auth.GET(route.ViewAttendance, middleware.PermissionMiddleware(permission.ViewAttendance), attendancecontroller.GetAttendance)
		auth.GET(route.ViewAttendanceDraft, middleware.PermissionMiddleware(permission.ViewAttendance), attendancecontroller.GetAttendanceDraft)
		auth.GET(route.GenerateAttendancePDF, middleware.PermissionMiddleware(permission.ViewAttendance), attendancecontroller.GetAttendancePDF)
		auth.DELETE(route.DeleteAttendance, middleware.PermissionMiddleware(permission.DeleteBackup), attendancecontroller.DeleteAttendance)

		// Payroll
		auth.GET(route.ViewPayrollDraft, middleware.PermissionMiddleware(permission.ViewPayroll), payrollcontroller.GetDraftPayroll)
		auth.POST(route.AddPayroll, middleware.PermissionMiddleware(permission.AddPayroll), payrollcontroller.CreatePayroll)
		auth.GET(route.ViewPayroll, middleware.PermissionMiddleware(permission.ViewPayroll), payrollcontroller.GetPayroll)
		auth.POST(route.DeletePayroll, middleware.PermissionMiddleware(permission.EditPayroll), payrollcontroller.DeletePayroll)

		// Backup
		auth.POST(route.CreateBackup, middleware.PermissionMiddleware(permission.CreateBackup), backupcontroller.TriggerBackup)
		auth.GET(route.ViewBackup, middleware.PermissionMiddleware(permission.ViewBackup), backupcontroller.ListBackups)
		auth.GET(route.DownloadBackup, middleware.PermissionMiddleware(permission.DownloadBackup), backupcontroller.DownloadBackup)
		auth.DELETE(route.DeleteBackup, middleware.PermissionMiddleware(permission.DeleteBackup), backupcontroller.DeleteBackup)

		// LeaveDeductType
		auth.GET(route.ViewLeaveDeductType, middleware.PermissionMiddleware(permission.ViewLeaveDeductType), leavededucttypecontroller.GetLeaveDeductType)

		// LeaveType
		auth.GET(route.ViewLeave, middleware.PermissionMiddleware(permission.ViewLeave), leavetypecontroller.GetLeaveTypes)
		auth.POST(route.AddLeaveType, middleware.PermissionMiddleware(permission.AddLeaveType), leavetypecontroller.CreateLeaveType)
		auth.PUT(route.EditLeaveType, middleware.PermissionMiddleware(permission.EditLeaveType), leavetypecontroller.UpdateLeaveType)

		// LeaveRequest
		auth.GET(route.ViewLeaveRequest, middleware.PermissionMiddleware(permission.ViewLeaveRequest), leaverequestcontroller.GetLeaveRequest)
		auth.POST(route.AddLeaveRequest, middleware.PermissionMiddleware(permission.AddLeaveRequest), leaverequestcontroller.CreateLeaveRequest)
		auth.PUT(route.EditLeaveRequest, middleware.PermissionMiddleware(permission.EditLeaveRequest), leaverequestcontroller.UpdateLeaveRequest)
		auth.PUT(route.EditStatusLeaveRequest, middleware.PermissionMiddleware(permission.EditStatusLeaveRequest), leaverequestcontroller.UpdateStatusLeaveRequest)
		auth.DELETE(route.DeleteLeaveRequest, middleware.PermissionMiddleware(permission.EditStatusLeaveRequest), leaverequestcontroller.DeleteLeaveRequest)

		// RoleHasPermission
		auth.GET(route.ViewRoleHasPermission, middleware.PermissionMiddleware(permission.ViewRoleHasPermission), rolehaspermissioncontroller.GetRolePermission)
		auth.POST(route.AddRoleHasPermission, middleware.PermissionMiddleware(permission.AddRoleHasPermission), rolehaspermissioncontroller.CreateRoleHasPermission)
		auth.DELETE(route.DeleteRoleHasPermission, middleware.PermissionMiddleware(permission.DeleteRoleHasPermission), rolehaspermissioncontroller.DeleteRoleHasPermission)
		auth.PUT(route.EditRole, middleware.PermissionMiddleware(permission.ViewRoleHasPermission), rolehaspermissioncontroller.UpdateRole)
	}
}
