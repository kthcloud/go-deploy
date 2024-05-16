// Code generated by go-swagger; DO NOT EDIT.

package immutable

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

//go:generate mockery -name API -inpkg

// API is the interface of the immutable client
type API interface {
	/*
	   CreateImmuRule adds an immutable tag rule to current project

	   This endpoint add an immutable tag rule to the project
	*/
	CreateImmuRule(ctx context.Context, params *CreateImmuRuleParams) (*CreateImmuRuleCreated, error)
	/*
	   DeleteImmuRule deletes the immutable tag rule*/
	DeleteImmuRule(ctx context.Context, params *DeleteImmuRuleParams) (*DeleteImmuRuleOK, error)
	/*
	   ListImmuRules lists all immutable tag rules of current project

	   This endpoint returns the immutable tag rules of a project
	*/
	ListImmuRules(ctx context.Context, params *ListImmuRulesParams) (*ListImmuRulesOK, error)
	/*
	   UpdateImmuRule updates the immutable tag rule or enable or disable the rule*/
	UpdateImmuRule(ctx context.Context, params *UpdateImmuRuleParams) (*UpdateImmuRuleOK, error)
}

// New creates a new immutable API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry, authInfo runtime.ClientAuthInfoWriter) *Client {
	return &Client{
		transport: transport,
		formats:   formats,
		authInfo:  authInfo,
	}
}

/*
Client for immutable API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
	authInfo  runtime.ClientAuthInfoWriter
}

/*
CreateImmuRule adds an immutable tag rule to current project

This endpoint add an immutable tag rule to the project
*/
func (a *Client) CreateImmuRule(ctx context.Context, params *CreateImmuRuleParams) (*CreateImmuRuleCreated, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "CreateImmuRule",
		Method:             "POST",
		PathPattern:        "/projects/{project_name_or_id}/immutabletagrules",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CreateImmuRuleReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*CreateImmuRuleCreated), nil

}

/*
DeleteImmuRule deletes the immutable tag rule
*/
func (a *Client) DeleteImmuRule(ctx context.Context, params *DeleteImmuRuleParams) (*DeleteImmuRuleOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "DeleteImmuRule",
		Method:             "DELETE",
		PathPattern:        "/projects/{project_name_or_id}/immutabletagrules/{immutable_rule_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &DeleteImmuRuleReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*DeleteImmuRuleOK), nil

}

/*
ListImmuRules lists all immutable tag rules of current project

This endpoint returns the immutable tag rules of a project
*/
func (a *Client) ListImmuRules(ctx context.Context, params *ListImmuRulesParams) (*ListImmuRulesOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "ListImmuRules",
		Method:             "GET",
		PathPattern:        "/projects/{project_name_or_id}/immutabletagrules",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ListImmuRulesReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*ListImmuRulesOK), nil

}

/*
UpdateImmuRule updates the immutable tag rule or enable or disable the rule
*/
func (a *Client) UpdateImmuRule(ctx context.Context, params *UpdateImmuRuleParams) (*UpdateImmuRuleOK, error) {

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "UpdateImmuRule",
		Method:             "PUT",
		PathPattern:        "/projects/{project_name_or_id}/immutabletagrules/{immutable_rule_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &UpdateImmuRuleReader{formats: a.formats},
		AuthInfo:           a.authInfo,
		Context:            ctx,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*UpdateImmuRuleOK), nil

}
