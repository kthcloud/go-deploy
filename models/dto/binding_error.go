package dto

type BindingError struct {
	ValidationErrors map[string][]string `json:"validationErrors"`
}
