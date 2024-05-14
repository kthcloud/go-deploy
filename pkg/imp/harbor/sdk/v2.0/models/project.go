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

// Project project
//
// swagger:model Project
type Project struct {

	// The creation time of the project.
	// Format: date-time
	CreationTime strfmt.DateTime `json:"creation_time,omitempty"`

	// The role ID with highest permission of the current user who triggered the API (for UI).  This attribute is deprecated and will be removed in future versions.
	CurrentUserRoleID int64 `json:"current_user_role_id,omitempty"`

	// The list of role ID of the current user who triggered the API (for UI)
	CurrentUserRoleIds []int32 `json:"current_user_role_ids"`

	// The CVE allowlist of this project.
	CVEAllowlist *CVEAllowlist `json:"cve_allowlist,omitempty"`

	// A deletion mark of the project.
	Deleted bool `json:"deleted,omitempty"`

	// The metadata of the project.
	Metadata *ProjectMetadata `json:"metadata,omitempty"`

	// The name of the project.
	Name string `json:"name,omitempty"`

	// The owner ID of the project always means the creator of the project.
	OwnerID int32 `json:"owner_id,omitempty"`

	// The owner name of the project.
	OwnerName string `json:"owner_name,omitempty"`

	// Project ID
	ProjectID int32 `json:"project_id,omitempty"`

	// The ID of referenced registry when the project is a proxy cache project.
	RegistryID int64 `json:"registry_id,omitempty"`

	// The number of the repositories under this project.
	RepoCount int64 `json:"repo_count"`

	// Correspond to the UI about whether the project's publicity is  updatable (for UI)
	Togglable bool `json:"togglable,omitempty"`

	// The update time of the project.
	// Format: date-time
	UpdateTime strfmt.DateTime `json:"update_time,omitempty"`
}

// Validate validates this project
func (m *Project) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCreationTime(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCVEAllowlist(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMetadata(formats); err != nil {
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

func (m *Project) validateCreationTime(formats strfmt.Registry) error {
	if swag.IsZero(m.CreationTime) { // not required
		return nil
	}

	if err := validate.FormatOf("creation_time", "body", "date-time", m.CreationTime.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Project) validateCVEAllowlist(formats strfmt.Registry) error {
	if swag.IsZero(m.CVEAllowlist) { // not required
		return nil
	}

	if m.CVEAllowlist != nil {
		if err := m.CVEAllowlist.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cve_allowlist")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("cve_allowlist")
			}
			return err
		}
	}

	return nil
}

func (m *Project) validateMetadata(formats strfmt.Registry) error {
	if swag.IsZero(m.Metadata) { // not required
		return nil
	}

	if m.Metadata != nil {
		if err := m.Metadata.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("metadata")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("metadata")
			}
			return err
		}
	}

	return nil
}

func (m *Project) validateUpdateTime(formats strfmt.Registry) error {
	if swag.IsZero(m.UpdateTime) { // not required
		return nil
	}

	if err := validate.FormatOf("update_time", "body", "date-time", m.UpdateTime.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this project based on the context it is used
func (m *Project) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateCVEAllowlist(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateMetadata(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Project) contextValidateCVEAllowlist(ctx context.Context, formats strfmt.Registry) error {

	if m.CVEAllowlist != nil {
		if err := m.CVEAllowlist.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cve_allowlist")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("cve_allowlist")
			}
			return err
		}
	}

	return nil
}

func (m *Project) contextValidateMetadata(ctx context.Context, formats strfmt.Registry) error {

	if m.Metadata != nil {
		if err := m.Metadata.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("metadata")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("metadata")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Project) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Project) UnmarshalBinary(b []byte) error {
	var res Project
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
