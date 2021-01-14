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

	heartOfGold := &Service{Name: "heartOfGold", URL: "http://8.8.8.8", HealthStatus: "healthy"}
	services = append(services, heartOfGold)
	content := fmt.Sprintf("services: %v", services)
	assert.Equal(t, "services: [{name=heartOfGold, URL=http://8.8.8.8, HealthStatus=healthy}]", content)
}

func TestCreateService(t *testing.T) {
	raw := confd.RawData{}
	raw["name"] = "heartOfGold"
	raw["service"] = "8.8.8.8"

	service := createService(raw)
	assert.Equal(t, "heartOfGold", service.Name)
	assert.Equal(t, "http://8.8.8.8", service.URL)
}

func TestCreateServiceWithoutNameKey(t *testing.T) {
	raw := confd.RawData{}
	raw["service"] = "8.8.8.8"

	service := createService(raw)
	require.Nil(t, service)
}

func TestCreateServiceWithoutServiceKey(t *testing.T) {
	raw := confd.RawData{}
	raw["name"] = "heartOfGold"

	service := createService(raw)
	require.Nil(t, service)
}

func TestCreateServiceWithNonStringServiceKey(t *testing.T) {
	raw := confd.RawData{}
	raw["name"] = "heartOfGold"
	raw["service"] = false

	service := createService(raw)
	require.Nil(t, service)
}

func TestCreateServiceWithNonStringNameKey(t *testing.T) {
	raw := confd.RawData{}
	raw["name"] = false
	raw["service"] = "8.8.8.8"

	service := createService(raw)
	require.Nil(t, service)
}
