// Code generated by go-swagger; DO NOT EDIT.

package preheat

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// UpdatePolicyReader is a Reader for the UpdatePolicy structure.
type UpdatePolicyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UpdatePolicyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUpdatePolicyOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewUpdatePolicyBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewUpdatePolicyUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewUpdatePolicyForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewUpdatePolicyNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 409:
		result := NewUpdatePolicyConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewUpdatePolicyInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewUpdatePolicyOK creates a UpdatePolicyOK with default headers values
func NewUpdatePolicyOK() *UpdatePolicyOK {
	return &UpdatePolicyOK{}
}

/*
UpdatePolicyOK describes a response with status code 200, with default header values.

Success
*/
type UpdatePolicyOK struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string
}

// IsSuccess returns true when this update policy o k response has a 2xx status code
func (o *UpdatePolicyOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this update policy o k response has a 3xx status code
func (o *UpdatePolicyOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update policy o k response has a 4xx status code
func (o *UpdatePolicyOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this update policy o k response has a 5xx status code
func (o *UpdatePolicyOK) IsServerError() bool {
	return false
}

// IsCode returns true when this update policy o k response a status code equal to that given
func (o *UpdatePolicyOK) IsCode(code int) bool {
	return code == 200
}

func (o *UpdatePolicyOK) Error() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyOK ", 200)
}

func (o *UpdatePolicyOK) String() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyOK ", 200)
}

func (o *UpdatePolicyOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header X-Request-Id
	hdrXRequestID := response.GetHeader("X-Request-Id")

	if hdrXRequestID != "" {
		o.XRequestID = hdrXRequestID
	}

	return nil
}

// NewUpdatePolicyBadRequest creates a UpdatePolicyBadRequest with default headers values
func NewUpdatePolicyBadRequest() *UpdatePolicyBadRequest {
	return &UpdatePolicyBadRequest{}
}

/*
UpdatePolicyBadRequest describes a response with status code 400, with default header values.

Bad request
*/
type UpdatePolicyBadRequest struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update policy bad request response has a 2xx status code
func (o *UpdatePolicyBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update policy bad request response has a 3xx status code
func (o *UpdatePolicyBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update policy bad request response has a 4xx status code
func (o *UpdatePolicyBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this update policy bad request response has a 5xx status code
func (o *UpdatePolicyBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this update policy bad request response a status code equal to that given
func (o *UpdatePolicyBadRequest) IsCode(code int) bool {
	return code == 400
}

func (o *UpdatePolicyBadRequest) Error() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyBadRequest  %+v", 400, o.Payload)
}

func (o *UpdatePolicyBadRequest) String() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyBadRequest  %+v", 400, o.Payload)
}

func (o *UpdatePolicyBadRequest) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdatePolicyBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewUpdatePolicyUnauthorized creates a UpdatePolicyUnauthorized with default headers values
func NewUpdatePolicyUnauthorized() *UpdatePolicyUnauthorized {
	return &UpdatePolicyUnauthorized{}
}

/*
UpdatePolicyUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type UpdatePolicyUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update policy unauthorized response has a 2xx status code
func (o *UpdatePolicyUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update policy unauthorized response has a 3xx status code
func (o *UpdatePolicyUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update policy unauthorized response has a 4xx status code
func (o *UpdatePolicyUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this update policy unauthorized response has a 5xx status code
func (o *UpdatePolicyUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this update policy unauthorized response a status code equal to that given
func (o *UpdatePolicyUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *UpdatePolicyUnauthorized) Error() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyUnauthorized  %+v", 401, o.Payload)
}

func (o *UpdatePolicyUnauthorized) String() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyUnauthorized  %+v", 401, o.Payload)
}

func (o *UpdatePolicyUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdatePolicyUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewUpdatePolicyForbidden creates a UpdatePolicyForbidden with default headers values
func NewUpdatePolicyForbidden() *UpdatePolicyForbidden {
	return &UpdatePolicyForbidden{}
}

/*
UpdatePolicyForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type UpdatePolicyForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update policy forbidden response has a 2xx status code
func (o *UpdatePolicyForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update policy forbidden response has a 3xx status code
func (o *UpdatePolicyForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update policy forbidden response has a 4xx status code
func (o *UpdatePolicyForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this update policy forbidden response has a 5xx status code
func (o *UpdatePolicyForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this update policy forbidden response a status code equal to that given
func (o *UpdatePolicyForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *UpdatePolicyForbidden) Error() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyForbidden  %+v", 403, o.Payload)
}

func (o *UpdatePolicyForbidden) String() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyForbidden  %+v", 403, o.Payload)
}

func (o *UpdatePolicyForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdatePolicyForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewUpdatePolicyNotFound creates a UpdatePolicyNotFound with default headers values
func NewUpdatePolicyNotFound() *UpdatePolicyNotFound {
	return &UpdatePolicyNotFound{}
}

/*
UpdatePolicyNotFound describes a response with status code 404, with default header values.

Not found
*/
type UpdatePolicyNotFound struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update policy not found response has a 2xx status code
func (o *UpdatePolicyNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update policy not found response has a 3xx status code
func (o *UpdatePolicyNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update policy not found response has a 4xx status code
func (o *UpdatePolicyNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this update policy not found response has a 5xx status code
func (o *UpdatePolicyNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this update policy not found response a status code equal to that given
func (o *UpdatePolicyNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *UpdatePolicyNotFound) Error() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyNotFound  %+v", 404, o.Payload)
}

func (o *UpdatePolicyNotFound) String() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyNotFound  %+v", 404, o.Payload)
}

func (o *UpdatePolicyNotFound) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdatePolicyNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewUpdatePolicyConflict creates a UpdatePolicyConflict with default headers values
func NewUpdatePolicyConflict() *UpdatePolicyConflict {
	return &UpdatePolicyConflict{}
}

/*
UpdatePolicyConflict describes a response with status code 409, with default header values.

Conflict
*/
type UpdatePolicyConflict struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update policy conflict response has a 2xx status code
func (o *UpdatePolicyConflict) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update policy conflict response has a 3xx status code
func (o *UpdatePolicyConflict) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update policy conflict response has a 4xx status code
func (o *UpdatePolicyConflict) IsClientError() bool {
	return true
}

// IsServerError returns true when this update policy conflict response has a 5xx status code
func (o *UpdatePolicyConflict) IsServerError() bool {
	return false
}

// IsCode returns true when this update policy conflict response a status code equal to that given
func (o *UpdatePolicyConflict) IsCode(code int) bool {
	return code == 409
}

func (o *UpdatePolicyConflict) Error() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyConflict  %+v", 409, o.Payload)
}

func (o *UpdatePolicyConflict) String() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyConflict  %+v", 409, o.Payload)
}

func (o *UpdatePolicyConflict) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdatePolicyConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewUpdatePolicyInternalServerError creates a UpdatePolicyInternalServerError with default headers values
func NewUpdatePolicyInternalServerError() *UpdatePolicyInternalServerError {
	return &UpdatePolicyInternalServerError{}
}

/*
UpdatePolicyInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type UpdatePolicyInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update policy internal server error response has a 2xx status code
func (o *UpdatePolicyInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update policy internal server error response has a 3xx status code
func (o *UpdatePolicyInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update policy internal server error response has a 4xx status code
func (o *UpdatePolicyInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this update policy internal server error response has a 5xx status code
func (o *UpdatePolicyInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this update policy internal server error response a status code equal to that given
func (o *UpdatePolicyInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *UpdatePolicyInternalServerError) Error() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *UpdatePolicyInternalServerError) String() string {
	return fmt.Sprintf("[PUT /projects/{project_name}/preheat/policies/{preheat_policy_name}][%d] updatePolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *UpdatePolicyInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdatePolicyInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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
