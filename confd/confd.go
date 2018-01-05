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
