package confd_test

import (
	"testing"

	"github.com/cloudogu/ces-confd/confd"
	"github.com/stretchr/testify/assert"
)

func TestGetStringValue(t *testing.T) {
	raw := confd.RawData{}
	raw["name"] = "heartOfGold"

	stringValue := raw.GetStringValue("name")

	assert.Equal(t, "heartOfGold", stringValue)
}

func TestGetStringValueWhenKeyNotExist(t *testing.T) {
	raw := confd.RawData{}
	raw["name"] = "heartOfGold"

	stringValue := raw.GetStringValue("displayname")

	assert.Empty(t, stringValue)
}

func TestGetStringValueWithNonStringValue(t *testing.T) {
	raw := confd.RawData{}
	raw["name"] = false

	stringValue := raw.GetStringValue("name")

	assert.Empty(t, stringValue)
}

func TestGetAttributes(t *testing.T) {
	raw := confd.RawData{}

	attributes := map[string]interface{}{
		"location": "heartOfGoldLocation",
	}
	raw["attributes"] = attributes
	attributeValue := raw.GetAttributeValue("location")

	assert.Equal(t, "heartOfGoldLocation", attributeValue)
}

func TestGetAttributesWithoutAttributesKey(t *testing.T) {
	raw := confd.RawData{}
	attributeValue := raw.GetAttributeValue("location")

	assert.Empty(t, attributeValue)
}

func TestGetAttributesWithNonMapValue(t *testing.T) {
	raw := confd.RawData{}
	raw["attributes"] = "location:heartOfGoldLocation"
	attributeValue := raw.GetAttributeValue("location")

	assert.Empty(t, attributeValue)
}

func TestGetAttributesWithNonExistingAttributeKey(t *testing.T) {
	raw := confd.RawData{}

	attributes := map[string]interface{}{
		"location": "heartOfGoldLocation",
	}
	raw["attributes"] = attributes
	attributeValue := raw.GetAttributeValue("place")

	assert.Empty(t, attributeValue)
}
