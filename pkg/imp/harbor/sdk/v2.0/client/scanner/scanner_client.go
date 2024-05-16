// Code generated by go-swagger; DO NOT EDIT.

package scanner

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

//go:generate mockery -name API -inpkg

// API is the interface of the scanner client
type API interface {
	/*
	   CreateScanner creates a scanner registration

	   Creats a new scanner registration with the given data.
	*/
	CreateScanner(ctx context.Context, params *CreateScannerParams) (*CreateScannerCreated, error)
	/*
	   DeleteScanner deletes a scanner registration

	   Deletes the specified scanner registration.
	*/
	DeleteScanner(ctx context.Context, params *DeleteScannerParams) (*DeleteScannerOK, error)
	/*
	   GetScanner gets a scanner registration details

	   Retruns the details of the specified scanner registration.
	*/
	GetScanner(ctx context.Context, params *GetScannerParams) (*GetScannerOK, error)
	/*
	   GetScannerMetadata gets the metadata of the specified scanner registration

	   Get the metadata of the specified scanner registration, including the capabilities and customized properties.
	*/
	GetScannerMetadata(ctx context.Context, params *GetScannerMetadataParams) (*GetScannerMetadataOK, error)
	/*
	   ListScanners lists scanner registrations

	   Returns a list of currently configured scanner registrations.
	*/
	ListScanners(ctx context.Context, params *ListScannersParams) (*ListScannersOK, error)
	/*
	   PingScanner tests scanner registration settings

	   Pings scanner adapter to test endpoint URL and authorization settings.
	*/
	PingScanner(ctx context.Context, params *PingScannerParams) (*PingScannerOK, error)
	/*
	   SetScannerAsDefault sets system default scanner registration

	   Set the specified scanner registration as the system default one.
	*/
	SetScannerAsDefault(ctx context.Context, params *SetScannerAsDefaultParams) (*SetScannerAsDefaultOK, error)
	/*
	   UpdateScanner updates a scanner registration

	   Updates the specified scanner registration.
	*/
	UpdateScanner(ctx context.Context, params *UpdateScannerParams) (*UpdateScannerOK, error)
}

// New creates a new scanner API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry, authInfo runtime.ClientAuthInfoWriter) *Client {
	return &Client{
		transport: transport,
		formats:   formats,
		authInfo:  authInfo,
	}
}

/*
Client for scanner API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
	authInfo  runtime.ClientAuthInfoWriter
}

/*
CreateScanner creates a scanner registration

Creats a new scanner registration with the given data.
*/
func (a *Client) CreateScanner(ctx context.Context, params *CreateScannerParams) (*CreateScannerCreated, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createScanner",
		Method:             "POST",
		PathPattern:        "/scanners",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CreateScannerReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*CreateScannerCreated), nil

}

/*
DeleteScanner deletes a scanner registration

Deletes the specified scanner registration.
*/
func (a *Client) DeleteScanner(ctx context.Context, params *DeleteScannerParams) (*DeleteScannerOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "deleteScanner",
		Method:             "DELETE",
		PathPattern:        "/scanners/{registration_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &DeleteScannerReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*DeleteScannerOK), nil

}

/*
GetScanner gets a scanner registration details

Retruns the details of the specified scanner registration.
*/
func (a *Client) GetScanner(ctx context.Context, params *GetScannerParams) (*GetScannerOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getScanner",
		Method:             "GET",
		PathPattern:        "/scanners/{registration_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetScannerReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*GetScannerOK), nil

}

/*
GetScannerMetadata gets the metadata of the specified scanner registration

Get the metadata of the specified scanner registration, including the capabilities and customized properties.
*/
func (a *Client) GetScannerMetadata(ctx context.Context, params *GetScannerMetadataParams) (*GetScannerMetadataOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getScannerMetadata",
		Method:             "GET",
		PathPattern:        "/scanners/{registration_id}/metadata",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetScannerMetadataReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*GetScannerMetadataOK), nil

}

/*
ListScanners lists scanner registrations

Returns a list of currently configured scanner registrations.
*/
func (a *Client) ListScanners(ctx context.Context, params *ListScannersParams) (*ListScannersOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "listScanners",
		Method:             "GET",
		PathPattern:        "/scanners",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ListScannersReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*ListScannersOK), nil

}

/*
PingScanner tests scanner registration settings

Pings scanner adapter to test endpoint URL and authorization settings.
*/
func (a *Client) PingScanner(ctx context.Context, params *PingScannerParams) (*PingScannerOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "pingScanner",
		Method:             "POST",
		PathPattern:        "/scanners/ping",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &PingScannerReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*PingScannerOK), nil

}

/*
SetScannerAsDefault sets system default scanner registration

Set the specified scanner registration as the system default one.
*/
func (a *Client) SetScannerAsDefault(ctx context.Context, params *SetScannerAsDefaultParams) (*SetScannerAsDefaultOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "setScannerAsDefault",
		Method:             "PATCH",
		PathPattern:        "/scanners/{registration_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &SetScannerAsDefaultReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*SetScannerAsDefaultOK), nil

}

/*
UpdateScanner updates a scanner registration

Updates the specified scanner registration.
*/
func (a *Client) UpdateScanner(ctx context.Context, params *UpdateScannerParams) (*UpdateScannerOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "updateScanner",
		Method:             "PUT",
		PathPattern:        "/scanners/{registration_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &UpdateScannerReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*UpdateScannerOK), nil

}
