package v1handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1dto "huydq2801.com/golearn/internal/dto/v1"
	v1service "huydq2801.com/golearn/internal/service/v1"
	"huydq2801.com/golearn/internal/utils"
	"huydq2801.com/golearn/internal/validation"
)

type UserHandler struct {
	service v1service.UserService
}

func NewUserHandler(service v1service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

type GetUserByUuidParam struct {
	UUID string `uri:"uuid" binding:"required"`
}
type GetUsersParam struct {
	Search string `form:"search" binding:"omitempty,min=3,max=50,search"`
	Page   int    `form:"page" binding:"omitempty,min_int=1"`
	Limit  int    `form:"limit" binding:"omitempty,min_int=1,max_int=100"`
}

func (uh *UserHandler) GetAllUsers(ctx *gin.Context) {
	var param GetUsersParam
	if err := ctx.ShouldBind(&param); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}
	if param.Page == 0 {
		param.Page = 1
	}
	if param.Limit == 0 {
		param.Limit = 10
	}
	users, err := uh.service.GetAllUsers(param.Search, param.Page, param.Limit)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	dtoList := v1dto.MapUsersToDTOs(users)
	utils.ResponseSuccess(ctx, http.StatusOK, &dtoList)
}

func (uh *UserHandler) CreateUser(ctx *gin.Context) {
	var input v1dto.CreateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}
	user := input.MapCreateInputToModel()
	createdUser, err := uh.service.CreateUser(user)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	userDTO := v1dto.MapUserToDTO(createdUser)
	utils.ResponseSuccess(ctx, http.StatusCreated, &userDTO)
}

func (uh *UserHandler) GetUserByUUID(ctx *gin.Context) {
	var param GetUserByUuidParam
	if err := ctx.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}
	user, err := uh.service.GetUserByUUID(param.UUID)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	utils.ResponseSuccess(ctx, http.StatusOK, v1dto.MapUserToDTO(user))
}

func (uh *UserHandler) UpdateUser(ctx *gin.Context) {
	var param GetUserByUuidParam
	if err := ctx.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}
	var input v1dto.UpdateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}
	updatedUser, err := uh.service.UpdateUser(param.UUID, input.MapUpdateInputToModel())
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	utils.ResponseSuccess(ctx, http.StatusOK, v1dto.MapUserToDTO(updatedUser))
}

func (uh *UserHandler) DeleteUser(ctx *gin.Context) {
	var param GetUserByUuidParam
	if err := ctx.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}
	err := uh.service.DeleteUser(param.UUID)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	utils.ResponseStatusCode(ctx, http.StatusNoContent)
}
