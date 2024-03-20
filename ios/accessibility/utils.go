package accessibility

func convertToStringList(payload []interface{}) []string {
	if len(payload) == 0 {
		return make([]string, 0)
	}
	list := payload[0].([]interface{})
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}
