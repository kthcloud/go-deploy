package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	userModel "go-deploy/models/sys/user"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"go-deploy/service/user_service"
	"go-deploy/service/vm_service"
	"net/http"
)

func collectUsage(context *sys.ClientContext, userID string) (bool, *userModel.Usage) {
	vmUsage, err := vm_service.GetUsageByUserID(userID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get quota for user: %s", err))
		return false, nil
	}

	deploymentUsage, err := deployment_service.GetUsageByUserID(userID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get quota for user: %s", err))
		return false, nil
	}

	usage := &userModel.Usage{
		Deployments: deploymentUsage.Count,
		CpuCores:    vmUsage.CpuCores,
		RAM:         vmUsage.RAM,
		DiskSize:    vmUsage.DiskSize,
	}

	return true, usage
}

// GetList
// @Summary Get user list
// @Description Get user list
// @Tags User
// @Accept  json
// @Produce  json
// @Param wantAll query bool false "Want all users"
// @Success 200 {array}  body.UserRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/users [get]
func GetList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.UserList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	quotas := &auth.GetEffectiveRole().Quotas

	if requestQuery.WantAll && auth.IsAdmin {
		users, err := user_service.GetAll()
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		usersDto := make([]body.UserRead, 0)
		for _, user := range users {
			if user.ID == auth.UserID {
				updatedUser, err := user_service.GetOrCreate(auth)
				if err != nil {
					continue
				}

				if updatedUser != nil {
					user = *updatedUser
				}
			}

			ok, usage := collectUsage(&context, user.ID)
			if !ok {
				usage = &userModel.Usage{}
			}

			usersDto = append(usersDto, user.ToDTO(quotas, usage))
		}

		context.JSONResponse(200, usersDto)
		return
	}

	user, err := user_service.GetByID(auth.UserID, auth.UserID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", auth.UserID))
		return
	}

	ok, usage := collectUsage(&context, user.ID)
	if !ok {
		return
	}

	context.JSONResponse(200, user.ToDTO(quotas, usage))
}

// Get
// @Summary Get user by id
// @Description Get user by id
// @Tags User
// @Accept  json
// @Produce  json
// @Param userId path string true "User ID"
// @Success 200 {object}  body.UserRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/users/{userId} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	requestedUserID := requestURI.UserID
	if requestedUserID == "" {
		requestedUserID = auth.UserID
	}

	var user *userModel.User
	if requestedUserID == auth.UserID {
		user, err = user_service.GetOrCreate(auth)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user: %s", err))
			return
		}
	} else {
		user, err = user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user: %s", err))
			return
		}

		if user == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", requestedUserID))
			return
		}
	}

	ok, usage := collectUsage(&context, user.ID)
	if !ok {
		return
	}

	context.JSONResponse(200, user.ToDTO(&auth.GetEffectiveRole().Quotas, usage))
}

func Update(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var userUpdate body.UserUpdate
	if err := context.GinContext.ShouldBindJSON(&userUpdate); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	requestedUserID := requestURI.UserID
	if requestedUserID == "" {
		requestedUserID = auth.UserID
	}

	user, err := user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	err = user_service.Update(requestedUserID, auth.UserID, auth.IsAdmin, &userUpdate)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	updatedUser, err := user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	if updatedUser == nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	ok, usage := collectUsage(&context, updatedUser.ID)
	if !ok {
		return
	}

	context.JSONResponse(200, updatedUser.ToDTO(&auth.GetEffectiveRole().Quotas, usage))
}
