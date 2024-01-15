package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	roleModels "go-deploy/models/sys/role"
	userModels "go-deploy/models/sys/user"
	"go-deploy/pkg/config"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/deployment_service"
	"go-deploy/service/sm_service"
	"go-deploy/service/user_service"
	"go-deploy/service/vm_service"
	"go-deploy/utils"
)

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
		users, err := user_service.New().WithAuth(auth).Discover(user_service.DiscoverUsersOpts{
			Search:     requestQuery.Search,
			Pagination: &service.Pagination{Page: requestQuery.Page, PageSize: requestQuery.PageSize},
		})
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

	usc := user_service.New().WithAuth(auth)

	users, err := usc.List(user_service.ListUsersOpts{
		Pagination: &service.Pagination{Page: requestQuery.Page, PageSize: requestQuery.PageSize},
		Search:     requestQuery.Search,
		All:        requestQuery.All,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	usersDto := make([]body.UserRead, 0)
	for _, user := range users {
		// if we list ourselves, take the opportunity to update our role
		if user.ID == auth.UserID {
			updatedUser, err := usc.Create()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get or create a user when listing: %w", err))
				continue
			}

			if updatedUser != nil {
				user = *updatedUser
			}
		}

		storageURL, err := getSmURL(user.ID, auth)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get storage url for a user when listing: %w", err))
			continue
		}

		role := config.Config.GetRole(user.EffectiveRole.Name)

		usage, _ := collectUsage(user.ID)
		if usage == nil {
			usage = &userModels.Usage{}
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

	var effectiveRole *roleModels.Role
	var user *userModels.User

	usc := user_service.New().WithAuth(auth)

	if requestURI.UserID == auth.UserID {
		effectiveRole = auth.GetEffectiveRole()
		user, err = usc.Create()
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if user == nil {
			context.NotFound("User not found")
			return
		}
	} else {
		user, err = usc.Get(requestURI.UserID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if user == nil {
			context.NotFound("User not found")
			return
		}

		effectiveRole = config.Config.GetRole(user.EffectiveRole.Name)
		if effectiveRole == nil {
			effectiveRole = &roleModels.Role{}
		}
	}

	usage, err := collectUsage(user.ID)
	if usage == nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	storageURL, err := getSmURL(user.ID, auth)
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

	var effectiveRole *roleModels.Role

	usc := user_service.New().WithAuth(auth)

	if requestURI.UserID == auth.UserID {
		effectiveRole = auth.GetEffectiveRole()
		_, err = usc.Create()
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	}

	updated, err := usc.Update(requestURI.UserID, &userUpdate)
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

	storageURL, err := getSmURL(updated.ID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.JSONResponse(200, updated.ToDTO(effectiveRole, usage, storageURL))
}

// collectUsage is helper function to collect usage for a user.
// This includes how many deployments, cpu cores, ram and disk size etc. the user has.
func collectUsage(userID string) (*userModels.Usage, error) {
	vmUsage, err := vm_service.New().GetUsage(userID)
	if err != nil {
		return nil, err
	}

	deploymentUsage, err := deployment_service.New().GetUsage(userID)
	if err != nil {
		return nil, err
	}

	usage := &userModels.Usage{
		Deployments: deploymentUsage.Count,
		CpuCores:    vmUsage.CpuCores,
		RAM:         vmUsage.RAM,
		DiskSize:    vmUsage.DiskSize,
	}

	return usage, nil
}

// getSmURL is helper function to get storage manager url for a user.
func getSmURL(userID string, auth *service.AuthInfo) (*string, error) {
	sm, err := sm_service.New().WithAuth(auth).GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if sm == nil {
		return nil, nil
	}

	return sm.GetURL(), nil
}
