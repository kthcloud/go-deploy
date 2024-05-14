// Code generated by go-swagger; DO NOT EDIT.

package retention

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
	"github.com/go-openapi/swag"
)

// NewOperateRetentionExecutionParams creates a new OperateRetentionExecutionParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewOperateRetentionExecutionParams() *OperateRetentionExecutionParams {
	return &OperateRetentionExecutionParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewOperateRetentionExecutionParamsWithTimeout creates a new OperateRetentionExecutionParams object
// with the ability to set a timeout on a request.
func NewOperateRetentionExecutionParamsWithTimeout(timeout time.Duration) *OperateRetentionExecutionParams {
	return &OperateRetentionExecutionParams{
		timeout: timeout,
	}
}

// NewOperateRetentionExecutionParamsWithContext creates a new OperateRetentionExecutionParams object
// with the ability to set a context for a request.
func NewOperateRetentionExecutionParamsWithContext(ctx context.Context) *OperateRetentionExecutionParams {
	return &OperateRetentionExecutionParams{
		Context: ctx,
	}
}

// NewOperateRetentionExecutionParamsWithHTTPClient creates a new OperateRetentionExecutionParams object
// with the ability to set a custom HTTPClient for a request.
func NewOperateRetentionExecutionParamsWithHTTPClient(client *http.Client) *OperateRetentionExecutionParams {
	return &OperateRetentionExecutionParams{
		HTTPClient: client,
	}
}

/*
OperateRetentionExecutionParams contains all the parameters to send to the API endpoint

	for the operate retention execution operation.

	Typically these are written to a http.Request.
*/
type OperateRetentionExecutionParams struct {

	/* XRequestID.

	   An unique ID for the request
	*/
	XRequestID *string

	/* Body.

	   The action, only support "stop" now.
	*/
	Body OperateRetentionExecutionBody

	/* Eid.

	   Retention execution ID.

	   Format: int64
	*/
	Eid int64

	/* ID.

	   Retention ID.

	   Format: int64
	*/
	ID int64

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the operate retention execution params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *OperateRetentionExecutionParams) WithDefaults() *OperateRetentionExecutionParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the operate retention execution params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *OperateRetentionExecutionParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the operate retention execution params
func (o *OperateRetentionExecutionParams) WithTimeout(timeout time.Duration) *OperateRetentionExecutionParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the operate retention execution params
func (o *OperateRetentionExecutionParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the operate retention execution params
func (o *OperateRetentionExecutionParams) WithContext(ctx context.Context) *OperateRetentionExecutionParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the operate retention execution params
func (o *OperateRetentionExecutionParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the operate retention execution params
func (o *OperateRetentionExecutionParams) WithHTTPClient(client *http.Client) *OperateRetentionExecutionParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the operate retention execution params
func (o *OperateRetentionExecutionParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithXRequestID adds the xRequestID to the operate retention execution params
func (o *OperateRetentionExecutionParams) WithXRequestID(xRequestID *string) *OperateRetentionExecutionParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the operate retention execution params
func (o *OperateRetentionExecutionParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithBody adds the body to the operate retention execution params
func (o *OperateRetentionExecutionParams) WithBody(body OperateRetentionExecutionBody) *OperateRetentionExecutionParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the operate retention execution params
func (o *OperateRetentionExecutionParams) SetBody(body OperateRetentionExecutionBody) {
	o.Body = body
}

// WithEid adds the eid to the operate retention execution params
func (o *OperateRetentionExecutionParams) WithEid(eid int64) *OperateRetentionExecutionParams {
	o.SetEid(eid)
	return o
}

// SetEid adds the eid to the operate retention execution params
func (o *OperateRetentionExecutionParams) SetEid(eid int64) {
	o.Eid = eid
}

// WithID adds the id to the operate retention execution params
func (o *OperateRetentionExecutionParams) WithID(id int64) *OperateRetentionExecutionParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the operate retention execution params
func (o *OperateRetentionExecutionParams) SetID(id int64) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *OperateRetentionExecutionParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	// path param eid
	if err := r.SetPathParam("eid", swag.FormatInt64(o.Eid)); err != nil {
		return err
	}

	// path param id
	if err := r.SetPathParam("id", swag.FormatInt64(o.ID)); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
