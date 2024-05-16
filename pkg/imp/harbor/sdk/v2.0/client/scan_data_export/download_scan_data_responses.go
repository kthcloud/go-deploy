// Code generated by go-swagger; DO NOT EDIT.

package scan_data_export

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// DownloadScanDataReader is a Reader for the DownloadScanData structure.
type DownloadScanDataReader struct {
	formats strfmt.Registry
	writer  io.Writer
}

// ReadResponse reads a server response into the received o.
func (o *DownloadScanDataReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewDownloadScanDataOK(o.writer)
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewDownloadScanDataUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewDownloadScanDataForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewDownloadScanDataNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewDownloadScanDataInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewDownloadScanDataOK creates a DownloadScanDataOK with default headers values
func NewDownloadScanDataOK(writer io.Writer) *DownloadScanDataOK {
	return &DownloadScanDataOK{

		Payload: writer,
	}
}

/*
DownloadScanDataOK describes a response with status code 200, with default header values.

Data file containing the export data
*/
type DownloadScanDataOK struct {

	/* Value is a CSV formatted file; filename=export.csv
	 */
	ContentDisposition string

	Payload io.Writer
}

// IsSuccess returns true when this download scan data o k response has a 2xx status code
func (o *DownloadScanDataOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this download scan data o k response has a 3xx status code
func (o *DownloadScanDataOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this download scan data o k response has a 4xx status code
func (o *DownloadScanDataOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this download scan data o k response has a 5xx status code
func (o *DownloadScanDataOK) IsServerError() bool {
	return false
}

// IsCode returns true when this download scan data o k response a status code equal to that given
func (o *DownloadScanDataOK) IsCode(code int) bool {
	return code == 200
}

func (o *DownloadScanDataOK) Error() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataOK  %+v", 200, o.Payload)
}

func (o *DownloadScanDataOK) String() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataOK  %+v", 200, o.Payload)
}

func (o *DownloadScanDataOK) GetPayload() io.Writer {
	return o.Payload
}

func (o *DownloadScanDataOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header Content-Disposition
	hdrContentDisposition := response.GetHeader("Content-Disposition")

	if hdrContentDisposition != "" {
		o.ContentDisposition = hdrContentDisposition
	}

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDownloadScanDataUnauthorized creates a DownloadScanDataUnauthorized with default headers values
func NewDownloadScanDataUnauthorized() *DownloadScanDataUnauthorized {
	return &DownloadScanDataUnauthorized{}
}

/*
DownloadScanDataUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type DownloadScanDataUnauthorized struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this download scan data unauthorized response has a 2xx status code
func (o *DownloadScanDataUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this download scan data unauthorized response has a 3xx status code
func (o *DownloadScanDataUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this download scan data unauthorized response has a 4xx status code
func (o *DownloadScanDataUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this download scan data unauthorized response has a 5xx status code
func (o *DownloadScanDataUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this download scan data unauthorized response a status code equal to that given
func (o *DownloadScanDataUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *DownloadScanDataUnauthorized) Error() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataUnauthorized  %+v", 401, o.Payload)
}

func (o *DownloadScanDataUnauthorized) String() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataUnauthorized  %+v", 401, o.Payload)
}

func (o *DownloadScanDataUnauthorized) GetPayload() *models.Errors {
	return o.Payload
}

func (o *DownloadScanDataUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewDownloadScanDataForbidden creates a DownloadScanDataForbidden with default headers values
func NewDownloadScanDataForbidden() *DownloadScanDataForbidden {
	return &DownloadScanDataForbidden{}
}

/*
DownloadScanDataForbidden describes a response with status code 403, with default header values.

Forbidden
*/
type DownloadScanDataForbidden struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this download scan data forbidden response has a 2xx status code
func (o *DownloadScanDataForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this download scan data forbidden response has a 3xx status code
func (o *DownloadScanDataForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this download scan data forbidden response has a 4xx status code
func (o *DownloadScanDataForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this download scan data forbidden response has a 5xx status code
func (o *DownloadScanDataForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this download scan data forbidden response a status code equal to that given
func (o *DownloadScanDataForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *DownloadScanDataForbidden) Error() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataForbidden  %+v", 403, o.Payload)
}

func (o *DownloadScanDataForbidden) String() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataForbidden  %+v", 403, o.Payload)
}

func (o *DownloadScanDataForbidden) GetPayload() *models.Errors {
	return o.Payload
}

func (o *DownloadScanDataForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewDownloadScanDataNotFound creates a DownloadScanDataNotFound with default headers values
func NewDownloadScanDataNotFound() *DownloadScanDataNotFound {
	return &DownloadScanDataNotFound{}
}

/*
DownloadScanDataNotFound describes a response with status code 404, with default header values.

Not found
*/
type DownloadScanDataNotFound struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this download scan data not found response has a 2xx status code
func (o *DownloadScanDataNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this download scan data not found response has a 3xx status code
func (o *DownloadScanDataNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this download scan data not found response has a 4xx status code
func (o *DownloadScanDataNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this download scan data not found response has a 5xx status code
func (o *DownloadScanDataNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this download scan data not found response a status code equal to that given
func (o *DownloadScanDataNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *DownloadScanDataNotFound) Error() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataNotFound  %+v", 404, o.Payload)
}

func (o *DownloadScanDataNotFound) String() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataNotFound  %+v", 404, o.Payload)
}

func (o *DownloadScanDataNotFound) GetPayload() *models.Errors {
	return o.Payload
}

func (o *DownloadScanDataNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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

// NewDownloadScanDataInternalServerError creates a DownloadScanDataInternalServerError with default headers values
func NewDownloadScanDataInternalServerError() *DownloadScanDataInternalServerError {
	return &DownloadScanDataInternalServerError{}
}

/*
DownloadScanDataInternalServerError describes a response with status code 500, with default header values.

Internal server error
*/
type DownloadScanDataInternalServerError struct {

	/* The ID of the corresponding request for the response
	 */
	XRequestID string

	Payload *models.Errors
}

// IsSuccess returns true when this download scan data internal server error response has a 2xx status code
func (o *DownloadScanDataInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this download scan data internal server error response has a 3xx status code
func (o *DownloadScanDataInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this download scan data internal server error response has a 4xx status code
func (o *DownloadScanDataInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this download scan data internal server error response has a 5xx status code
func (o *DownloadScanDataInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this download scan data internal server error response a status code equal to that given
func (o *DownloadScanDataInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *DownloadScanDataInternalServerError) Error() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataInternalServerError  %+v", 500, o.Payload)
}

func (o *DownloadScanDataInternalServerError) String() string {
	return fmt.Sprintf("[GET /export/cve/download/{execution_id}][%d] downloadScanDataInternalServerError  %+v", 500, o.Payload)
}

func (o *DownloadScanDataInternalServerError) GetPayload() *models.Errors {
	return o.Payload
}

func (o *DownloadScanDataInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

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