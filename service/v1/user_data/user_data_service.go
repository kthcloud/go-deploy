package user_data

import (
	userDataModels "go-deploy/models/sys/user_data"
	"go-deploy/service/errors"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v1/user_data/opts"
)

// Get gets a user data
//
// It uses service.AuthInfo to only return the resource the requesting user has access to
func (c *Client) Get(id string, opts ...opts.GetOpts) (*userDataModels.UserData, error) {
	_ = sUtils.GetFirstOrDefault(opts)

	umc := userDataModels.New()

	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		umc.WithUserID(c.V1.Auth().UserID)
	}

	return umc.GetByID(id)
}

// List lists user data
//
// It uses service.AuthInfo to only return the resources the requesting user has access to
func (c *Client) List(opts ...opts.ListOpts) ([]userDataModels.UserData, error) {
	o := sUtils.GetFirstOrDefault(opts)

	umc := userDataModels.New()

	if o.Pagination != nil {
		umc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's user data is requested
		if !c.V1.HasAuth() || c.V1.Auth().UserID == *o.UserID || c.V1.Auth().IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.V1.Auth().UserID
		}
	} else {
		// All userdata is requested
		if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
			effectiveUserID = c.V1.Auth().UserID
		}
	}

	if effectiveUserID != "" {
		umc.WithUserID(effectiveUserID)
	}

	return umc.List()
}

// Create creates a user data
//
// It uses service.AuthInfo to only create the resource the requesting user has access to
func (c *Client) Create(id, data, userID string) (*userDataModels.UserData, error) {
	if c.V1.Auth() != nil && userID != c.V1.Auth().UserID && !c.V1.Auth().IsAdmin {
		return nil, nil
	}

	udc := userDataModels.New().WithUserID(userID)

	// Ensure max 10 user data per user
	userDataCount, err := udc.Count()
	if err != nil {
		return nil, err
	}

	if userDataCount >= 10 {
		return nil, errors.NewQuotaExceededError("User Data quota exceeded. Max 10 user data per user allowed.")
	}

	return udc.Create(id, data, userID)
}

// Update updates a user data
//
// It uses service.AuthInfo to only update the resource the requesting user has access to
func (c *Client) Update(id, data string) (*userDataModels.UserData, error) {
	umc := userDataModels.New()

	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		umc.WithUserID(c.V1.Auth().UserID)
	}

	return umc.Update(id, data)
}

// Delete deletes a user data
//
// It uses service.AuthInfo to only delete the resource the requesting user has access to
func (c *Client) Delete(id string) error {
	umc := userDataModels.New()

	// User data ID are not unique, so not specifying the user ID here,
	// and multiple user data share the same ID, will result in
	// delete one random user data with the given ID
	if c.V1.Auth() != nil {
		umc.WithUserID(c.V1.Auth().UserID)
	}

	exists, err := umc.ExistsByID(id)
	if err != nil {
		return err
	}

	if !exists {
		return errors.UserDataNotFoundErr
	}

	return umc.EraseByID(id)
}

// Exists checks if user data exists
//
// It uses service.AuthInfo to only check if the resource the requesting user has access to exists
func (c *Client) Exists(id string) (bool, error) {
	umc := userDataModels.New()

	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		umc.WithUserID(c.V1.Auth().UserID)
	}

	return umc.ExistsByID(id)
}
