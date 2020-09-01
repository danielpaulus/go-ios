package accessibility

func convertToStringList(payload []interface{}) []string {
	list := payload[0].([]interface{})
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}
