package confd

// RawData is a map of raw data, it can be used to unmarshal json data
type RawData map[string]interface{}

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
