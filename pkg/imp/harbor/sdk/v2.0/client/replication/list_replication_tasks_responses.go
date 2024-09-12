// Code generated by go-swagger; DO NOT EDIT.

package replication

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// ListReplicationTasksReader is a Reader for the ListReplicationTasks structure.
type ListReplicationTasksReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListReplicationTasksReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListReplicationTasksOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewListReplicationTasksUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewListReplicationTasksForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewListReplicationTasksInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListReplicationTasksOK creates a ListReplicationTasksOK with default headers values
func NewListReplicationTasksOK() *ListReplicationTasksOK {
	return &ListReplicationTasksOK{}
}

/*
ListReplicationTasksOK describes a response with status code 200, with default header values.

Success
*/
type ListReplicationTasksOK struct {

	/* Link refers to the previous page and next page
	 */
	Link string

	/* The total count of the resources
	 */
	XTotalCount int64

	Payload []*models.ReplicationTask
}

// IsSuccess returns true when this list replication tasks o k response has a 2xx status code
func (o *ListReplicationTasksOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list replication tasks o k response has a 3xx status code
func (o *ListReplicationTasksOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list replication tasks o k response has a 4xx status code
func (o *ListReplicationTasksOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list replication tasks o k response has a 5xx status code
func (o *ListReplicationTasksOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list replication tasks o k response a status code equal to that given
func (o *ListReplicationTasksOK) IsCode(code int) bool {
	return code == 200
}

func (o *ListReplicationTasksOK) Error() string {
	return fmt.Sprintf("[GET /replication/executions/{id}/tasks][%d] listReplicationTasksOK  %+v", 200, o.Payload)
}

func (o *ListReplicationTasksOK) String() string {
	return fmt.Sprintf("[GET /replication/executions/{id}/tasks][%d] listReplicationTasksOK  %+v", 200, o.Payload)
}

func (o *ListReplicationTasksOK) GetPayload() []*models.ReplicationTask {
	return o.Payload
}

func (o *ListReplicationTasksOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListReplicationTasksUnauthorized creates a ListReplicationTasksUnauthorized with default headers values
func NewListReplicationTasksUnauthorized() *ListReplicationTasksUnauthorized {
	return &ListReplicationTasksUnauthorized{}
}

/*
ListReplicationTasksUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ListReplicationTasksUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list replication tasks unauthorized response has a 2xx status code
func (o *ListReplicationTasksUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list replication tasks unauthorized response has a 3xx status code
func (o *ListReplicationTasksUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list replication tasks unauthorized response has a 4xx status code
func (o *ListReplicationTasksUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this list replication tasks unauthorized response has a 5xx status code
func (o *ListReplicationTasksUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this list replication tasks unauthorized response a status code equal to that given
func (o *ListReplicationTasksUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ListReplicationTasksUnauthorized) Error() string {
	return fmt.Sprintf("[GET /replication/executions/{id}/tasks][%d] listReplicationTasksUnauthorized  %+v", 401, o.Payload)
}

func (o *ListReplicationTasksUnauthorized) String() string {
	return fmt.Sprintf("[GET /replication/executions/{id}/tasks][%d] listReplicationTasksUnauthorized  %+v", 401, o.Payload)
}

func (o *ListReplicationTasksUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListReplicationTasksUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListReplicationTasksForbidden creates a ListReplicationTasksForbidden with default headers values
func NewListReplicationTasksForbidden() *ListReplicationTasksForbidden {
	return &ListReplicationTasksForbidden{}
}

/*
ListReplicationTasksForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type ListReplicationTasksForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list replication tasks forbidden response has a 2xx status code
func (o *ListReplicationTasksForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list replication tasks forbidden response has a 3xx status code
func (o *ListReplicationTasksForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list replication tasks forbidden response has a 4xx status code
func (o *ListReplicationTasksForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this list replication tasks forbidden response has a 5xx status code
func (o *ListReplicationTasksForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this list replication tasks forbidden response a status code equal to that given
func (o *ListReplicationTasksForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *ListReplicationTasksForbidden) Error() string {
	return fmt.Sprintf("[GET /replication/executions/{id}/tasks][%d] listReplicationTasksForbidden  %+v", 403, o.Payload)
}

func (o *ListReplicationTasksForbidden) String() string {
	return fmt.Sprintf("[GET /replication/executions/{id}/tasks][%d] listReplicationTasksForbidden  %+v", 403, o.Payload)
}

func (o *ListReplicationTasksForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListReplicationTasksForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListReplicationTasksInternalServerError creates a ListReplicationTasksInternalServerError with default headers values
func NewListReplicationTasksInternalServerError() *ListReplicationTasksInternalServerError {
	return &ListReplicationTasksInternalServerError{}
}

/*
ListReplicationTasksInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type ListReplicationTasksInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list replication tasks internal server error response has a 2xx status code
func (o *ListReplicationTasksInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list replication tasks internal server error response has a 3xx status code
func (o *ListReplicationTasksInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list replication tasks internal server error response has a 4xx status code
func (o *ListReplicationTasksInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this list replication tasks internal server error response has a 5xx status code
func (o *ListReplicationTasksInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this list replication tasks internal server error response a status code equal to that given
func (o *ListReplicationTasksInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ListReplicationTasksInternalServerError) Error() string {
	return fmt.Sprintf("[GET /replication/executions/{id}/tasks][%d] listReplicationTasksInternalServerError  %+v", 500, o.Payload)
}

func (o *ListReplicationTasksInternalServerError) String() string {
	return fmt.Sprintf("[GET /replication/executions/{id}/tasks][%d] listReplicationTasksInternalServerError  %+v", 500, o.Payload)
}

func (o *ListReplicationTasksInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListReplicationTasksInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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
