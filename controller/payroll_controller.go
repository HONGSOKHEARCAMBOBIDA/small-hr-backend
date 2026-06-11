package controller

import (
	"mysql/constant/share"
	"mysql/helper"
	"mysql/request"
	"mysql/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PayrollController struct {
	service service.PayrollService
}

func NewPayrollController() PayrollController {
	return PayrollController{
		service: service.NewPayrollService(),
	}
}

func (cr *PayrollController) GetDraftPayroll(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	payrolltypeparam := c.Query("payroll_type")
	payrolltype, err := strconv.Atoi(payrolltypeparam)
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	data, err := cr.service.GetDraftPayroll(c, payrolltype, userID)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, data)
}

func (cr *PayrollController) CreatePayroll(c *gin.Context) {
	var input request.PayrollRequestCreate
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.CreatePayroll(c, input); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "payroll create")
}
