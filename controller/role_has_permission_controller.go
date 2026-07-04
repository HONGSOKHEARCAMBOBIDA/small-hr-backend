package controller

import (
	"mysql/constant/share"
	"mysql/request"
	"mysql/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHasPermissionController struct {
	service service.RoleService
}

func NewRoleHasPermissionController() RoleHasPermissionController {
	return RoleHasPermissionController{
		service: service.NewRoleService(),
	}
}

func (cr *RoleHasPermissionController) CreateRoleHasPermission(c *gin.Context) {
	var input request.CreateRolePermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.CreateRoleHasPermission(c, input); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "Created")
}

func (cr *RoleHasPermissionController) DeleteRoleHasPermission(c *gin.Context) {
	var input request.DeleteRolePermissionsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.DeleteRoleHasPermission(c, input); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "Deleted")
}

func (cr *RoleHasPermissionController) GetRolePermission(c *gin.Context) {
	idparam := c.Param("id")
	id, err := strconv.Atoi(idparam)
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := cr.service.GetRolePermission(c, id)
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, data)

}
