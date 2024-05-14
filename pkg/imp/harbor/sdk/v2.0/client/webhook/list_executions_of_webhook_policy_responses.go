// Code generated by go-swagger; DO NOT EDIT.

package webhook

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

// ListExecutionsOfWebhookPolicyReader is a Reader for the ListExecutionsOfWebhookPolicy structure.
type ListExecutionsOfWebhookPolicyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListExecutionsOfWebhookPolicyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListExecutionsOfWebhookPolicyOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewListExecutionsOfWebhookPolicyBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewListExecutionsOfWebhookPolicyUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewListExecutionsOfWebhookPolicyForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewListExecutionsOfWebhookPolicyNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewListExecutionsOfWebhookPolicyInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListExecutionsOfWebhookPolicyOK creates a ListExecutionsOfWebhookPolicyOK with default headers values
func NewListExecutionsOfWebhookPolicyOK() *ListExecutionsOfWebhookPolicyOK {
	return &ListExecutionsOfWebhookPolicyOK{}
}

/*
ListExecutionsOfWebhookPolicyOK describes a response with status code 200, with default header values.

List webhook executions success
*/
type ListExecutionsOfWebhookPolicyOK struct {

	/* Link refers to the previous page and next page
	 */
	Link string

	/* The total count of executions
	 */
	XTotalCount int64

	Payload []*models.Execution
}

// IsSuccess returns true when this list executions of webhook policy o k response has a 2xx status code
func (o *ListExecutionsOfWebhookPolicyOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list executions of webhook policy o k response has a 3xx status code
func (o *ListExecutionsOfWebhookPolicyOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list executions of webhook policy o k response has a 4xx status code
func (o *ListExecutionsOfWebhookPolicyOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list executions of webhook policy o k response has a 5xx status code
func (o *ListExecutionsOfWebhookPolicyOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list executions of webhook policy o k response a status code equal to that given
func (o *ListExecutionsOfWebhookPolicyOK) IsCode(code int) bool {
	return code == 200
}

func (o *ListExecutionsOfWebhookPolicyOK) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyOK  %+v", 200, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyOK) String() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyOK  %+v", 200, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyOK) GetPayload() []*models.Execution {
	return o.Payload
}

func (o *ListExecutionsOfWebhookPolicyOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListExecutionsOfWebhookPolicyBadRequest creates a ListExecutionsOfWebhookPolicyBadRequest with default headers values
func NewListExecutionsOfWebhookPolicyBadRequest() *ListExecutionsOfWebhookPolicyBadRequest {
	return &ListExecutionsOfWebhookPolicyBadRequest{}
}

/*
ListExecutionsOfWebhookPolicyBadRequest describes a response with status code 400, with default header values.

Bad request
*/
type ListExecutionsOfWebhookPolicyBadRequest struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list executions of webhook policy bad request response has a 2xx status code
func (o *ListExecutionsOfWebhookPolicyBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list executions of webhook policy bad request response has a 3xx status code
func (o *ListExecutionsOfWebhookPolicyBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list executions of webhook policy bad request response has a 4xx status code
func (o *ListExecutionsOfWebhookPolicyBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this list executions of webhook policy bad request response has a 5xx status code
func (o *ListExecutionsOfWebhookPolicyBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this list executions of webhook policy bad request response a status code equal to that given
func (o *ListExecutionsOfWebhookPolicyBadRequest) IsCode(code int) bool {
	return code == 400
}

func (o *ListExecutionsOfWebhookPolicyBadRequest) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyBadRequest  %+v", 400, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyBadRequest) String() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyBadRequest  %+v", 400, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyBadRequest) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListExecutionsOfWebhookPolicyBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListExecutionsOfWebhookPolicyUnauthorized creates a ListExecutionsOfWebhookPolicyUnauthorized with default headers values
func NewListExecutionsOfWebhookPolicyUnauthorized() *ListExecutionsOfWebhookPolicyUnauthorized {
	return &ListExecutionsOfWebhookPolicyUnauthorized{}
}

/*
ListExecutionsOfWebhookPolicyUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ListExecutionsOfWebhookPolicyUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list executions of webhook policy unauthorized response has a 2xx status code
func (o *ListExecutionsOfWebhookPolicyUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list executions of webhook policy unauthorized response has a 3xx status code
func (o *ListExecutionsOfWebhookPolicyUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list executions of webhook policy unauthorized response has a 4xx status code
func (o *ListExecutionsOfWebhookPolicyUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this list executions of webhook policy unauthorized response has a 5xx status code
func (o *ListExecutionsOfWebhookPolicyUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this list executions of webhook policy unauthorized response a status code equal to that given
func (o *ListExecutionsOfWebhookPolicyUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ListExecutionsOfWebhookPolicyUnauthorized) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyUnauthorized  %+v", 401, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyUnauthorized) String() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyUnauthorized  %+v", 401, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListExecutionsOfWebhookPolicyUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListExecutionsOfWebhookPolicyForbidden creates a ListExecutionsOfWebhookPolicyForbidden with default headers values
func NewListExecutionsOfWebhookPolicyForbidden() *ListExecutionsOfWebhookPolicyForbidden {
	return &ListExecutionsOfWebhookPolicyForbidden{}
}

/*
ListExecutionsOfWebhookPolicyForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type ListExecutionsOfWebhookPolicyForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list executions of webhook policy forbidden response has a 2xx status code
func (o *ListExecutionsOfWebhookPolicyForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list executions of webhook policy forbidden response has a 3xx status code
func (o *ListExecutionsOfWebhookPolicyForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list executions of webhook policy forbidden response has a 4xx status code
func (o *ListExecutionsOfWebhookPolicyForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this list executions of webhook policy forbidden response has a 5xx status code
func (o *ListExecutionsOfWebhookPolicyForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this list executions of webhook policy forbidden response a status code equal to that given
func (o *ListExecutionsOfWebhookPolicyForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *ListExecutionsOfWebhookPolicyForbidden) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyForbidden  %+v", 403, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyForbidden) String() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyForbidden  %+v", 403, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListExecutionsOfWebhookPolicyForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListExecutionsOfWebhookPolicyNotFound creates a ListExecutionsOfWebhookPolicyNotFound with default headers values
func NewListExecutionsOfWebhookPolicyNotFound() *ListExecutionsOfWebhookPolicyNotFound {
	return &ListExecutionsOfWebhookPolicyNotFound{}
}

/*
ListExecutionsOfWebhookPolicyNotFound describes a response with status code 404, with default header values.

Not found
*/
type ListExecutionsOfWebhookPolicyNotFound struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list executions of webhook policy not found response has a 2xx status code
func (o *ListExecutionsOfWebhookPolicyNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list executions of webhook policy not found response has a 3xx status code
func (o *ListExecutionsOfWebhookPolicyNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list executions of webhook policy not found response has a 4xx status code
func (o *ListExecutionsOfWebhookPolicyNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this list executions of webhook policy not found response has a 5xx status code
func (o *ListExecutionsOfWebhookPolicyNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this list executions of webhook policy not found response a status code equal to that given
func (o *ListExecutionsOfWebhookPolicyNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *ListExecutionsOfWebhookPolicyNotFound) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyNotFound  %+v", 404, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyNotFound) String() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyNotFound  %+v", 404, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyNotFound) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListExecutionsOfWebhookPolicyNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListExecutionsOfWebhookPolicyInternalServerError creates a ListExecutionsOfWebhookPolicyInternalServerError with default headers values
func NewListExecutionsOfWebhookPolicyInternalServerError() *ListExecutionsOfWebhookPolicyInternalServerError {
	return &ListExecutionsOfWebhookPolicyInternalServerError{}
}

/*
ListExecutionsOfWebhookPolicyInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type ListExecutionsOfWebhookPolicyInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list executions of webhook policy internal server error response has a 2xx status code
func (o *ListExecutionsOfWebhookPolicyInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list executions of webhook policy internal server error response has a 3xx status code
func (o *ListExecutionsOfWebhookPolicyInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list executions of webhook policy internal server error response has a 4xx status code
func (o *ListExecutionsOfWebhookPolicyInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this list executions of webhook policy internal server error response has a 5xx status code
func (o *ListExecutionsOfWebhookPolicyInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this list executions of webhook policy internal server error response a status code equal to that given
func (o *ListExecutionsOfWebhookPolicyInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ListExecutionsOfWebhookPolicyInternalServerError) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyInternalServerError) String() string {
	return fmt.Sprintf("[GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}/executions][%d] listExecutionsOfWebhookPolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *ListExecutionsOfWebhookPolicyInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListExecutionsOfWebhookPolicyInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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
