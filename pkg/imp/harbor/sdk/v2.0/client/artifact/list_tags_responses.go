// Code generated by go-swagger; DO NOT EDIT.

package artifact

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

// ListTagsReader is a Reader for the ListTags structure.
type ListTagsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListTagsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListTagsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewListTagsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewListTagsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewListTagsForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewListTagsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewListTagsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListTagsOK creates a ListTagsOK with default headers values
func NewListTagsOK() *ListTagsOK {
	return &ListTagsOK{}
}

/*
ListTagsOK describes a response with status code 200, with default header values.

Success
*/
type ListTagsOK struct {

	/* Link refers to the previous page and next page
	 */
	Link string

	/* The total count of tags
	 */
	XTotalCount int64

	Payload []*models.Tag
}

// IsSuccess returns true when this list tags o k response has a 2xx status code
func (o *ListTagsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list tags o k response has a 3xx status code
func (o *ListTagsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tags o k response has a 4xx status code
func (o *ListTagsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list tags o k response has a 5xx status code
func (o *ListTagsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list tags o k response a status code equal to that given
func (o *ListTagsOK) IsCode(code int) bool {
	return code == 200
}

func (o *ListTagsOK) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsOK  %+v", 200, o.Payload)
}

func (o *ListTagsOK) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsOK  %+v", 200, o.Payload)
}

func (o *ListTagsOK) GetPayload() []*models.Tag {
	return o.Payload
}

func (o *ListTagsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTagsBadRequest creates a ListTagsBadRequest with default headers values
func NewListTagsBadRequest() *ListTagsBadRequest {
	return &ListTagsBadRequest{}
}

/*
ListTagsBadRequest describes a response with status code 400, with default header values.

Bad request
*/
type ListTagsBadRequest struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tags bad request response has a 2xx status code
func (o *ListTagsBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tags bad request response has a 3xx status code
func (o *ListTagsBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tags bad request response has a 4xx status code
func (o *ListTagsBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this list tags bad request response has a 5xx status code
func (o *ListTagsBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this list tags bad request response a status code equal to that given
func (o *ListTagsBadRequest) IsCode(code int) bool {
	return code == 400
}

func (o *ListTagsBadRequest) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsBadRequest  %+v", 400, o.Payload)
}

func (o *ListTagsBadRequest) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsBadRequest  %+v", 400, o.Payload)
}

func (o *ListTagsBadRequest) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTagsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTagsUnauthorized creates a ListTagsUnauthorized with default headers values
func NewListTagsUnauthorized() *ListTagsUnauthorized {
	return &ListTagsUnauthorized{}
}

/*
ListTagsUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ListTagsUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tags unauthorized response has a 2xx status code
func (o *ListTagsUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tags unauthorized response has a 3xx status code
func (o *ListTagsUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tags unauthorized response has a 4xx status code
func (o *ListTagsUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this list tags unauthorized response has a 5xx status code
func (o *ListTagsUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this list tags unauthorized response a status code equal to that given
func (o *ListTagsUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ListTagsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsUnauthorized  %+v", 401, o.Payload)
}

func (o *ListTagsUnauthorized) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsUnauthorized  %+v", 401, o.Payload)
}

func (o *ListTagsUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTagsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTagsForbidden creates a ListTagsForbidden with default headers values
func NewListTagsForbidden() *ListTagsForbidden {
	return &ListTagsForbidden{}
}

/*
ListTagsForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type ListTagsForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tags forbidden response has a 2xx status code
func (o *ListTagsForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tags forbidden response has a 3xx status code
func (o *ListTagsForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tags forbidden response has a 4xx status code
func (o *ListTagsForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this list tags forbidden response has a 5xx status code
func (o *ListTagsForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this list tags forbidden response a status code equal to that given
func (o *ListTagsForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *ListTagsForbidden) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsForbidden  %+v", 403, o.Payload)
}

func (o *ListTagsForbidden) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsForbidden  %+v", 403, o.Payload)
}

func (o *ListTagsForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTagsForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTagsNotFound creates a ListTagsNotFound with default headers values
func NewListTagsNotFound() *ListTagsNotFound {
	return &ListTagsNotFound{}
}

/*
ListTagsNotFound describes a response with status code 404, with default header values.

Not found
*/
type ListTagsNotFound struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tags not found response has a 2xx status code
func (o *ListTagsNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tags not found response has a 3xx status code
func (o *ListTagsNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tags not found response has a 4xx status code
func (o *ListTagsNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this list tags not found response has a 5xx status code
func (o *ListTagsNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this list tags not found response a status code equal to that given
func (o *ListTagsNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *ListTagsNotFound) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsNotFound  %+v", 404, o.Payload)
}

func (o *ListTagsNotFound) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsNotFound  %+v", 404, o.Payload)
}

func (o *ListTagsNotFound) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTagsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewListTagsInternalServerError creates a ListTagsInternalServerError with default headers values
func NewListTagsInternalServerError() *ListTagsInternalServerError {
	return &ListTagsInternalServerError{}
}

/*
ListTagsInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type ListTagsInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this list tags internal server error response has a 2xx status code
func (o *ListTagsInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list tags internal server error response has a 3xx status code
func (o *ListTagsInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list tags internal server error response has a 4xx status code
func (o *ListTagsInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this list tags internal server error response has a 5xx status code
func (o *ListTagsInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this list tags internal server error response a status code equal to that given
func (o *ListTagsInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ListTagsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsInternalServerError  %+v", 500, o.Payload)
}

func (o *ListTagsInternalServerError) String() string {
	return fmt.Sprintf("[GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags][%d] listTagsInternalServerError  %+v", 500, o.Payload)
}

func (o *ListTagsInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *ListTagsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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
