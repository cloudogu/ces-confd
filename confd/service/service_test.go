package service_test

import (
	"testing"
  "github.com/cloudogu/ces-confd/confd/service"
  "github.com/docker/docker/pkg/testutil/assert"
  assert2 "github.com/stretchr/testify/assert"
)

func TestConvertToService(t *testing.T) {
  serviceRaw := "{\"name\": \"exampleService\", \"service\": \"127.0.0.1:3000\", \"port\": 3000, \"tags\": [\"foo\", \"webapp\"]}"
  service, err := service.ConvertToService("webapp", serviceRaw)
  if err != nil {
    t.Errorf("Failed to convert to service", err)
  }
  assert.Equal(t, service.Name,"exampleService")
  assert.Equal(t, service.URL,"http://127.0.0.1:3000")
}

func TestConvertToServiceWithoutNeededTag(t *testing.T) {
  serviceRaw := "{\"name\": \"exampleService\", \"service\": \"127.0.0.1:3000\", \"port\": 3000, \"tags\": [\"foo\"]}"
  service, err := service.ConvertToService("webapp", serviceRaw)
  if err != nil {
    t.Errorf("Failed to convert to service", err)
  }
  assert2.Nil(t, service)
}

func TestConvertToServiceWithoutTags(t *testing.T) {
  serviceRaw := "{\"name\": \"exampleService\", \"service\": \"127.0.0.1:3000\", \"port\": 3000, \"tags\": []}"
  service, err := service.ConvertToService("webapp", serviceRaw)
  if err != nil {
    t.Errorf("Failed to convert to service", err)
  }
  assert2.Nil(t, service)
}

func TestConvertToServiceWithMultiplePorts(t *testing.T) {
  service0Raw := "{\"name\": \"exampleService-3000\", \"service\": \"127.0.0.1:3000\", \"port\": 3000, \"tags\": [\"webapp:port=3000\"]}"
  service1Raw := "{\"name\": \"exampleService-3001\", \"service\": \"127.0.0.1:3001\", \"port\": 3001, \"tags\": [\"webapp:port=3000\"]}"
  service0, err := service.ConvertToService("webapp", service0Raw)
  if err != nil {
    t.Errorf("Failed to convert to service.", err)
  }
  assert.NotNil(t, service0)
  assert.Equal(t, service0.Name,"exampleService")
  assert.Equal(t, service0.URL,"http://127.0.0.1:3000")

  service1, err := service.ConvertToService("webapp", service1Raw)
  if err != nil {
    t.Errorf("Failed to convert to service.", err)
  }
  assert2.Nil(t, service1)
}
