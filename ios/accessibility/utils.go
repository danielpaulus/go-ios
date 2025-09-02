package accessibility

import "fmt"

func convertToStringList(payload []interface{}) ([]string, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is nil")
	}
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
