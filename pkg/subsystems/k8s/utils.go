package k8s

// GetLabel returns the value of a label from a map.
// If the label does not exist, an empty string is returned.
func GetLabel(labels map[string]string, key string) string {
	if labels == nil {
		return ""
	}
	if v, ok := labels[key]; ok {
		return v
	}
	return ""
}
