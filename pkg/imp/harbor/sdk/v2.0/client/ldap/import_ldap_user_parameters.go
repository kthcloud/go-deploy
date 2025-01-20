// Code generated by go-swagger; DO NOT EDIT.

package ldap

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

// NewImportLdapUserParams creates a new ImportLdapUserParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewImportLdapUserParams() *ImportLdapUserParams {
	return &ImportLdapUserParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewImportLdapUserParamsWithTimeout creates a new ImportLdapUserParams object
// with the ability to set a timeout on a request.
func NewImportLdapUserParamsWithTimeout(timeout time.Duration) *ImportLdapUserParams {
	return &ImportLdapUserParams{
		timeout: timeout,
	}
}

// NewImportLdapUserParamsWithContext creates a new ImportLdapUserParams object
// with the ability to set a context for a request.
func NewImportLdapUserParamsWithContext(ctx context.Context) *ImportLdapUserParams {
	return &ImportLdapUserParams{
		Context: ctx,
	}
}

// NewImportLdapUserParamsWithHTTPClient creates a new ImportLdapUserParams object
// with the ability to set a custom HTTPClient for a request.
func NewImportLdapUserParamsWithHTTPClient(client *http.Client) *ImportLdapUserParams {
	return &ImportLdapUserParams{
		HTTPClient: client,
	}
}

/*
ImportLdapUserParams contains all the parameters to send to the API endpoint

	for the import ldap user operation.

	Typically these are written to a http.Request.
*/
type ImportLdapUserParams struct {

	/* XRequestID.

	   An unique ID for the request
	*/
	XRequestID *string

	/* UIDList.

	   The uid listed for importing. This list will check users validity of ldap service based on configuration from the system.
	*/
	UIDList *models.LdapImportUsers

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the import ldap user params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ImportLdapUserParams) WithDefaults() *ImportLdapUserParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the import ldap user params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ImportLdapUserParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the import ldap user params
func (o *ImportLdapUserParams) WithTimeout(timeout time.Duration) *ImportLdapUserParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the import ldap user params
func (o *ImportLdapUserParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the import ldap user params
func (o *ImportLdapUserParams) WithContext(ctx context.Context) *ImportLdapUserParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the import ldap user params
func (o *ImportLdapUserParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the import ldap user params
func (o *ImportLdapUserParams) WithHTTPClient(client *http.Client) *ImportLdapUserParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the import ldap user params
func (o *ImportLdapUserParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithXRequestID adds the xRequestID to the import ldap user params
func (o *ImportLdapUserParams) WithXRequestID(xRequestID *string) *ImportLdapUserParams {
	o.SetXRequestID(xRequestID)
	return o
}

// SetXRequestID adds the xRequestId to the import ldap user params
func (o *ImportLdapUserParams) SetXRequestID(xRequestID *string) {
	o.XRequestID = xRequestID
}

// WithUIDList adds the uIDList to the import ldap user params
func (o *ImportLdapUserParams) WithUIDList(uIDList *models.LdapImportUsers) *ImportLdapUserParams {
	o.SetUIDList(uIDList)
	return o
}

// SetUIDList adds the uidList to the import ldap user params
func (o *ImportLdapUserParams) SetUIDList(uIDList *models.LdapImportUsers) {
	o.UIDList = uIDList
}

// WriteToRequest writes these params to a swagger request
func (o *ImportLdapUserParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
	if o.UIDList != nil {
		if err := r.SetBodyParam(o.UIDList); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
