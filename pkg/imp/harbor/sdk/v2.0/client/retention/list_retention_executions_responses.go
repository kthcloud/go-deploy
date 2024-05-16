// Code generated by go-swagger; DO NOT EDIT.

package retention

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// ListRetentionExecutionsReader is a Reader for the ListRetentionExecutions structure.
type ListRetentionExecutionsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListRetentionExecutionsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListRetentionExecutionsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewListRetentionExecutionsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewListRetentionExecutionsForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewListRetentionExecutionsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListRetentionExecutionsOK creates a ListRetentionExecutionsOK with default headers values
func NewListRetentionExecutionsOK() *ListRetentionExecutionsOK {
	return &ListRetentionExecutionsOK{}
}

/*
ListRetentionExecutionsOK describes a response with status code 200, with default header values.

Get a Retention execution successfully.
*/
type ListRetentionExecutionsOK struct {

	/* Link to previous page and next page
	 */
	Link string

	/* The total count of available items
	 */
	XTotalCount int64

	Payload []*models.RetentionExecution
}

// IsSuccess returns true when this list retention executions o k response has a 2xx status code
func (o *ListRetentionExecutionsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list retention executions o k response has a 3xx status code
func (o *ListRetentionExecutionsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list retention executions o k response has a 4xx status code
func (o *ListRetentionExecutionsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list retention executions o k response has a 5xx status code
func (o *ListRetentionExecutionsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list retention executions o k response a status code equal to that given
func (o *ListRetentionExecutionsOK) IsCode(code int) bool {
	return code == 200
}

func (o *ListRetentionExecutionsOK) Error() string {
	return fmt.Sprintf("[GET /retentions/{id}/executions][%d] listRetentionExecutionsOK  %+v", 200, o.Payload)
}

func (o *ListRetentionExecutionsOK) String() string {
	return fmt.Sprintf("[GET /retentions/{id}/executions][%d] listRetentionExecutionsOK  %+v", 200, o.Payload)
}

func (o *ListRetentionExecutionsOK) GetPayload() []*models.RetentionExecution {
	return o.Payload
}

func (o *ListRetentionExecutionsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header Link
	hdrLink := response.GetHeader("Link")

	if hdrLink != "" {
		o.Link = hdrLink
	}

	// hydrates response header X-Total-Count
	hdrXTotalCount := response.GetHeader("X-Total-Count")

	if hdrXTotalCount != "" {
		valxTotalCount, err := swag.ConvertInt64(hdrXTotalCount)
		if err != nil {
			return errors.InvalidType("X-Total-Count", "header", "int64", hdrXTotalCount)
		}
		o.XTotalCount = valxTotalCount
	}

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListRetentionExecutionsUnauthorized creates a ListRetentionExecutionsUnauthorized with default headers values
func NewListRetentionExecutionsUnauthorized() *ListRetentionExecutionsUnauthorized {
	return &ListRetentionExecutionsUnauthorized{}
}

/*
ListRetentionExecutionsUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ListRetentionExecutionsUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list retention executions unauthorized response has a 2xx status code
func (o *ListRetentionExecutionsUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list retention executions unauthorized response has a 3xx status code
func (o *ListRetentionExecutionsUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list retention executions unauthorized response has a 4xx status code
func (o *ListRetentionExecutionsUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this list retention executions unauthorized response has a 5xx status code
func (o *ListRetentionExecutionsUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this list retention executions unauthorized response a status code equal to that given
func (o *ListRetentionExecutionsUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ListRetentionExecutionsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /retentions/{id}/executions][%d] listRetentionExecutionsUnauthorized  %+v", 401, o.Payload)
}

func (o *ListRetentionExecutionsUnauthorized) String() string {
	return fmt.Sprintf("[GET /retentions/{id}/executions][%d] listRetentionExecutionsUnauthorized  %+v", 401, o.Payload)
}

func (o *ListRetentionExecutionsUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListRetentionExecutionsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListRetentionExecutionsForbidden creates a ListRetentionExecutionsForbidden with default headers values
func NewListRetentionExecutionsForbidden() *ListRetentionExecutionsForbidden {
	return &ListRetentionExecutionsForbidden{}
}

/*
ListRetentionExecutionsForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type ListRetentionExecutionsForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list retention executions forbidden response has a 2xx status code
func (o *ListRetentionExecutionsForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list retention executions forbidden response has a 3xx status code
func (o *ListRetentionExecutionsForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list retention executions forbidden response has a 4xx status code
func (o *ListRetentionExecutionsForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this list retention executions forbidden response has a 5xx status code
func (o *ListRetentionExecutionsForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this list retention executions forbidden response a status code equal to that given
func (o *ListRetentionExecutionsForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *ListRetentionExecutionsForbidden) Error() string {
	return fmt.Sprintf("[GET /retentions/{id}/executions][%d] listRetentionExecutionsForbidden  %+v", 403, o.Payload)
}

func (o *ListRetentionExecutionsForbidden) String() string {
	return fmt.Sprintf("[GET /retentions/{id}/executions][%d] listRetentionExecutionsForbidden  %+v", 403, o.Payload)
}

func (o *ListRetentionExecutionsForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListRetentionExecutionsForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListRetentionExecutionsInternalServerError creates a ListRetentionExecutionsInternalServerError with default headers values
func NewListRetentionExecutionsInternalServerError() *ListRetentionExecutionsInternalServerError {
	return &ListRetentionExecutionsInternalServerError{}
}

/*
ListRetentionExecutionsInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type ListRetentionExecutionsInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list retention executions internal server error response has a 2xx status code
func (o *ListRetentionExecutionsInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list retention executions internal server error response has a 3xx status code
func (o *ListRetentionExecutionsInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list retention executions internal server error response has a 4xx status code
func (o *ListRetentionExecutionsInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this list retention executions internal server error response has a 5xx status code
func (o *ListRetentionExecutionsInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this list retention executions internal server error response a status code equal to that given
func (o *ListRetentionExecutionsInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ListRetentionExecutionsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /retentions/{id}/executions][%d] listRetentionExecutionsInternalServerError  %+v", 500, o.Payload)
}

func (o *ListRetentionExecutionsInternalServerError) String() string {
	return fmt.Sprintf("[GET /retentions/{id}/executions][%d] listRetentionExecutionsInternalServerError  %+v", 500, o.Payload)
}

func (o *ListRetentionExecutionsInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListRetentionExecutionsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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
