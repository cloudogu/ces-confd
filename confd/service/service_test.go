package service

import (
  "github.com/stretchr/testify/require"
  "testing"
  "github.com/coreos/etcd/client"
  "github.com/cloudogu/ces-confd/confd"
  "github.com/stretchr/testify/assert"
  "fmt"
)

func TestServicesString(t *testing.T) {
  services := Services{}

  heartOfGold := &Service{Name: "heartOfGold", URL: "http://8.8.8.8"}
  services = append(services, heartOfGold)
  content := fmt.Sprintf("services: %v", services)
  assert.Equal(t, "services: [{name=heartOfGold, URL=http://8.8.8.8}]", content)
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

func TestConvertToService(t *testing.T) {
  service, err := convertToService("webapp", "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}")
  require.Nil(t, err)
  require.NotNil(t, service)
}

func TestConvertToServiceWithoutTag(t *testing.T) {
  service, err := convertToService("", "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\"}")
  require.Nil(t, err)
  require.NotNil(t, service)
}

func TestConvertToServiceWithoutTags(t *testing.T) {
  service, err := convertToService("webapp", "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\"}")
  require.Nil(t, err)
  require.Nil(t, service)
}

func TestConvertToServiceWithOtherTag(t *testing.T) {
  service, err := convertToService("webapp", "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"web\"]}")
  require.Nil(t, err)
  require.Nil(t, service)
}

func TestConvertToServiceWithNonArrayTags(t *testing.T) {
  _, err := convertToService("webapp", "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": 12}")
  require.NotNil(t, err)
}

func TestHasServiceChanged(t *testing.T) {
  conf := Configuration{
    Tag: "webapp",
  }

  node := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
  }

  response := client.Response{
    Action:"create",
    Node: &node,
  }

  changed, err := hasServiceChanged(conf, &response)
  require.Nil(t, err)
  require.True(t, changed)
}

func TestHasServiceChangedIgnoreDirectories(t *testing.T) {
  conf := Configuration{
    Tag: "webapp",
  }

  node := client.Node{
    Dir: true,
  }

  response := client.Response{
    Action:"create",
    Node: &node,
  }

  changed, err := hasServiceChanged(conf, &response)
  require.Nil(t, err)
  require.False(t, changed)
}

func TestHasServiceChangedDeleteAction(t *testing.T) {
  conf := Configuration{
    Tag: "webapp",
  }

  node := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
  }

  response := client.Response{
    Action: "delete",
    Node: &node,
  }

  changed, err := hasServiceChanged(conf, &response)
  require.Nil(t, err)
  require.True(t, changed)
}

func TestHasServiceChangedUpdateAction(t *testing.T) {
  conf := Configuration{
    Tag: "webapp",
  }

  node := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
  }

  prevNode := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.4.4\", \"tags\": [\"webapp\"]}",
  }

  response := client.Response{
    Action: "update",
    Node: &node,
    PrevNode: &prevNode,
  }

  changed, err := hasServiceChanged(conf, &response)
  require.Nil(t, err)
  require.True(t, changed)
}


func TestHasServiceChangedSetAction(t *testing.T) {
  conf := Configuration{
    Tag: "webapp",
  }

  node := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
  }

  prevNode := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.4.4\", \"tags\": [\"webapp\"]}",
  }

  response := client.Response{
    Action: "set",
    Node: &node,
    PrevNode: &prevNode,
  }

  changed, err := hasServiceChanged(conf, &response)
  require.Nil(t, err)
  require.True(t, changed)
}

func TestHasServiceChangedSetPreviousNonService(t *testing.T) {
  conf := Configuration{
    Tag: "webapp",
  }

  node := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
  }

  prevNode := client.Node{
    Dir: false,
    Value: "{}",
  }

  response := client.Response{
    Action: "set",
    Node: &node,
    PrevNode: &prevNode,
  }

  changed, err := hasServiceChanged(conf, &response)
  require.Nil(t, err)
  require.True(t, changed)
}


func TestHasServiceChangedSetPreviousServiceToNonService(t *testing.T) {
  conf := Configuration{
    Tag: "webapp",
  }

  node := client.Node{
    Dir: false,
    Value: "{}",
  }

  prevNode := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
  }

  response := client.Response{
    Action: "set",
    Node: &node,
    PrevNode: &prevNode,
  }

  changed, err := hasServiceChanged(conf, &response)
  require.Nil(t, err)
  require.True(t, changed)
}

func TestHasServiceChangedSetNodeErrorButPreviousNodeIsFine(t *testing.T) {
  conf := Configuration{
    Tag: "webapp",
  }

  node := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": 42}",
  }

  prevNode := client.Node{
    Dir: false,
    Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
  }

  response := client.Response{
    Action: "set",
    Node: &node,
    PrevNode: &prevNode,
  }

  _, err := hasServiceChanged(conf, &response)
  require.NotNil(t, err)
}
