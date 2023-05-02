package body

type BindingError struct {
	ValidationErrors map[string][]string `json:"validationErrors"`
}
