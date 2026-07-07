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

func (cr *AuthController) LoginByQr(c *gin.Context) {
	var input request.LoginQrRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := cr.service.LoginByQr(input, c)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, result)
}

func (cr *AuthController) Refresh(c *gin.Context) {

	// var input request.RefreshTokenRequest
	///	log.Printf(input.RefreshToken)
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := cr.service.RefreshToken(cookie, c)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, result)
}

func (cr *AuthController) Register(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	var input request.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.Register(c, input, c, userID); err != nil {
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
	idparam := c.Param("id")
	id, err := strconv.Atoi(idparam)
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.ToggleUserStatus(c, id, userID); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "status changed")
}

func (cr *AuthController) ChangePassword(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	var input request.NewPasswordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.ChangePassword(c, userID, input); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "password changed")
}

func (cr *AuthController) UpdateUser(c *gin.Context) {
	idparam := c.Param("id")
	id, err := strconv.Atoi(idparam)
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	var input request.UserRequestUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.UpdateUser(c, input, id); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "updated user")
}

func (cr *AuthController) CountUser(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	data, err := cr.service.CountUser(c, userID)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, data)
}

func (cr *AuthController) GetRole(c *gin.Context) {
	userID, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "please login")
		return
	}
	data, err := cr.service.GetRole(c, userID)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, data)
}

func (cr *AuthController) DeleteUser(c *gin.Context) {
	idparam := c.Param("id")
	id, err := strconv.Atoi(idparam)
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get the authenticated actor from context (set by your auth middleware)
	userlog, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "invalid user context")
		return
	}

	if err := cr.service.DeleteUser(c, id, userlog); err != nil {
		share.ResponseError(c, http.StatusForbidden, err.Error())
		return
	}

	share.ResponseSuccess(c, http.StatusOK, "Deleted Success")
}

func (cr *AuthController) GetUserData(c *gin.Context) {
	userlog, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	data, err := cr.service.GetUserData(c, userlog)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, data)
}

func (cr *AuthController) GetUserApprove(c *gin.Context) {
	userlog, ok := helper.GetUserID(c)
	if !ok {
		share.ResponseError(c, http.StatusUnauthorized, "invalid user context")
		return
	}
	data, err := cr.service.GetUserApprove(c, userlog)
	if err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.RespondDate(c, http.StatusOK, data)
}

func (cr *AuthController) VerifyUser(c *gin.Context) {
	idparam := c.Param("id")
	id, err := strconv.Atoi(idparam)
	if err != nil {
		share.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := cr.service.VerifyUser(c, id); err != nil {
		share.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	share.ResponseSuccess(c, http.StatusOK, "User Verify")
}
