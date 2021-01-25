package confd

// RawData is a map of raw data, it can be used to unmarshal json data
type RawData map[string]interface{}

// GetStringValue returns string from RawData.
// An empty string is returned if the key could not be found or the value is not a string.
func (data RawData) GetStringValue(key string) string {
	value, ok := data[key]
	if !ok {
		return ""
	}

	stringValue, ok := value.(string)
	if !ok {
		return ""
	}

	return stringValue
}

// GetAttributeValue returns string from RawData with key 'attributes' and the attribute key passed as parameter.
// An empty string is returned if no value can be found for 'attributes' or no attribute key exists with the key passed as parameter.
func (data RawData) GetAttributeValue(key string) string {
	attributesInterface, ok := data["attributes"]
	if !ok {
		return ""
	}

	attributeMap, ok := attributesInterface.(map[string]interface{})
	if !ok {
		return ""
	}

	value, ok := attributeMap[key].(string)
	if !ok {
		return ""
	}

	return value
}

// Order can be used to modify ordering via configuration
type Order map[string]int

// Contains returns true if the slice contains the item
func Contains(slice []interface{}, item interface{}) bool {
	for _, sliceItem := range slice {
		if sliceItem == item {
			return true
		}
	}
	return false
}

// ContainsString returns true if the slice contains the item
func ContainsString(slice []string, item string) bool {
	for _, sliceItem := range slice {
		if sliceItem == item {
			return true
		}
	}
	return false
}
