package controller

import (
	"mysql/constant/share"
	"mysql/helper"
	"mysql/request"
	"mysql/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AttendanceController struct {
	service service.AttendanceService
}

func NewAttendanceController() AttendanceController {
	return AttendanceController{
		service: service.NewAttendanceService(),
	}
}

func (cr *AttendanceController) CreateAttendance(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	var input request.AttendanceRequestCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.CreateAttendance(c, userID, input); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "attendance created")
}
