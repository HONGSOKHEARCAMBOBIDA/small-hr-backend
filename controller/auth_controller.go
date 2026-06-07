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

type AuthController struct {
	service service.AuthService
}

func NewAuthController() AuthController {
	return AuthController{
		service: service.NewAuthService(),
	}
}

func (cr *AuthController) Login(c *gin.Context) {
	var input request.AuthRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, 400, err.Error())
		return
	}
	result, err := cr.service.Login(input, c)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	share.RespondDate(c, http.StatusOK, result)
}

func (cr *AuthController) Refresh(c *gin.Context) {
	var input request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := cr.service.RefreshToken(input, c)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, result)
}

func (cr *AuthController) Register(c *gin.Context) {
	var input request.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.Register(c, input, c); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "user create")
}

func (cr *AuthController) GetUser(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter := map[string]string{
		"name":       c.Query("name"),
		"company_id": c.Query("company_id"),
		"role_id":    c.Query("role_id"),
	}

	users, metadata, err := cr.service.GetUser(c, userID, request.Pagination{
		Page:     page,
		PageSize: pageSize,
	}, filter)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       users,
		"pagination": metadata,
	})
}

func (cr *AuthController) ToggleUserStatus(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	idparam := c.Param("id")
	id, err := strconv.Atoi(idparam)
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.ToggleUserStatus(c, userID, id); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "status changed")
}
