package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	roleModel "go-deploy/models/sys/enviroment/role"
	userModel "go-deploy/models/sys/user"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/deployment_service"
	"go-deploy/service/user_service"
	"go-deploy/service/vm_service"
	"go-deploy/utils"
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

func getStorageURL(userID string, auth *service.AuthInfo) (*string, error) {
	storageManager, err := deployment_service.GetStorageManagerByOwnerID(userID, auth)
	if err != nil {
		return nil, err
	}

	var storageURL *string
	if storageManager != nil {
		storageURL = storageManager.GetURL()
	}

	return storageURL, nil
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
// @Router /users [get]
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

	effectiveRole := auth.GetEffectiveRole()

	if requestQuery.All {
		users, err := user_service.GetAll(auth)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		usersDto := make([]body.UserRead, 0)
		for _, user := range users {
			if user.ID == auth.UserID {
				updatedUser, err := user_service.GetOrCreate(auth)
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("failed to get or create a user when listing: %w", err))
					continue
				}

				if updatedUser != nil {
					user = *updatedUser
				}
			}

			storageURL, err := getStorageURL(user.ID, auth)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get storage url for a user when listing: %w", err))
				continue
			}

			role := conf.Env.GetRole(user.EffectiveRole.Name)

			ok, usage := collectUsage(&context, user.ID)
			if !ok {
				usage = &userModel.Usage{}
			}

			usersDto = append(usersDto, user.ToDTO(role, usage, storageURL))
		}

		context.JSONResponse(200, usersDto)
		return
	}

	user, err := user_service.GetByID(auth.UserID, auth)
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

	storageURL, err := getStorageURL(user.ID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get storage url for a user when listing: %s", err))
		return
	}

	context.JSONResponse(200, user.ToDTO(effectiveRole, usage, storageURL))
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
// @Router /users/{userId} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
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
	user, err = user_service.GetByID(requestedUserID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user: %s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	ok, usage := collectUsage(&context, user.ID)
	if !ok {
		return
	}

	var effectiveRole *roleModel.Role
	if user.ID == auth.UserID {
		effectiveRole = auth.GetEffectiveRole()
	} else {
		effectiveRole = conf.Env.GetRole(user.EffectiveRole.Name)
	}

	storageURL, err := getStorageURL(user.ID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get storage url for a user when listing: %s", err))
		return
	}

	context.JSONResponse(200, user.ToDTO(effectiveRole, usage, storageURL))
}

// Update
// @Summary Update user by id
// @Description Update user by id
// @Tags User
// @Accept  json
// @Produce  json
// @Param userId path string true "User ID"
// @Param body body body.UserUpdate true "User update"
// @Success 200 {object} body.UserRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /users/{userId} [post]
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

	user, err := user_service.GetByID(requestedUserID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	err = user_service.Update(requestedUserID, &userUpdate, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	updatedUser, err := user_service.GetByID(requestedUserID, auth)
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

	storageURL, err := getStorageURL(user.ID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get storage url for a user when listing: %s", err))
		return
	}

	context.JSONResponse(200, updatedUser.ToDTO(auth.GetEffectiveRole(), usage, storageURL))
}
