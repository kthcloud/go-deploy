package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	roleModel "go-deploy/models/sys/role"
	userModel "go-deploy/models/sys/user"
	"go-deploy/pkg/config"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/deployment_service"
	"go-deploy/service/storage_manager_service"
	smClient "go-deploy/service/storage_manager_service/client"
	"go-deploy/service/user_service"
	"go-deploy/service/vm_service"
	"go-deploy/utils"
)

func collectUsage(userID string) (*userModel.Usage, error) {
	vmUsage, err := vm_service.New().GetUsage(userID)
	if err != nil {
		return nil, err
	}

	deploymentUsage, err := deployment_service.New().WithUserID(userID).GetUsage()
	if err != nil {
		return nil, err
	}

	usage := &userModel.Usage{
		Deployments: deploymentUsage.Count,
		CpuCores:    vmUsage.CpuCores,
		RAM:         vmUsage.RAM,
		DiskSize:    vmUsage.DiskSize,
	}

	return usage, nil
}

func getStorageURL(userID string, auth *service.AuthInfo) (*string, error) {
	storageManager, err := storage_manager_service.New().WithUserID(userID).WithAuth(auth).Get(&smClient.GetOptions{})
	if err != nil {
		return nil, err
	}

	var storageURL *string
	if storageManager != nil {
		storageURL = storageManager.GetURL()
	}

	return storageURL, nil
}

// ListUsers
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
func ListUsers(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.UserList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	if requestQuery.Discover {
		users, err := user_service.DiscoverAuth(requestQuery.Search, auth)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if users == nil {
			context.Ok([]interface{}{})
			return
		}

		context.Ok(users)
		return
	}

	users, err := user_service.ListAuth(requestQuery.All, requestQuery.Search, auth, &requestQuery.Pagination)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	usersDto := make([]body.UserRead, 0)
	for _, user := range users {
		// if we list ourselves, take the opportunity to update our role
		if user.ID == auth.UserID {
			updatedUser, err := user_service.Create(auth)
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

		role := config.Config.GetRole(user.EffectiveRole.Name)

		usage, _ := collectUsage(user.ID)
		if usage == nil {
			usage = &userModel.Usage{}
		}

		usersDto = append(usersDto, user.ToDTO(role, usage, storageURL))
	}

	context.Ok(usersDto)
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
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	if requestURI.UserID == "" {
		requestURI.UserID = auth.UserID
	}

	var effectiveRole *roleModel.Role
	var user *userModel.User

	if requestURI.UserID == auth.UserID {
		effectiveRole = auth.GetEffectiveRole()
		user, err = user_service.Create(auth)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	} else {
		effectiveRole = config.Config.GetRole(user.EffectiveRole.Name)
		user, err = user_service.GetAuth(requestURI.UserID, auth)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	}

	if user == nil {
		context.NotFound("User not found")
		return
	}

	usage, err := collectUsage(user.ID)
	if usage == nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	storageURL, err := getStorageURL(user.ID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
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
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var userUpdate body.UserUpdate
	if err := context.GinContext.ShouldBindJSON(&userUpdate); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	if requestURI.UserID == "" {
		requestURI.UserID = auth.UserID
	}

	var effectiveRole *roleModel.Role

	if requestURI.UserID == auth.UserID {
		effectiveRole = auth.GetEffectiveRole()
		_, err = user_service.Create(auth)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	}

	updated, err := user_service.UpdatedAuth(requestURI.UserID, &userUpdate, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if updated == nil {
		context.NotFound("User not found")
		return
	}

	usage, err := collectUsage(updated.ID)
	if usage == nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	storageURL, err := getStorageURL(updated.ID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.JSONResponse(200, updated.ToDTO(effectiveRole, usage, storageURL))
}
