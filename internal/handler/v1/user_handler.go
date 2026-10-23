package v1handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"huydq2801.com/golearn/internal/db/sqlc"
	v1dto "huydq2801.com/golearn/internal/dto/v1"
	v1service "huydq2801.com/golearn/internal/service/v1"
	"huydq2801.com/golearn/internal/utils"
)

type UserHandler struct{ service v1service.UserService }

func NewUserHandler(service v1service.UserService) *UserHandler {
	return &UserHandler{service: service}
}
func parseUUID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		utils.ResponseError(c, utils.NewError("Invalid UUID", utils.ErrCodeBadRequest))
		return uuid.Nil, false
	}
	return id, true
}
func (h *UserHandler) CreateUser(c *gin.Context) {
	var input v1dto.CreateUserInput
	if !utils.DecodeJSON(c, &input) {
		return
	}
	u, err := h.service.CreateUser(c.Request.Context(), input.MapCreateInputToModel())
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Header("Location", "/api/v1/users/"+u.Uuid.String())
	utils.ResponseSuccess(c, http.StatusCreated, v1dto.MapUserToDTO(u))
}
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var p utils.Page
	if err := c.ShouldBindQuery(&p); err != nil {
		utils.ResponseError(c, utils.NewError("Invalid pagination", utils.ErrCodeBadRequest))
		return
	}
	p.Defaults()
	users, total, err := h.service.GetAllUsers(c.Request.Context(), p)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponsePage(c, v1dto.MapUsersToDTOs(users), p, total)
}
func (h *UserHandler) GetUserByUUID(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	u, err := h.service.GetUserByUUID(c.Request.Context(), id)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, v1dto.MapUserToDTO(u))
}
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	var input v1dto.UpdateUserInput
	if !utils.DecodeJSON(c, &input) {
		return
	}
	if input.Name == nil && input.Email == nil && input.Age == nil && input.Password == nil && input.Status == nil && input.Level == nil {
		utils.ResponseError(c, utils.NewError("At least one field is required", utils.ErrCodeBadRequest))
		return
	}
	u, err := h.service.UpdateUser(c.Request.Context(), sqlc.UpdateUserParams{Uuid: id, UserFullname: input.Name, UserEmail: input.Email, UserAge: input.Age, UserPassword: input.Password, UserStatus: input.Status, UserLevel: input.Level})
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseSuccess(c, 200, v1dto.MapUserToDTO(u))
}
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := parseUUID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Status(204)
}
