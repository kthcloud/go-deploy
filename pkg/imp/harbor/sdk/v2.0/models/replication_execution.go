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

// ReplicationExecution The replication execution
//
// swagger:model ReplicationExecution
type ReplicationExecution struct {

	// The end time
	// Format: date-time
	EndTime strfmt.DateTime `json:"end_time,omitempty"`

	// The count of failed executions
	Failed int64 `json:"failed"`

	// The ID of the execution
	ID int64 `json:"id,omitempty"`

	// The count of in_progress executions
	InProgress int64 `json:"in_progress"`

	// The ID if the policy that the execution belongs to
	PolicyID int64 `json:"policy_id,omitempty"`

	// The start time
	// Format: date-time
	StartTime strfmt.DateTime `json:"start_time,omitempty"`

	// The status of the execution
	Status string `json:"status,omitempty"`

	// The status text
	StatusText string `json:"status_text"`

	// The count of stopped executions
	Stopped int64 `json:"stopped"`

	// The count of succeed executions
	Succeed int64 `json:"succeed"`

	// The total count of all executions
	Total int64 `json:"total"`

	// The trigger mode
	Trigger string `json:"trigger,omitempty"`
}

// Validate validates this replication execution
func (m *ReplicationExecution) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateEndTime(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStartTime(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ReplicationExecution) validateEndTime(formats strfmt.Registry) error {
	if swag.IsZero(m.EndTime) { // not required
		return nil
	}

	if err := validate.FormatOf("end_time", "body", "date-time", m.EndTime.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *ReplicationExecution) validateStartTime(formats strfmt.Registry) error {
	if swag.IsZero(m.StartTime) { // not required
		return nil
	}

	if err := validate.FormatOf("start_time", "body", "date-time", m.StartTime.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this replication execution based on context it is used
func (m *ReplicationExecution) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ReplicationExecution) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ReplicationExecution) UnmarshalBinary(b []byte) error {
	var res ReplicationExecution
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
