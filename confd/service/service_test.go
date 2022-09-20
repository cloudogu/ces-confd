package service

import (
	"fmt"
	"github.com/cloudogu/ces-confd/confd"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServicesString(t *testing.T) {
	services := Services{}

	heartOfGold := &Service{
		Name: "heartOfGold",
		URL:  "http://8.8.8.8", HealthStatus: "healthy",
		Location: "heartOfGoldLocation",
		Rewrite: &Rewrite{
			Pattern: "rewriteme",
			Rewrite: "iwillrewriteyou",
		},
	}
	services = append(services, heartOfGold)
	content := fmt.Sprintf("services: %v", services)
	assert.Equal(t, "services: [{name=heartOfGold, URL=http://8.8.8.8, HealthStatus=healthy, Location=heartOfGoldLocation, Rewrite=&{Pattern:rewriteme Rewrite:iwillrewriteyou}}]", content)
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

		service, err := createService(raw)
		require.NoError(t, err)
		assert.Equal(t, "heartOfGold", service.Name)
		assert.Equal(t, "http://8.8.8.8", service.URL)
		assert.Equal(t, "heartOfGoldLocation", service.Location)
	})

	t.Run("should return nil when name is missing ", func(t *testing.T) {
		raw := confd.RawData{}
		raw["service"] = "8.8.8.8"

		service, err := createService(raw)
		require.NoError(t, err)
		require.Nil(t, service)
	})

	t.Run("should return nil when value for service is missing", func(t *testing.T) {
		raw := confd.RawData{}
		raw["name"] = "heartOfGold"

		service, err := createService(raw)
		require.NoError(t, err)
		require.Nil(t, service)
	})

	t.Run("should return nil when name is not a string", func(t *testing.T) {
		raw["name"] = false
		raw["service"] = "8.8.8.8"

		service, err := createService(raw)
		require.NoError(t, err)
		require.Nil(t, service)
	})

	t.Run("should return nil when value for service is not a string", func(t *testing.T) {
		raw["name"] = "heartOfGold"
		raw["service"] = false

		service, err := createService(raw)
		require.NoError(t, err)
		require.Nil(t, service)
	})

	t.Run("should return created service even without attributes", func(t *testing.T) {
		raw := confd.RawData{}
		raw["name"] = "heartOfGold"
		raw["service"] = "8.8.8.8"

		service, err := createService(raw)
		require.NoError(t, err)
		assert.Equal(t, "heartOfGold", service.Location)
	})

	t.Run("should return created service even when attributes is not a map", func(t *testing.T) {
		raw["name"] = "heartOfGold"
		raw["service"] = "8.8.8.8"
		raw["attributes"] = "location:heartOfGoldLocation"

		service, err := createService(raw)
		require.NoError(t, err)
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

		service, err := createService(raw)
		require.NoError(t, err)
		assert.Equal(t, "heartOfGold", service.Name)
		assert.Equal(t, "http://8.8.8.8", service.URL)
		assert.Equal(t, "heartOfGold", service.Location)
	})

	t.Run("should return created service with rewrite rul", func(t *testing.T) {
		raw["name"] = "heartOfGold"
		raw["service"] = "8.8.8.8"
		attributes := map[string]interface{}{
			"rewrite": "{\"pattern\": \"elasticsearch\", \"rewrite\": \"test\"}",
		}
		raw["attributes"] = attributes

		service, err := createService(raw)
		require.NoError(t, err)
		assert.Equal(t, "elasticsearch", service.Rewrite.Pattern)
		assert.Equal(t, "test", service.Rewrite.Rewrite)
	})

	t.Run("should return error with invalid rewrite rul", func(t *testing.T) {
		raw["name"] = "heartOfGold"
		raw["service"] = "8.8.8.8"
		attributes := map[string]interface{}{
			"rewrite": "{\"paer\":: \"elasticsearch\"}",
		}
		raw["attributes"] = attributes

		_, err := createService(raw)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal rewrite rule")
	})
}
