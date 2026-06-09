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
	r.Static("/clientimage", "./public/clientimage")
	r.POST(route.Login, authcontroller.Login)
	r.POST(route.LoginByQr, authcontroller.LoginByQr)
	r.POST(route.Refresh, authcontroller.Refresh)
	auth := r.Group("/")
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

		// Shift
		auth.PUT(route.EditShift, middleware.PermissionMiddleware(permission.EditUser), shiftcontroller.UpdateShift)

		// Attendance
		auth.POST(route.AddAttendance, middleware.PermissionMiddleware(permission.AddAttendance), attendancecontroller.CreateAttendance)
		auth.GET(route.ViewAttendance, middleware.PermissionMiddleware(permission.ViewAttendance), attendancecontroller.GetAttendance)

	}
}
