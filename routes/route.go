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

		// User
		auth.POST(route.AddUser, middleware.PermissionMiddleware(permission.AddUser), authcontroller.Register)
		auth.GET(route.ViewUser, middleware.PermissionMiddleware(permission.ViewUser), authcontroller.GetUser)
		auth.PUT(route.ToggleUserStatus, middleware.PermissionMiddleware(permission.EditUser), authcontroller.ToggleUserStatus)
		auth.PUT(route.ChangePassword, middleware.PermissionMiddleware(permission.EditUser), authcontroller.ChangePassword)
		auth.PUT(route.EditUser, middleware.PermissionMiddleware(permission.EditUser), authcontroller.UpdateUser)
		auth.GET(route.CountUser, middleware.PermissionMiddleware(permission.ViewUser), authcontroller.CountUser)
		auth.GET(route.ViewRole, middleware.PermissionMiddleware(permission.ViewUser), authcontroller.GetRole)

		// Shift
		auth.PUT(route.EditShift, middleware.PermissionMiddleware(permission.EditUser), shiftcontroller.UpdateShift)
		auth.POST(route.AddShift, middleware.PermissionMiddleware(permission.EditUser), shiftcontroller.CreateShift)

		// Attendance
		auth.POST(route.AddAttendance, middleware.PermissionMiddleware(permission.AddAttendance), attendancecontroller.CreateAttendance)
		auth.GET(route.ViewAttendance, middleware.PermissionMiddleware(permission.ViewAttendance), attendancecontroller.GetAttendance)
		auth.GET(route.ViewAttendanceDraft, middleware.PermissionMiddleware(permission.ViewAttendance), attendancecontroller.GetAttendanceDraft)

		// Payroll
		auth.GET(route.ViewPayrollDraft, middleware.PermissionMiddleware(permission.ViewPayroll), payrollcontroller.GetDraftPayroll)
		auth.POST(route.AddPayroll, middleware.PermissionMiddleware(permission.AddPayroll), payrollcontroller.CreatePayroll)
		auth.GET(route.ViewPayroll, middleware.PermissionMiddleware(permission.ViewPayroll), payrollcontroller.GetPayroll)

		// Backup
		auth.POST(route.CreateBackup, middleware.PermissionMiddleware(permission.CreateBackup), backupcontroller.TriggerBackup)
		auth.GET(route.ViewBackup, middleware.PermissionMiddleware(permission.ViewBackup), backupcontroller.ListBackups)
		auth.GET(route.DownloadBackup, middleware.PermissionMiddleware(permission.DownloadBackup), backupcontroller.DownloadBackup)
		auth.DELETE(route.DeleteBackup, middleware.PermissionMiddleware(permission.DeleteBackup), backupcontroller.DeleteBackup)
	}
}
