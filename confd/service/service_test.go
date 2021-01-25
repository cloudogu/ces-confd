package service

import (
	"fmt"
	"testing"

	"github.com/cloudogu/ces-confd/confd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServicesString(t *testing.T) {
	services := Services{}

	heartOfGold := &Service{Name: "heartOfGold", URL: "http://8.8.8.8", HealthStatus: "healthy", Location: "heartOfGoldLocation"}
	services = append(services, heartOfGold)
	content := fmt.Sprintf("services: %v", services)
	assert.Equal(t, "services: [{name=heartOfGold, URL=http://8.8.8.8, HealthStatus=healthy, Location=heartOfGoldLocation}]", content)
}

func TestCreateService(t *testing.T) {
	raw := confd.RawData{}

	t.Run("should return created service ", func(t *testing.T) {
		raw["name"] = "heartOfGold"
		raw["service"] = "8.8.8.8"

		attributes := map[string]interface{}{
			"location": "heartOfGoldLocation",
		}
		raw["attributes"] = attributes

		service := createService(raw)
		assert.Equal(t, "heartOfGold", service.Name)
		assert.Equal(t, "http://8.8.8.8", service.URL)
		assert.Equal(t, "heartOfGoldLocation", service.Location)
	})

	t.Run("should return nil when name is missing ", func(t *testing.T) {
		raw := confd.RawData{}
		raw["service"] = "8.8.8.8"

		service := createService(raw)
		require.Nil(t, service)
	})

	t.Run("should return nil when value for service is missing", func(t *testing.T) {
		raw := confd.RawData{}
		raw["name"] = "heartOfGold"

		service := createService(raw)
		require.Nil(t, service)
	})

	t.Run("should return nil when name is not a string", func(t *testing.T) {
		raw["name"] = false
		raw["service"] = "8.8.8.8"

		service := createService(raw)
		require.Nil(t, service)
	})

	t.Run("should return nil when value for service is not a string", func(t *testing.T) {
		raw["name"] = "heartOfGold"
		raw["service"] = false

		service := createService(raw)
		require.Nil(t, service)
	})

	t.Run("should return created service even without attributes", func(t *testing.T) {
		raw := confd.RawData{}
		raw["name"] = "heartOfGold"
		raw["service"] = "8.8.8.8"

		service := createService(raw)
		assert.Equal(t, "heartOfGold", service.Location)
	})

	t.Run("should return created service even when attributes is not a map", func(t *testing.T) {
		raw["name"] = "heartOfGold"
		raw["service"] = "8.8.8.8"
		raw["attributes"] = "location:heartOfGoldLocation"

		service := createService(raw)
		assert.Equal(t, "heartOfGold", service.Location)
	})

	t.Run("should return created service even when location attribute does not exist", func(t *testing.T) {
		raw["name"] = "heartOfGold"
		raw["service"] = "8.8.8.8"

		attributes := map[string]interface{}{
			"day":   "Friday",
			"month": "December",
			"year":  2022,
		}
		raw["attributes"] = attributes

		service := createService(raw)
		assert.Equal(t, "heartOfGold", service.Name)
		assert.Equal(t, "http://8.8.8.8", service.URL)
		assert.Equal(t, "heartOfGold", service.Location)
	})
}
