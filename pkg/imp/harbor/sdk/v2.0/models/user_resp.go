// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// UserResp user resp
//
// swagger:model UserResp
type UserResp struct {

	// indicate the admin privilege is grant by authenticator (LDAP), is always false unless it is the current login user
	AdminRoleInAuth bool `json:"admin_role_in_auth"`

	// comment
	Comment string `json:"comment,omitempty"`

	// The creation time of the user.
	// Format: date-time
	CreationTime strfmt.DateTime `json:"creation_time,omitempty"`

	// email
	Email string `json:"email,omitempty"`

	// oidc user meta
	OIDCUserMeta *OIDCUserInfo `json:"oidc_user_meta,omitempty"`

	// realname
	Realname string `json:"realname,omitempty"`

	// sysadmin flag
	SysadminFlag bool `json:"sysadmin_flag"`

	// The update time of the user.
	// Format: date-time
	UpdateTime strfmt.DateTime `json:"update_time,omitempty"`

	// user id
	UserID int64 `json:"user_id,omitempty"`

	// username
	Username string `json:"username,omitempty"`
}

// Validate validates this user resp
func (m *UserResp) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCreationTime(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOIDCUserMeta(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUpdateTime(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *UserResp) validateCreationTime(formats strfmt.Registry) error {
	if swag.IsZero(m.CreationTime) { // not required
		return nil
	}

	if err := validate.FormatOf("creation_time", "body", "date-time", m.CreationTime.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *UserResp) validateOIDCUserMeta(formats strfmt.Registry) error {
	if swag.IsZero(m.OIDCUserMeta) { // not required
		return nil
	}

	if m.OIDCUserMeta != nil {
		if err := m.OIDCUserMeta.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("oidc_user_meta")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("oidc_user_meta")
			}
			return err
		}
	}

	return nil
}

func (m *UserResp) validateUpdateTime(formats strfmt.Registry) error {
	if swag.IsZero(m.UpdateTime) { // not required
		return nil
	}

	if err := validate.FormatOf("update_time", "body", "date-time", m.UpdateTime.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this user resp based on the context it is used
func (m *UserResp) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateOIDCUserMeta(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *UserResp) contextValidateOIDCUserMeta(ctx context.Context, formats strfmt.Registry) error {

	if m.OIDCUserMeta != nil {
		if err := m.OIDCUserMeta.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("oidc_user_meta")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("oidc_user_meta")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *UserResp) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UserResp) UnmarshalBinary(b []byte) error {
	var res UserResp
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
