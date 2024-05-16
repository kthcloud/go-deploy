// Code generated by go-swagger; DO NOT EDIT.

package health

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

//go:generate mockery -name API -inpkg

// API is the interface of the health client
type API interface {
	/*
	   GetHealth checks the status of harbor components

	   Check the status of Harbor components. This path does not require authentication.*/
	GetHealth(ctx context.Context, params *GetHealthParams) (*GetHealthOK, error)
}

// New creates a new health API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry, authInfo runtime.ClientAuthInfoWriter) *Client {
	return &Client{
		transport: transport,
		formats:   formats,
		authInfo:  authInfo,
	}
}

/*
Client for health API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
	authInfo  runtime.ClientAuthInfoWriter
}

/*
GetHealth checks the status of harbor components

Check the status of Harbor components. This path does not require authentication.
*/
func (a *Client) GetHealth(ctx context.Context, params *GetHealthParams) (*GetHealthOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getHealth",
		Method:             "GET",
		PathPattern:        "/health",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetHealthReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*GetHealthOK), nil

}
