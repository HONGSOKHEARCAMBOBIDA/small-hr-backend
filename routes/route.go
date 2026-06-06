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
	r.Static("/clientimage", "./public/clientimage")
	r.POST(route.Login, authcontroller.Login)
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

	}
}
