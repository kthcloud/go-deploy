// Code generated by go-swagger; DO NOT EDIT.

package robotv1

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// CreateRobotV1Reader is a Reader for the CreateRobotV1 structure.
type CreateRobotV1Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateRobotV1Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewCreateRobotV1Created()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewCreateRobotV1BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewCreateRobotV1Unauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewCreateRobotV1Forbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewCreateRobotV1NotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewCreateRobotV1InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewCreateRobotV1Created creates a CreateRobotV1Created with default headers values
func NewCreateRobotV1Created() *CreateRobotV1Created {
	return &CreateRobotV1Created{}
}

/*
CreateRobotV1Created describes a response with status code 201, with default header values.

Created
*/
type CreateRobotV1Created struct {

	/* The location of the resource
	 */
	Location string

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.RobotCreated
}

// IsSuccess returns true when this create robot v1 created response has a 2xx status code
func (o *CreateRobotV1Created) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this create robot v1 created response has a 3xx status code
func (o *CreateRobotV1Created) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create robot v1 created response has a 4xx status code
func (o *CreateRobotV1Created) IsClientError() bool {
	return false
}

// IsServerError returns true when this create robot v1 created response has a 5xx status code
func (o *CreateRobotV1Created) IsServerError() bool {
	return false
}

// IsCode returns true when this create robot v1 created response a status code equal to that given
func (o *CreateRobotV1Created) IsCode(code int) bool {
	return code == 201
}

func (o *CreateRobotV1Created) Error() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1Created  %+v", 201, o.Payload)
}

func (o *CreateRobotV1Created) String() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1Created  %+v", 201, o.Payload)
}

func (o *CreateRobotV1Created) GetPayload() *models.RobotCreated {
	return o.Payload
}

func (o *CreateRobotV1Created) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header Location
	hdrLocation := response.GetHeader("Location")

	if hdrLocation != "" {
		o.Location = hdrLocation
	}

	// hydrates response header X-Request-Id
	hdrXRequestID := response.GetHeader("X-Request-Id")

	if hdrXRequestID != "" {
		o.XRequestID = hdrXRequestID
	}

	o.Payload = new(models.RobotCreated)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateRobotV1BadRequest creates a CreateRobotV1BadRequest with default headers values
func NewCreateRobotV1BadRequest() *CreateRobotV1BadRequest {
	return &CreateRobotV1BadRequest{}
}

/*
CreateRobotV1BadRequest describes a response with status code 400, with default header values.

Bad request
*/
type CreateRobotV1BadRequest struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this create robot v1 bad request response has a 2xx status code
func (o *CreateRobotV1BadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this create robot v1 bad request response has a 3xx status code
func (o *CreateRobotV1BadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create robot v1 bad request response has a 4xx status code
func (o *CreateRobotV1BadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this create robot v1 bad request response has a 5xx status code
func (o *CreateRobotV1BadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this create robot v1 bad request response a status code equal to that given
func (o *CreateRobotV1BadRequest) IsCode(code int) bool {
	return code == 400
}

func (o *CreateRobotV1BadRequest) Error() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1BadRequest  %+v", 400, o.Payload)
}

func (o *CreateRobotV1BadRequest) String() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1BadRequest  %+v", 400, o.Payload)
}

func (o *CreateRobotV1BadRequest) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateRobotV1BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewCreateRobotV1Unauthorized creates a CreateRobotV1Unauthorized with default headers values
func NewCreateRobotV1Unauthorized() *CreateRobotV1Unauthorized {
	return &CreateRobotV1Unauthorized{}
}

/*
CreateRobotV1Unauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type CreateRobotV1Unauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this create robot v1 unauthorized response has a 2xx status code
func (o *CreateRobotV1Unauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this create robot v1 unauthorized response has a 3xx status code
func (o *CreateRobotV1Unauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create robot v1 unauthorized response has a 4xx status code
func (o *CreateRobotV1Unauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this create robot v1 unauthorized response has a 5xx status code
func (o *CreateRobotV1Unauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this create robot v1 unauthorized response a status code equal to that given
func (o *CreateRobotV1Unauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *CreateRobotV1Unauthorized) Error() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1Unauthorized  %+v", 401, o.Payload)
}

func (o *CreateRobotV1Unauthorized) String() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1Unauthorized  %+v", 401, o.Payload)
}

func (o *CreateRobotV1Unauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateRobotV1Unauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewCreateRobotV1Forbidden creates a CreateRobotV1Forbidden with default headers values
func NewCreateRobotV1Forbidden() *CreateRobotV1Forbidden {
	return &CreateRobotV1Forbidden{}
}

/*
CreateRobotV1Forbidden describes a response with status code 403, with default header values.

Forbidden
*/
type CreateRobotV1Forbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this create robot v1 forbidden response has a 2xx status code
func (o *CreateRobotV1Forbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this create robot v1 forbidden response has a 3xx status code
func (o *CreateRobotV1Forbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create robot v1 forbidden response has a 4xx status code
func (o *CreateRobotV1Forbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this create robot v1 forbidden response has a 5xx status code
func (o *CreateRobotV1Forbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this create robot v1 forbidden response a status code equal to that given
func (o *CreateRobotV1Forbidden) IsCode(code int) bool {
	return code == 403
}

func (o *CreateRobotV1Forbidden) Error() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1Forbidden  %+v", 403, o.Payload)
}

func (o *CreateRobotV1Forbidden) String() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1Forbidden  %+v", 403, o.Payload)
}

func (o *CreateRobotV1Forbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateRobotV1Forbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewCreateRobotV1NotFound creates a CreateRobotV1NotFound with default headers values
func NewCreateRobotV1NotFound() *CreateRobotV1NotFound {
	return &CreateRobotV1NotFound{}
}

/*
CreateRobotV1NotFound describes a response with status code 404, with default header values.

Not found
*/
type CreateRobotV1NotFound struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this create robot v1 not found response has a 2xx status code
func (o *CreateRobotV1NotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this create robot v1 not found response has a 3xx status code
func (o *CreateRobotV1NotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create robot v1 not found response has a 4xx status code
func (o *CreateRobotV1NotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this create robot v1 not found response has a 5xx status code
func (o *CreateRobotV1NotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this create robot v1 not found response a status code equal to that given
func (o *CreateRobotV1NotFound) IsCode(code int) bool {
	return code == 404
}

func (o *CreateRobotV1NotFound) Error() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1NotFound  %+v", 404, o.Payload)
}

func (o *CreateRobotV1NotFound) String() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1NotFound  %+v", 404, o.Payload)
}

func (o *CreateRobotV1NotFound) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateRobotV1NotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewCreateRobotV1InternalServerError creates a CreateRobotV1InternalServerError with default headers values
func NewCreateRobotV1InternalServerError() *CreateRobotV1InternalServerError {
	return &CreateRobotV1InternalServerError{}
}

/*
CreateRobotV1InternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type CreateRobotV1InternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this create robot v1 internal server error response has a 2xx status code
func (o *CreateRobotV1InternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this create robot v1 internal server error response has a 3xx status code
func (o *CreateRobotV1InternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this create robot v1 internal server error response has a 4xx status code
func (o *CreateRobotV1InternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this create robot v1 internal server error response has a 5xx status code
func (o *CreateRobotV1InternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this create robot v1 internal server error response a status code equal to that given
func (o *CreateRobotV1InternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *CreateRobotV1InternalServerError) Error() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1InternalServerError  %+v", 500, o.Payload)
}

func (o *CreateRobotV1InternalServerError) String() string {
	return fmt.Sprintf("[POST /projects/{project_name_or_id}/robots][%d] createRobotV1InternalServerError  %+v", 500, o.Payload)
}

func (o *CreateRobotV1InternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *CreateRobotV1InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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