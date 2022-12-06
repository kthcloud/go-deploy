package app

import (
	"deploy-api-go/pkg/validator"
)

func (context *ClientContext) ValidateJSON(rules *validator.MapData, output interface{}) map[string][]string {
	dummy := validator.MapData{}
	return context.ValidateJSONCustomMessages(rules, &dummy, output)
}

func (context *ClientContext) ValidateJSONCustomMessages(rules *validator.MapData, messages *validator.MapData, output interface{}) map[string][]string {
	opts := validator.Options{
		Request:         context.GinContext.Request,
		Context:         context.GinContext,
		Rules:           *rules,
		RequiredDefault: true,
		Messages:        *messages,
		Data:            output,
	}

	v := validator.New(opts)
	validationErrs := v.ValidateJSON()

	return validationErrs
}

func (context *ClientContext) ValidateParams(rules *validator.MapData) map[string][]string {
	dummy := validator.MapData{}
	return context.ValidateParamsCustomMessages(rules, &dummy)
}

func (context *ClientContext) ValidateParamsCustomMessages(rules *validator.MapData, messages *validator.MapData) map[string][]string {
	opts := validator.Options{
		Request:         context.GinContext.Request,
		Context:         context.GinContext,
		Rules:           *rules,
		RequiredDefault: true,
		Messages:        *messages,
	}

	v := validator.New(opts)
	validationErrs := v.ValidateParams()

	return validationErrs
}
