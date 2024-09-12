// Code generated by go-swagger; DO NOT EDIT.

package gc

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// UpdateGCScheduleReader is a Reader for the UpdateGCSchedule structure.
type UpdateGCScheduleReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UpdateGCScheduleReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUpdateGCScheduleOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewUpdateGCScheduleBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewUpdateGCScheduleUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewUpdateGCScheduleForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewUpdateGCScheduleInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewUpdateGCScheduleOK creates a UpdateGCScheduleOK with default headers values
func NewUpdateGCScheduleOK() *UpdateGCScheduleOK {
	return &UpdateGCScheduleOK{}
}

/*
UpdateGCScheduleOK describes a response with status code 200, with default header values.

Updated gc's schedule successfully.
*/
type UpdateGCScheduleOK struct {
}

// IsSuccess returns true when this update Gc schedule o k response has a 2xx status code
func (o *UpdateGCScheduleOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this update Gc schedule o k response has a 3xx status code
func (o *UpdateGCScheduleOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update Gc schedule o k response has a 4xx status code
func (o *UpdateGCScheduleOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this update Gc schedule o k response has a 5xx status code
func (o *UpdateGCScheduleOK) IsServerError() bool {
	return false
}

// IsCode returns true when this update Gc schedule o k response a status code equal to that given
func (o *UpdateGCScheduleOK) IsCode(code int) bool {
	return code == 200
}

func (o *UpdateGCScheduleOK) Error() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleOK ", 200)
}

func (o *UpdateGCScheduleOK) String() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleOK ", 200)
}

func (o *UpdateGCScheduleOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewUpdateGCScheduleBadRequest creates a UpdateGCScheduleBadRequest with default headers values
func NewUpdateGCScheduleBadRequest() *UpdateGCScheduleBadRequest {
	return &UpdateGCScheduleBadRequest{}
}

/*
UpdateGCScheduleBadRequest describes a response with status code 400, with default header values.

Bad request
*/
type UpdateGCScheduleBadRequest struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update Gc schedule bad request response has a 2xx status code
func (o *UpdateGCScheduleBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update Gc schedule bad request response has a 3xx status code
func (o *UpdateGCScheduleBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update Gc schedule bad request response has a 4xx status code
func (o *UpdateGCScheduleBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this update Gc schedule bad request response has a 5xx status code
func (o *UpdateGCScheduleBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this update Gc schedule bad request response a status code equal to that given
func (o *UpdateGCScheduleBadRequest) IsCode(code int) bool {
	return code == 400
}

func (o *UpdateGCScheduleBadRequest) Error() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleBadRequest  %+v", 400, o.Payload)
}

func (o *UpdateGCScheduleBadRequest) String() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleBadRequest  %+v", 400, o.Payload)
}

func (o *UpdateGCScheduleBadRequest) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdateGCScheduleBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewUpdateGCScheduleUnauthorized creates a UpdateGCScheduleUnauthorized with default headers values
func NewUpdateGCScheduleUnauthorized() *UpdateGCScheduleUnauthorized {
	return &UpdateGCScheduleUnauthorized{}
}

/*
UpdateGCScheduleUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type UpdateGCScheduleUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update Gc schedule unauthorized response has a 2xx status code
func (o *UpdateGCScheduleUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update Gc schedule unauthorized response has a 3xx status code
func (o *UpdateGCScheduleUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update Gc schedule unauthorized response has a 4xx status code
func (o *UpdateGCScheduleUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this update Gc schedule unauthorized response has a 5xx status code
func (o *UpdateGCScheduleUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this update Gc schedule unauthorized response a status code equal to that given
func (o *UpdateGCScheduleUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *UpdateGCScheduleUnauthorized) Error() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleUnauthorized  %+v", 401, o.Payload)
}

func (o *UpdateGCScheduleUnauthorized) String() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleUnauthorized  %+v", 401, o.Payload)
}

func (o *UpdateGCScheduleUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdateGCScheduleUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewUpdateGCScheduleForbidden creates a UpdateGCScheduleForbidden with default headers values
func NewUpdateGCScheduleForbidden() *UpdateGCScheduleForbidden {
	return &UpdateGCScheduleForbidden{}
}

/*
UpdateGCScheduleForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type UpdateGCScheduleForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update Gc schedule forbidden response has a 2xx status code
func (o *UpdateGCScheduleForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update Gc schedule forbidden response has a 3xx status code
func (o *UpdateGCScheduleForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update Gc schedule forbidden response has a 4xx status code
func (o *UpdateGCScheduleForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this update Gc schedule forbidden response has a 5xx status code
func (o *UpdateGCScheduleForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this update Gc schedule forbidden response a status code equal to that given
func (o *UpdateGCScheduleForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *UpdateGCScheduleForbidden) Error() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleForbidden  %+v", 403, o.Payload)
}

func (o *UpdateGCScheduleForbidden) String() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleForbidden  %+v", 403, o.Payload)
}

func (o *UpdateGCScheduleForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdateGCScheduleForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewUpdateGCScheduleInternalServerError creates a UpdateGCScheduleInternalServerError with default headers values
func NewUpdateGCScheduleInternalServerError() *UpdateGCScheduleInternalServerError {
	return &UpdateGCScheduleInternalServerError{}
}

/*
UpdateGCScheduleInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type UpdateGCScheduleInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this update Gc schedule internal server error response has a 2xx status code
func (o *UpdateGCScheduleInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update Gc schedule internal server error response has a 3xx status code
func (o *UpdateGCScheduleInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update Gc schedule internal server error response has a 4xx status code
func (o *UpdateGCScheduleInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this update Gc schedule internal server error response has a 5xx status code
func (o *UpdateGCScheduleInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this update Gc schedule internal server error response a status code equal to that given
func (o *UpdateGCScheduleInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *UpdateGCScheduleInternalServerError) Error() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleInternalServerError  %+v", 500, o.Payload)
}

func (o *UpdateGCScheduleInternalServerError) String() string {
	return fmt.Sprintf("[PUT /system/gc/schedule][%d] updateGcScheduleInternalServerError  %+v", 500, o.Payload)
}

func (o *UpdateGCScheduleInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *UpdateGCScheduleInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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
