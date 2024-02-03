package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/v1/body"
	"go-deploy/models/dto/v1/query"
	"go-deploy/models/dto/v1/uri"
	roleModels "go-deploy/models/sys/role"
	userModels "go-deploy/models/sys/user"
	"go-deploy/pkg/config"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/clients"
	"go-deploy/service/v1/users/opts"
	sUtils "go-deploy/service/v1/utils"
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
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	if requestQuery.Discover {
		userList, err := deployV1.Users().Discover(opts.DiscoverOpts{
			Search:     requestQuery.Search,
			Pagination: sUtils.GetOrDefaultPagination(requestQuery.Pagination),
		})
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if userList == nil {
			context.Ok([]interface{}{})
			return
		}

		context.Ok(userList)
		return
	}

	userList, err := deployV1.Users().List(opts.ListOpts{
		Pagination: sUtils.GetOrDefaultPagination(requestQuery.Pagination),
		Search:     requestQuery.Search,
		All:        requestQuery.All,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	usersDto := make([]body.UserRead, 0)
	for _, user := range userList {
		// if we list ourselves, take the opportunity to update our role
		if user.ID == auth.UserID {
			updatedUser, err := deployV1.Users().Create()
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to get or create a user when listing: %w", err))
				continue
			}

			if updatedUser != nil {
				user = *updatedUser
			}
		}

		role := config.Config.GetRole(user.EffectiveRole.Name)
		usage, _ := collectUsage(deployV1, user.ID)
		if usage == nil {
			usage = &userModels.Usage{}
		}

		usersDto = append(usersDto, user.ToDTO(role, usage, deployV1.SMs().GetURL(user.ID)))
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

	deployV1 := service.V1(auth)

	if requestURI.UserID == auth.UserID {
		effectiveRole = auth.GetEffectiveRole()
		user, err = deployV1.Users().Create()
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if user == nil {
			context.NotFound("User not found")
			return
		}
	} else {
		user, err = deployV1.Users().Get(requestURI.UserID)
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

	usage, err := collectUsage(deployV1, user.ID)
	if usage == nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.JSONResponse(200, user.ToDTO(effectiveRole, usage, deployV1.SMs().GetURL(user.ID)))
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

	deployV1 := service.V1(auth)

	if requestURI.UserID == auth.UserID {
		effectiveRole = auth.GetEffectiveRole()
		_, err = deployV1.Users().Create()
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	}

	updated, err := deployV1.Users().Update(requestURI.UserID, &userUpdate)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if updated == nil {
		context.NotFound("User not found")
		return
	}

	usage, err := collectUsage(deployV1, updated.ID)
	if usage == nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.JSONResponse(200, updated.ToDTO(effectiveRole, usage, deployV1.SMs().GetURL(updated.ID)))
}

// collectUsage is helper function to collect usage for a user.
// This includes how many deployments, cpu cores, ram and disk size etc. the user has.
func collectUsage(deployV1 clients.V1, userID string) (*userModels.Usage, error) {
	vmUsage, err := deployV1.VMs().GetUsage(userID)
	if err != nil {
		return nil, err
	}

	deploymentUsage, err := deployV1.Deployments().GetUsage(userID)
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
