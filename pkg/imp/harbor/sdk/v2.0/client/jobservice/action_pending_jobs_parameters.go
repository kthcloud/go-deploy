// Code generated by go-swagger; DO NOT EDIT.

package jobservice

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/models"
)

// NewActionPendingJobsParams creates a new ActionPendingJobsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewActionPendingJobsParams() *ActionPendingJobsParams {
	return &ActionPendingJobsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewActionPendingJobsParamsWithTimeout creates a new ActionPendingJobsParams object
// with the ability to set a timeout on a request.
func NewActionPendingJobsParamsWithTimeout(timeout time.Duration) *ActionPendingJobsParams {
	return &ActionPendingJobsParams{
		timeout: timeout,
	}
}

// NewActionPendingJobsParamsWithContext creates a new ActionPendingJobsParams object
// with the ability to set a context for a request.
func NewActionPendingJobsParamsWithContext(ctx context.Context) *ActionPendingJobsParams {
	return &ActionPendingJobsParams{
		Context: ctx,
	}
}

// NewActionPendingJobsParamsWithHTTPClient creates a new ActionPendingJobsParams object
// with the ability to set a custom HTTPClient for a request.
func NewActionPendingJobsParamsWithHTTPClient(client *http.Client) *ActionPendingJobsParams {
	return &ActionPendingJobsParams{
		HTTPClient: client,
	}
}

/*
ActionPendingJobsParams contains all the parameters to send to the API endpoint

	for the action pending jobs operation.

	Typically these are written to a http.Request.
*/
type ActionPendingJobsParams struct {

	/* XRequestID.

	   An unique ID for the request
	*/
	XRequestID *string

	// ActionRequest.
	ActionRequest *models.ActionRequest

	/* JobType.

	   The type of the job. 'all' stands for all job types
	*/
	JobType string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the action pending jobs params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ActionPendingJobsParams) WithDefaults() *ActionPendingJobsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the action pending jobs params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ActionPendingJobsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the action pending jobs params
func (o *ActionPendingJobsParams) WithTimeout(timeout time.Duration) *ActionPendingJobsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the action pending jobs params
func (o *ActionPendingJobsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the action pending jobs params
func (o *ActionPendingJobsParams) WithContext(ctx context.Context) *ActionPendingJobsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the action pending jobs params
func (o *ActionPendingJobsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the action pending jobs params
func (o *ActionPendingJobsParams) WithHTTPClient(client *http.Client) *ActionPendingJobsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the action pending jobs params
func (o *ActionPendingJobsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithXRequestID adds the xRequestID to the action pending jobs params
func (o *ActionPendingJobsParams) WithXRequestID(xRequestID *string) *ActionPendingJobsParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the action pending jobs params
func (o *ActionPendingJobsParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithActionRequest adds the actionRequest to the action pending jobs params
func (o *ActionPendingJobsParams) WithActionRequest(actionRequest *models.ActionRequest) *ActionPendingJobsParams {
	o.SetActionRequest(actionRequest)
	return o
}

// SetActionRequest adds the actionRequest to the action pending jobs params
func (o *ActionPendingJobsParams) SetActionRequest(actionRequest *models.ActionRequest) {
	o.ActionRequest = actionRequest
}

// WithJobType adds the jobType to the action pending jobs params
func (o *ActionPendingJobsParams) WithJobType(jobType string) *ActionPendingJobsParams {
	o.SetJobType(jobType)
	return o
}

// SetJobType adds the jobType to the action pending jobs params
func (o *ActionPendingJobsParams) SetJobType(jobType string) {
	o.JobType = jobType
}

// WriteToRequest writes these params to a swagger request
func (o *ActionPendingJobsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.XRequestID != nil {

		// header param X-Request-Id
		if err := r.SetHeaderParam("X-Request-Id", *o.XRequestID); err != nil {
			return err
		}
	}
	if o.ActionRequest != nil {
		if err := r.SetBodyParam(o.ActionRequest); err != nil {
			return err
		}
	}

	// path param job_type
	if err := r.SetPathParam("job_type", o.JobType); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
