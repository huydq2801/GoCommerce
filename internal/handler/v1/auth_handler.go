package v1handler

import (
	"github.com/gin-gonic/gin"

	v1dto "huydq2801.com/golearn/internal/dto/v1"
	"huydq2801.com/golearn/internal/middleware"
	v1service "huydq2801.com/golearn/internal/service/v1"
	"huydq2801.com/golearn/internal/utils"
)

type AuthHandler struct{ service *v1service.AuthService }

func NewAuthHandler(s *v1service.AuthService) *AuthHandler { return &AuthHandler{service: s} }
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email,max=150"`
		Password string `json:"password" binding:"required,max=72"`
	}
	if !utils.DecodeJSON(c, &input) {
		return
	}
	result, err := h.service.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, result)
}
func (h *AuthHandler) Me(c *gin.Context) {
	utils.ResponseSuccess(c, 200, v1dto.MapUserToDTO(middleware.CurrentUser(c)))
}
