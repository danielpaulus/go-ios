package accessibility

import (
	"fmt"
)

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

// getNestedMap safely extracts a nested map from a map[string]interface{}.
func getNestedMap(m map[string]interface{}, key string) (map[string]interface{}, error) {
	val, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found", key)
	}

	result, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key %q is not a map, got %T", key, val)
	}

	return result, nil
}

// getInnerValue extracts the inner value map from the response.
// Path: resp["Value"]["Value"]
func getInnerValue(resp map[string]interface{}) (map[string]interface{}, error) {
	value, err := getNestedMap(resp, "Value")
	if err != nil {
		return nil, fmt.Errorf("failed to get Value from response: %w", err)
	}

	innerValue, err := getNestedMap(value, "Value")
	if err != nil {
		return nil, fmt.Errorf("failed to get inner Value: %w", err)
	}

	return innerValue, nil
}

// deserializeObject unwraps {ObjectType: 'passthrough', Value: ...}
// and recursively processes containers(slices and maps). For other typed objects, returns their Value recursively.
func deserializeObject(d interface{}) interface{} {
	switch t := d.(type) {
	case []interface{}:
		out := make([]interface{}, 0, len(t))
		for _, v := range t {
			out = append(out, deserializeObject(v))
		}
		return out
	case map[string]interface{}:
		if ot, ok := t["ObjectType"]; ok {
			if ot == "passthrough" {
				return deserializeObject(t["Value"])
			}
			// For other typed objects, we generally care about their 'Value'
			if v, ok := t["Value"]; ok {
				return deserializeObject(v)
			}
			return t
		}
		// Plain dictionary: recursively process values
		out := make(map[string]interface{}, len(t))
		for k, v := range t {
			out[k] = deserializeObject(v)
		}
		return out
	default:
		return d
	}
}

func (a ControlInterface) extractStringFromField(innerValue map[string]interface{}, fieldName string) string {
	raw, ok := innerValue[fieldName]
	if !ok {
		return ""
	}

	val := deserializeObject(raw)
	if s, ok := val.(string); ok && s != "" {
		return s
	}

	if val != nil {
		desc := fmt.Sprintf("%v", val)
		return desc
	}

	return ""
}
