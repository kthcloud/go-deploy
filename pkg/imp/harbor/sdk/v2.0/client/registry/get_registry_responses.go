// Code generated by go-swagger; DO NOT EDIT.

package registry

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// GetRegistryReader is a Reader for the GetRegistry structure.
type GetRegistryReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetRegistryReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetRegistryOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetRegistryUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewGetRegistryForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetRegistryNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetRegistryInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetRegistryOK creates a GetRegistryOK with default headers values
func NewGetRegistryOK() *GetRegistryOK {
	return &GetRegistryOK{}
}

/*
GetRegistryOK describes a response with status code 200, with default header values.

Success
*/
type GetRegistryOK struct {
	Payload *models.Registry
}

// IsSuccess returns true when this get registry o k response has a 2xx status code
func (o *GetRegistryOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get registry o k response has a 3xx status code
func (o *GetRegistryOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get registry o k response has a 4xx status code
func (o *GetRegistryOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get registry o k response has a 5xx status code
func (o *GetRegistryOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get registry o k response a status code equal to that given
func (o *GetRegistryOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetRegistryOK) Error() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryOK  %+v", 200, o.Payload)
}

func (o *GetRegistryOK) String() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryOK  %+v", 200, o.Payload)
}

func (o *GetRegistryOK) GetPayload() *models.Registry {
	return o.Payload
}

func (o *GetRegistryOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Registry)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetRegistryUnauthorized creates a GetRegistryUnauthorized with default headers values
func NewGetRegistryUnauthorized() *GetRegistryUnauthorized {
	return &GetRegistryUnauthorized{}
}

/*
GetRegistryUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetRegistryUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this get registry unauthorized response has a 2xx status code
func (o *GetRegistryUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get registry unauthorized response has a 3xx status code
func (o *GetRegistryUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get registry unauthorized response has a 4xx status code
func (o *GetRegistryUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get registry unauthorized response has a 5xx status code
func (o *GetRegistryUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get registry unauthorized response a status code equal to that given
func (o *GetRegistryUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetRegistryUnauthorized) Error() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryUnauthorized  %+v", 401, o.Payload)
}

func (o *GetRegistryUnauthorized) String() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryUnauthorized  %+v", 401, o.Payload)
}

func (o *GetRegistryUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *GetRegistryUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewGetRegistryForbidden creates a GetRegistryForbidden with default headers values
func NewGetRegistryForbidden() *GetRegistryForbidden {
	return &GetRegistryForbidden{}
}

/*
GetRegistryForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type GetRegistryForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this get registry forbidden response has a 2xx status code
func (o *GetRegistryForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get registry forbidden response has a 3xx status code
func (o *GetRegistryForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get registry forbidden response has a 4xx status code
func (o *GetRegistryForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this get registry forbidden response has a 5xx status code
func (o *GetRegistryForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this get registry forbidden response a status code equal to that given
func (o *GetRegistryForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *GetRegistryForbidden) Error() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryForbidden  %+v", 403, o.Payload)
}

func (o *GetRegistryForbidden) String() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryForbidden  %+v", 403, o.Payload)
}

func (o *GetRegistryForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *GetRegistryForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewGetRegistryNotFound creates a GetRegistryNotFound with default headers values
func NewGetRegistryNotFound() *GetRegistryNotFound {
	return &GetRegistryNotFound{}
}

/*
GetRegistryNotFound describes a response with status code 404, with default header values.

Not found
*/
type GetRegistryNotFound struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this get registry not found response has a 2xx status code
func (o *GetRegistryNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get registry not found response has a 3xx status code
func (o *GetRegistryNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get registry not found response has a 4xx status code
func (o *GetRegistryNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get registry not found response has a 5xx status code
func (o *GetRegistryNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get registry not found response a status code equal to that given
func (o *GetRegistryNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetRegistryNotFound) Error() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryNotFound  %+v", 404, o.Payload)
}

func (o *GetRegistryNotFound) String() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryNotFound  %+v", 404, o.Payload)
}

func (o *GetRegistryNotFound) GetPayload() *models.Errors {
	return o.Payload
}

func (o *GetRegistryNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewGetRegistryInternalServerError creates a GetRegistryInternalServerError with default headers values
func NewGetRegistryInternalServerError() *GetRegistryInternalServerError {
	return &GetRegistryInternalServerError{}
}

/*
GetRegistryInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type GetRegistryInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this get registry internal server error response has a 2xx status code
func (o *GetRegistryInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get registry internal server error response has a 3xx status code
func (o *GetRegistryInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get registry internal server error response has a 4xx status code
func (o *GetRegistryInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get registry internal server error response has a 5xx status code
func (o *GetRegistryInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get registry internal server error response a status code equal to that given
func (o *GetRegistryInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetRegistryInternalServerError) Error() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryInternalServerError  %+v", 500, o.Payload)
}

func (o *GetRegistryInternalServerError) String() string {
	return fmt.Sprintf("[GET /registries/{id}][%d] getRegistryInternalServerError  %+v", 500, o.Payload)
}

func (o *GetRegistryInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *GetRegistryInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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
