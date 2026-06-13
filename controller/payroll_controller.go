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

func (cr *PayrollController) GetPayroll(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter := map[string]string{
		"name":         c.Query("name"),
		"payroll_date": c.Query("payroll_date"),
		"payroll_type": c.Query("payroll_type"),
	}

	payrolls, metadata, err := cr.service.GetPayroll(c, userID, request.Pagination{
		Page:     page,
		PageSize: pageSize,
	}, filter)

	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       payrolls,
		"pagination": metadata,
	})
}
