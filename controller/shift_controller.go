package controller

import (
	"mysql/constant/share"
	"mysql/request"
	"mysql/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShiftController struct {
	service service.ShiftService
}

func NewShiftController() ShiftController {
	return ShiftController{
		service: service.NewShiftService(),
	}
}

func (cr *ShiftController) UpdateShift(c *gin.Context) {
	var input request.ShiftRequestUpdate

	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.UpdateShift(c, input); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "Updated shift")
}
