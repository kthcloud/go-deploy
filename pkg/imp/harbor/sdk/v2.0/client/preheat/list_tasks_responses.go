// Code generated by go-swagger; DO NOT EDIT.

package preheat

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

// ListTasksReader is a Reader for the ListTasks structure.
type ListTasksReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListTasksReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListTasksOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewListTasksBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewListTasksUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewListTasksForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewListTasksNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewListTasksInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListTasksOK creates a ListTasksOK with default headers values
func NewListTasksOK() *ListTasksOK {
	return &ListTasksOK{}
}

/*
ListTasksOK describes a response with status code 200, with default header values.

List tasks success
*/
type ListTasksOK struct {

	/* Link refers to the previous page and next page
	 */
	Link string

	/* The total count of tasks
	 */
	XTotalCount int64

	Payload []*models.Task
}

// IsSuccess returns true when this list tasks o k response has a 2xx status code
func (o *ListTasksOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list tasks o k response has a 3xx status code
func (o *ListTasksOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tasks o k response has a 4xx status code
func (o *ListTasksOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list tasks o k response has a 5xx status code
func (o *ListTasksOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list tasks o k response a status code equal to that given
func (o *ListTasksOK) IsCode(code int) bool {
	return code == 200
}

func (o *ListTasksOK) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksOK  %+v", 200, o.Payload)
}

func (o *ListTasksOK) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksOK  %+v", 200, o.Payload)
}

func (o *ListTasksOK) GetPayload() []*models.Task {
	return o.Payload
}

func (o *ListTasksOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTasksBadRequest creates a ListTasksBadRequest with default headers values
func NewListTasksBadRequest() *ListTasksBadRequest {
	return &ListTasksBadRequest{}
}

/*
ListTasksBadRequest describes a response with status code 400, with default header values.

Bad request
*/
type ListTasksBadRequest struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tasks bad request response has a 2xx status code
func (o *ListTasksBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tasks bad request response has a 3xx status code
func (o *ListTasksBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tasks bad request response has a 4xx status code
func (o *ListTasksBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this list tasks bad request response has a 5xx status code
func (o *ListTasksBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this list tasks bad request response a status code equal to that given
func (o *ListTasksBadRequest) IsCode(code int) bool {
	return code == 400
}

func (o *ListTasksBadRequest) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksBadRequest  %+v", 400, o.Payload)
}

func (o *ListTasksBadRequest) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksBadRequest  %+v", 400, o.Payload)
}

func (o *ListTasksBadRequest) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTasksBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTasksUnauthorized creates a ListTasksUnauthorized with default headers values
func NewListTasksUnauthorized() *ListTasksUnauthorized {
	return &ListTasksUnauthorized{}
}

/*
ListTasksUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ListTasksUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tasks unauthorized response has a 2xx status code
func (o *ListTasksUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tasks unauthorized response has a 3xx status code
func (o *ListTasksUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tasks unauthorized response has a 4xx status code
func (o *ListTasksUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this list tasks unauthorized response has a 5xx status code
func (o *ListTasksUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this list tasks unauthorized response a status code equal to that given
func (o *ListTasksUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ListTasksUnauthorized) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksUnauthorized  %+v", 401, o.Payload)
}

func (o *ListTasksUnauthorized) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksUnauthorized  %+v", 401, o.Payload)
}

func (o *ListTasksUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTasksUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTasksForbidden creates a ListTasksForbidden with default headers values
func NewListTasksForbidden() *ListTasksForbidden {
	return &ListTasksForbidden{}
}

/*
ListTasksForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type ListTasksForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tasks forbidden response has a 2xx status code
func (o *ListTasksForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tasks forbidden response has a 3xx status code
func (o *ListTasksForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tasks forbidden response has a 4xx status code
func (o *ListTasksForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this list tasks forbidden response has a 5xx status code
func (o *ListTasksForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this list tasks forbidden response a status code equal to that given
func (o *ListTasksForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *ListTasksForbidden) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksForbidden  %+v", 403, o.Payload)
}

func (o *ListTasksForbidden) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksForbidden  %+v", 403, o.Payload)
}

func (o *ListTasksForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTasksForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTasksNotFound creates a ListTasksNotFound with default headers values
func NewListTasksNotFound() *ListTasksNotFound {
	return &ListTasksNotFound{}
}

/*
ListTasksNotFound describes a response with status code 404, with default header values.

Not found
*/
type ListTasksNotFound struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tasks not found response has a 2xx status code
func (o *ListTasksNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tasks not found response has a 3xx status code
func (o *ListTasksNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tasks not found response has a 4xx status code
func (o *ListTasksNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this list tasks not found response has a 5xx status code
func (o *ListTasksNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this list tasks not found response a status code equal to that given
func (o *ListTasksNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *ListTasksNotFound) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksNotFound  %+v", 404, o.Payload)
}

func (o *ListTasksNotFound) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksNotFound  %+v", 404, o.Payload)
}

func (o *ListTasksNotFound) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTasksNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTasksInternalServerError creates a ListTasksInternalServerError with default headers values
func NewListTasksInternalServerError() *ListTasksInternalServerError {
	return &ListTasksInternalServerError{}
}

/*
ListTasksInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type ListTasksInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tasks internal server error response has a 2xx status code
func (o *ListTasksInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tasks internal server error response has a 3xx status code
func (o *ListTasksInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tasks internal server error response has a 4xx status code
func (o *ListTasksInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this list tasks internal server error response has a 5xx status code
func (o *ListTasksInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this list tasks internal server error response a status code equal to that given
func (o *ListTasksInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ListTasksInternalServerError) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksInternalServerError  %+v", 500, o.Payload)
}

func (o *ListTasksInternalServerError) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks][%d] listTasksInternalServerError  %+v", 500, o.Payload)
}

func (o *ListTasksInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTasksInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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
