package v1handler

import (
	"fmt"
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

	utils.ResponseSuccess(ctx, http.StatusOK, "")
}

func (uh *UserHandler) CreateUser(ctx *gin.Context) {
	var input v1dto.CreateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}

	utils.ResponseSuccess(ctx, http.StatusCreated, "")
}

func (uh *UserHandler) GetUserByUUID(ctx *gin.Context) {
	var param GetUserByUuidParam
	if err := ctx.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}

	utils.ResponseSuccess(ctx, http.StatusOK, "")
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

	utils.ResponseSuccess(ctx, http.StatusOK, "")
}

func (uh *UserHandler) DeleteUser(ctx *gin.Context) {
	var param GetUserByUuidParam
	if err := ctx.ShouldBindUri(&param); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationError(err))
		return
	}

	utils.ResponseStatusCode(ctx, http.StatusNoContent)
}
func (uh *UserHandler) CrashApp(ctx *gin.Context) {
	var a []int
	fmt.Println(a[1])
}
