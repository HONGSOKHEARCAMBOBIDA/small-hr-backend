package controller

import (
	"mysql/constant/share"
	"mysql/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LeaveDeductTypeController struct {
	service service.LeaveDeductTypeService
}

func NewLeaveDeductTypeController() LeaveDeductTypeController {
	return LeaveDeductTypeController{
		service: service.NewLeaveDeductTypeService(),
	}
}

func (cr *LeaveDeductTypeController) GetLeaveDeductType(c *gin.Context) {
	data, err := cr.service.GetLeaveDeductType(c)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, data)
}
