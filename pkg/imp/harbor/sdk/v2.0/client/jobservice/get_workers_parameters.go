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
)

// NewGetWorkersParams creates a new GetWorkersParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetWorkersParams() *GetWorkersParams {
	return &GetWorkersParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetWorkersParamsWithTimeout creates a new GetWorkersParams object
// with the ability to set a timeout on a request.
func NewGetWorkersParamsWithTimeout(timeout time.Duration) *GetWorkersParams {
	return &GetWorkersParams{
		timeout: timeout,
	}
}

// NewGetWorkersParamsWithContext creates a new GetWorkersParams object
// with the ability to set a context for a request.
func NewGetWorkersParamsWithContext(ctx context.Context) *GetWorkersParams {
	return &GetWorkersParams{
		Context: ctx,
	}
}

// NewGetWorkersParamsWithHTTPClient creates a new GetWorkersParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetWorkersParamsWithHTTPClient(client *http.Client) *GetWorkersParams {
	return &GetWorkersParams{
		HTTPClient: client,
	}
}

/*
GetWorkersParams contains all the parameters to send to the API endpoint

	for the get workers operation.

	Typically these are written to a http.Request.
*/
type GetWorkersParams struct {

	/* XRequestID.

	   An unique ID for the request
	*/
	XRequestID *string

	/* PoolID.

	   The name of the pool. 'all' stands for all pools
	*/
	PoolID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get workers params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetWorkersParams) WithDefaults() *GetWorkersParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get workers params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetWorkersParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get workers params
func (o *GetWorkersParams) WithTimeout(timeout time.Duration) *GetWorkersParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get workers params
func (o *GetWorkersParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get workers params
func (o *GetWorkersParams) WithContext(ctx context.Context) *GetWorkersParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get workers params
func (o *GetWorkersParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get workers params
func (o *GetWorkersParams) WithHTTPClient(client *http.Client) *GetWorkersParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get workers params
func (o *GetWorkersParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithXRequestID adds the xRequestID to the get workers params
func (o *GetWorkersParams) WithXRequestID(xRequestID *string) *GetWorkersParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the get workers params
func (o *GetWorkersParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithPoolID adds the poolID to the get workers params
func (o *GetWorkersParams) WithPoolID(poolID string) *GetWorkersParams {
	o.SetPoolID(poolID)
	return o
}

// SetPoolID adds the poolId to the get workers params
func (o *GetWorkersParams) SetPoolID(poolID string) {
	o.PoolID = poolID
}

// WriteToRequest writes these params to a swagger request
func (o *GetWorkersParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param pool_id
	if err := r.SetPathParam("pool_id", o.PoolID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
