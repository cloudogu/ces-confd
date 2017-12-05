package service_test

import (
	"testing"
  "github.com/cloudogu/ces-confd/confd/service"
  "github.com/docker/docker/pkg/testutil/assert"
)

func TestConvertToService(t *testing.T) {
  serviceRaw := "{\"name\": \"exampleService\", \"service\": \"127.0.0.1:3000\", \"port\": 3000, \"tags\": [\"webapp\"]}"
  service, err := service.ConvertToService("webapp", serviceRaw)
  if err != nil {
    t.Errorf("Failed to convert to service", err)
  }
  assert.Equal(t, service.Name,"exampleService")
  assert.Equal(t, service.URL,"http://127.0.0.1:3000")
}
