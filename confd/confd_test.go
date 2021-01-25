package confd_test

import (
	"testing"

	"github.com/cloudogu/ces-confd/confd"
	"github.com/stretchr/testify/assert"
)

func TestGetStringValue(t *testing.T) {
	raw := confd.RawData{}

	t.Run("should return name", func(t *testing.T) {
		raw["name"] = "heartOfGold"

		stringValue := raw.GetStringValue("name")
		assert.Equal(t, "heartOfGold", stringValue)
	})

	t.Run("should return empty string when requested key does not exist ", func(t *testing.T) {
		raw["name"] = "heartOfGold"

		stringValue := raw.GetStringValue("displayname")
		assert.Empty(t, stringValue)
	})

	t.Run("should return empty string when value is not a string", func(t *testing.T) {
		raw["name"] = false

		stringValue := raw.GetStringValue("name")
		assert.Empty(t, stringValue)
	})
}

func TestGetAttributeValue(t *testing.T) {
	raw := confd.RawData{}

	t.Run("should return attribute value", func(t *testing.T) {
		attributes := map[string]interface{}{
			"location": "heartOfGoldLocation",
		}
		raw["attributes"] = attributes

		attributeValue := raw.GetAttributeValue("location")
		assert.Equal(t, "heartOfGoldLocation", attributeValue)
	})

	t.Run("should return empty string when attribute key does not exist", func(t *testing.T) {
		raw := confd.RawData{}

		attributeValue := raw.GetAttributeValue("location")
		assert.Empty(t, attributeValue)
	})

	t.Run("should return empty string when attribute value is not a map", func(t *testing.T) {
		raw["attributes"] = "location:heartOfGoldLocation"

		attributeValue := raw.GetAttributeValue("location")
		assert.Empty(t, attributeValue)
	})

	t.Run("should return empty string when requested attribute key does not exist", func(t *testing.T) {
		attributes := map[string]interface{}{
			"location": "heartOfGoldLocation",
		}
		raw["attributes"] = attributes

		attributeValue := raw.GetAttributeValue("place")
		assert.Empty(t, attributeValue)
	})
}
