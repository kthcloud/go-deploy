// Code generated by go-swagger; DO NOT EDIT.

package retention

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// OperateRetentionExecutionReader is a Reader for the OperateRetentionExecution structure.
type OperateRetentionExecutionReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *OperateRetentionExecutionReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewOperateRetentionExecutionOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewOperateRetentionExecutionUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewOperateRetentionExecutionForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewOperateRetentionExecutionInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewOperateRetentionExecutionOK creates a OperateRetentionExecutionOK with default headers values
func NewOperateRetentionExecutionOK() *OperateRetentionExecutionOK {
	return &OperateRetentionExecutionOK{}
}

/*
OperateRetentionExecutionOK describes a response with status code 200, with default header values.

Stop a Retention job successfully.
*/
type OperateRetentionExecutionOK struct {
}

// IsSuccess returns true when this operate retention execution o k response has a 2xx status code
func (o *OperateRetentionExecutionOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this operate retention execution o k response has a 3xx status code
func (o *OperateRetentionExecutionOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this operate retention execution o k response has a 4xx status code
func (o *OperateRetentionExecutionOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this operate retention execution o k response has a 5xx status code
func (o *OperateRetentionExecutionOK) IsServerError() bool {
	return false
}

// IsCode returns true when this operate retention execution o k response a status code equal to that given
func (o *OperateRetentionExecutionOK) IsCode(code int) bool {
	return code == 200
}

func (o *OperateRetentionExecutionOK) Error() string {
	return fmt.Sprintf("[PATCH /retentions/{id}/executions/{eid}][%d] operateRetentionExecutionOK ", 200)
}

func (o *OperateRetentionExecutionOK) String() string {
	return fmt.Sprintf("[PATCH /retentions/{id}/executions/{eid}][%d] operateRetentionExecutionOK ", 200)
}

func (o *OperateRetentionExecutionOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewOperateRetentionExecutionUnauthorized creates a OperateRetentionExecutionUnauthorized with default headers values
func NewOperateRetentionExecutionUnauthorized() *OperateRetentionExecutionUnauthorized {
	return &OperateRetentionExecutionUnauthorized{}
}

/*
OperateRetentionExecutionUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type OperateRetentionExecutionUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this operate retention execution unauthorized response has a 2xx status code
func (o *OperateRetentionExecutionUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this operate retention execution unauthorized response has a 3xx status code
func (o *OperateRetentionExecutionUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this operate retention execution unauthorized response has a 4xx status code
func (o *OperateRetentionExecutionUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this operate retention execution unauthorized response has a 5xx status code
func (o *OperateRetentionExecutionUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this operate retention execution unauthorized response a status code equal to that given
func (o *OperateRetentionExecutionUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *OperateRetentionExecutionUnauthorized) Error() string {
	return fmt.Sprintf("[PATCH /retentions/{id}/executions/{eid}][%d] operateRetentionExecutionUnauthorized  %+v", 401, o.Payload)
}

func (o *OperateRetentionExecutionUnauthorized) String() string {
	return fmt.Sprintf("[PATCH /retentions/{id}/executions/{eid}][%d] operateRetentionExecutionUnauthorized  %+v", 401, o.Payload)
}

func (o *OperateRetentionExecutionUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *OperateRetentionExecutionUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewOperateRetentionExecutionForbidden creates a OperateRetentionExecutionForbidden with default headers values
func NewOperateRetentionExecutionForbidden() *OperateRetentionExecutionForbidden {
	return &OperateRetentionExecutionForbidden{}
}

/*
OperateRetentionExecutionForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type OperateRetentionExecutionForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this operate retention execution forbidden response has a 2xx status code
func (o *OperateRetentionExecutionForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this operate retention execution forbidden response has a 3xx status code
func (o *OperateRetentionExecutionForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this operate retention execution forbidden response has a 4xx status code
func (o *OperateRetentionExecutionForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this operate retention execution forbidden response has a 5xx status code
func (o *OperateRetentionExecutionForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this operate retention execution forbidden response a status code equal to that given
func (o *OperateRetentionExecutionForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *OperateRetentionExecutionForbidden) Error() string {
	return fmt.Sprintf("[PATCH /retentions/{id}/executions/{eid}][%d] operateRetentionExecutionForbidden  %+v", 403, o.Payload)
}

func (o *OperateRetentionExecutionForbidden) String() string {
	return fmt.Sprintf("[PATCH /retentions/{id}/executions/{eid}][%d] operateRetentionExecutionForbidden  %+v", 403, o.Payload)
}

func (o *OperateRetentionExecutionForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *OperateRetentionExecutionForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewOperateRetentionExecutionInternalServerError creates a OperateRetentionExecutionInternalServerError with default headers values
func NewOperateRetentionExecutionInternalServerError() *OperateRetentionExecutionInternalServerError {
	return &OperateRetentionExecutionInternalServerError{}
}

/*
OperateRetentionExecutionInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type OperateRetentionExecutionInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this operate retention execution internal server error response has a 2xx status code
func (o *OperateRetentionExecutionInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this operate retention execution internal server error response has a 3xx status code
func (o *OperateRetentionExecutionInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this operate retention execution internal server error response has a 4xx status code
func (o *OperateRetentionExecutionInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this operate retention execution internal server error response has a 5xx status code
func (o *OperateRetentionExecutionInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this operate retention execution internal server error response a status code equal to that given
func (o *OperateRetentionExecutionInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *OperateRetentionExecutionInternalServerError) Error() string {
	return fmt.Sprintf("[PATCH /retentions/{id}/executions/{eid}][%d] operateRetentionExecutionInternalServerError  %+v", 500, o.Payload)
}

func (o *OperateRetentionExecutionInternalServerError) String() string {
	return fmt.Sprintf("[PATCH /retentions/{id}/executions/{eid}][%d] operateRetentionExecutionInternalServerError  %+v", 500, o.Payload)
}

func (o *OperateRetentionExecutionInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *OperateRetentionExecutionInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

/*
OperateRetentionExecutionBody operate retention execution body
swagger:model OperateRetentionExecutionBody
*/
type OperateRetentionExecutionBody struct {

	// action
	Action string `json:"action,omitempty"`
}

// Validate validates this operate retention execution body
func (o *OperateRetentionExecutionBody) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this operate retention execution body based on context it is used
func (o *OperateRetentionExecutionBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *OperateRetentionExecutionBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *OperateRetentionExecutionBody) UnmarshalBinary(b []byte) error {
	var res OperateRetentionExecutionBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
