package convutils

func ToNameMap[IN any, OUT any](
	items []IN,
	getName func(IN) string,
	convert func(IN) OUT,
) map[string]OUT {
	result := make(map[string]OUT, len(items))
	for _, item := range items {
		name := getName(item)
		result[name] = convert(item)
	}
	return result
}
