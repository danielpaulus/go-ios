package accessibility

import "fmt"

func convertToStringList(payload []interface{}) ([]string, error) {
	if len(payload) != 1 {
		return nil, fmt.Errorf("invalid payload length %d", len(payload))
	}
	list := payload[0].([]interface{})
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result, nil
}
