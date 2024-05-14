// Code generated by go-swagger; DO NOT EDIT.

package user

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// SetUserSysAdminReader is a Reader for the SetUserSysAdmin structure.
type SetUserSysAdminReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *SetUserSysAdminReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewSetUserSysAdminOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewSetUserSysAdminUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewSetUserSysAdminForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewSetUserSysAdminNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewSetUserSysAdminInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewSetUserSysAdminOK creates a SetUserSysAdminOK with default headers values
func NewSetUserSysAdminOK() *SetUserSysAdminOK {
	return &SetUserSysAdminOK{}
}

/*
SetUserSysAdminOK describes a response with status code 200, with default header values.

Success
*/
type SetUserSysAdminOK struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string
}

// IsSuccess returns true when this set user sys admin o k response has a 2xx status code
func (o *SetUserSysAdminOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this set user sys admin o k response has a 3xx status code
func (o *SetUserSysAdminOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this set user sys admin o k response has a 4xx status code
func (o *SetUserSysAdminOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this set user sys admin o k response has a 5xx status code
func (o *SetUserSysAdminOK) IsServerError() bool {
	return false
}

// IsCode returns true when this set user sys admin o k response a status code equal to that given
func (o *SetUserSysAdminOK) IsCode(code int) bool {
	return code == 200
}

func (o *SetUserSysAdminOK) Error() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminOK ", 200)
}

func (o *SetUserSysAdminOK) String() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminOK ", 200)
}

func (o *SetUserSysAdminOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header X-Request-Id
	hdrXRequestID := response.GetHeader("X-Request-Id")

	if hdrXRequestID != "" {
		o.XRequestID = hdrXRequestID
	}

	return nil
}

// NewSetUserSysAdminUnauthorized creates a SetUserSysAdminUnauthorized with default headers values
func NewSetUserSysAdminUnauthorized() *SetUserSysAdminUnauthorized {
	return &SetUserSysAdminUnauthorized{}
}

/*
SetUserSysAdminUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type SetUserSysAdminUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this set user sys admin unauthorized response has a 2xx status code
func (o *SetUserSysAdminUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this set user sys admin unauthorized response has a 3xx status code
func (o *SetUserSysAdminUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this set user sys admin unauthorized response has a 4xx status code
func (o *SetUserSysAdminUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this set user sys admin unauthorized response has a 5xx status code
func (o *SetUserSysAdminUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this set user sys admin unauthorized response a status code equal to that given
func (o *SetUserSysAdminUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *SetUserSysAdminUnauthorized) Error() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminUnauthorized  %+v", 401, o.Payload)
}

func (o *SetUserSysAdminUnauthorized) String() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminUnauthorized  %+v", 401, o.Payload)
}

func (o *SetUserSysAdminUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *SetUserSysAdminUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header X-Request-Id
	hdrXRequestID := response.GetHeader("X-Request-Id")

	if hdrXRequestID != "" {
		o.XRequestID = hdrXRequestID
	}

	o.Payload = new(models.Errors)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSetUserSysAdminForbidden creates a SetUserSysAdminForbidden with default headers values
func NewSetUserSysAdminForbidden() *SetUserSysAdminForbidden {
	return &SetUserSysAdminForbidden{}
}

/*
SetUserSysAdminForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type SetUserSysAdminForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this set user sys admin forbidden response has a 2xx status code
func (o *SetUserSysAdminForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this set user sys admin forbidden response has a 3xx status code
func (o *SetUserSysAdminForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this set user sys admin forbidden response has a 4xx status code
func (o *SetUserSysAdminForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this set user sys admin forbidden response has a 5xx status code
func (o *SetUserSysAdminForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this set user sys admin forbidden response a status code equal to that given
func (o *SetUserSysAdminForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *SetUserSysAdminForbidden) Error() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminForbidden  %+v", 403, o.Payload)
}

func (o *SetUserSysAdminForbidden) String() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminForbidden  %+v", 403, o.Payload)
}

func (o *SetUserSysAdminForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *SetUserSysAdminForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header X-Request-Id
	hdrXRequestID := response.GetHeader("X-Request-Id")

	if hdrXRequestID != "" {
		o.XRequestID = hdrXRequestID
	}

	o.Payload = new(models.Errors)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSetUserSysAdminNotFound creates a SetUserSysAdminNotFound with default headers values
func NewSetUserSysAdminNotFound() *SetUserSysAdminNotFound {
	return &SetUserSysAdminNotFound{}
}

/*
SetUserSysAdminNotFound describes a response with status code 404, with default header values.

Not found
*/
type SetUserSysAdminNotFound struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this set user sys admin not found response has a 2xx status code
func (o *SetUserSysAdminNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this set user sys admin not found response has a 3xx status code
func (o *SetUserSysAdminNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this set user sys admin not found response has a 4xx status code
func (o *SetUserSysAdminNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this set user sys admin not found response has a 5xx status code
func (o *SetUserSysAdminNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this set user sys admin not found response a status code equal to that given
func (o *SetUserSysAdminNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *SetUserSysAdminNotFound) Error() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminNotFound  %+v", 404, o.Payload)
}

func (o *SetUserSysAdminNotFound) String() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminNotFound  %+v", 404, o.Payload)
}

func (o *SetUserSysAdminNotFound) GetPayload() *models.Errors {
	return o.Payload
}

func (o *SetUserSysAdminNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header X-Request-Id
	hdrXRequestID := response.GetHeader("X-Request-Id")

	if hdrXRequestID != "" {
		o.XRequestID = hdrXRequestID
	}

	o.Payload = new(models.Errors)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSetUserSysAdminInternalServerError creates a SetUserSysAdminInternalServerError with default headers values
func NewSetUserSysAdminInternalServerError() *SetUserSysAdminInternalServerError {
	return &SetUserSysAdminInternalServerError{}
}

/*
SetUserSysAdminInternalServerError describes a response with status code 500, with default header values.

Unexpected internal errors.
*/
type SetUserSysAdminInternalServerError struct {
}

// IsSuccess returns true when this set user sys admin internal server error response has a 2xx status code
func (o *SetUserSysAdminInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this set user sys admin internal server error response has a 3xx status code
func (o *SetUserSysAdminInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this set user sys admin internal server error response has a 4xx status code
func (o *SetUserSysAdminInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this set user sys admin internal server error response has a 5xx status code
func (o *SetUserSysAdminInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this set user sys admin internal server error response a status code equal to that given
func (o *SetUserSysAdminInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *SetUserSysAdminInternalServerError) Error() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminInternalServerError ", 500)
}

func (o *SetUserSysAdminInternalServerError) String() string {
	return fmt.Sprintf("[PUT /users/{user_id}/sysadmin][%d] setUserSysAdminInternalServerError ", 500)
}

func (o *SetUserSysAdminInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
