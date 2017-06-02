package confd


// RawData is a map of raw data, it can be used to unmarshal json data
type RawData map[string]interface{}

type Order map[string]int

func Contains(s []interface{}, e interface{}) bool {
  for _, a := range s {
    if a == e {
      return true
    }
  }
  return false
}

