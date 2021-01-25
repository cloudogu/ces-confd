package service

import (
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertTotServicesShouldNotFailOnError(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	heartOfGold := &client.Node{
		Key:   "/services/heartOfGold",
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"], \"attributes\":{\"location\":\"heartOfGoldLocation\"}}",
	}

	restaurantAtTheEndOfTheUniverse := &client.Node{
		Key:   "/services/restaurantAtTheEndOfTheUniverse",
		Value: "{\"name\": \"restaurantAtTheEndOfTheUniverse\", \"service\": \"8.8.4.4\", \"tags\": [\"webapp\"]}",
	}

	invalidService := &client.Node{
		Key:   "/services/invalid",
		Value: "{\"name\": \"invalid\", \"service\": \"8.8.4.4\", \"tags\": 42}",
	}

	childNodes := client.Nodes{heartOfGold, restaurantAtTheEndOfTheUniverse, invalidService}
	services := loader.convertChildNodesToServices(childNodes)
	assert.Equal(t, 2, len(services))
}

func TestConvertToService(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	service, err := loader.convertToService("{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"],\"healthStatus\":\"healthy\",\"attributes\":{\"day\":\"Friday\",\"location\":\"heartOfGoldLocation\"}}")
	require.Nil(t, err)
	require.NotNil(t, service)
	require.Equal(t, "healthy", service.HealthStatus)
}

func TestConvertToServiceWithEmptyHealth(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	service, err := loader.convertToService("{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}")
	require.Nil(t, err)
	require.NotNil(t, service)
	require.Equal(t, "", service.HealthStatus)
}

func TestConvertToServiceWithIgnoredHealth(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp", IgnoreHealth: true}}
	service, err := loader.convertToService("{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"],\"healthStatus\":\"healthy\"}")
	require.Nil(t, err)
	require.NotNil(t, service)
	require.Equal(t, "", service.HealthStatus)
}

func TestConvertToServiceWithoutTag(t *testing.T) {
	loader := &Loader{config: Configuration{}}
	service, err := loader.convertToService("{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\"}")
	require.Nil(t, err)
	require.NotNil(t, service)
}

func TestConvertToServiceWithoutTags(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	service, err := loader.convertToService("{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\"}")
	require.Nil(t, err)
	require.Nil(t, service)
}

func TestConvertToServiceWithOtherTag(t *testing.T) {

	loader := &Loader{config: Configuration{Tag: "webapp"}}
	service, err := loader.convertToService("{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"web\"]}")
	require.Nil(t, err)
	require.Nil(t, service)
}

func TestConvertToServiceWithNonArrayTags(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}

	_, err := loader.convertToService("{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": 12}")
	require.NotNil(t, err)
}

func TestHasServiceChanged(t *testing.T) {

	loader := &Loader{config: Configuration{Tag: "webapp"}}

	node := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
	}

	response := client.Response{
		Action: "create",
		Node:   &node,
	}

	changed, err := loader.HasServiceChanged(&response)
	require.Nil(t, err)
	require.True(t, changed)
}

func TestHasServiceChangedIgnoreDirectories(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}

	node := client.Node{
		Dir: true,
	}

	response := client.Response{
		Action: "create",
		Node:   &node,
	}

	changed, err := loader.HasServiceChanged(&response)
	require.Nil(t, err)
	require.False(t, changed)
}

func TestHasServiceChangedDeleteAction(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}

	node := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
	}

	response := client.Response{
		Action: "delete",
		Node:   &node,
	}

	changed, err := loader.HasServiceChanged(&response)
	require.Nil(t, err)
	require.True(t, changed)
}

func TestHasServiceChangedUpdateAction(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	node := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
	}

	prevNode := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.4.4\", \"tags\": [\"webapp\"]}",
	}

	response := client.Response{
		Action:   "update",
		Node:     &node,
		PrevNode: &prevNode,
	}

	changed, err := loader.HasServiceChanged(&response)
	require.Nil(t, err)
	require.True(t, changed)
}

func TestHasServiceChangedSetAction(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	node := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
	}

	prevNode := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.4.4\", \"tags\": [\"webapp\"]}",
	}

	response := client.Response{
		Action:   "set",
		Node:     &node,
		PrevNode: &prevNode,
	}

	changed, err := loader.HasServiceChanged(&response)
	require.Nil(t, err)
	require.True(t, changed)
}

func TestHasServiceChangedSetPreviousNonService(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}

	node := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
	}

	prevNode := client.Node{
		Dir:   false,
		Value: "{}",
	}

	response := client.Response{
		Action:   "set",
		Node:     &node,
		PrevNode: &prevNode,
	}

	changed, err := loader.HasServiceChanged(&response)
	require.Nil(t, err)
	require.True(t, changed)
}

func TestHasServiceChangedSetPreviousServiceToNonService(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	node := client.Node{
		Dir:   false,
		Value: "{}",
	}

	prevNode := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
	}

	response := client.Response{
		Action:   "set",
		Node:     &node,
		PrevNode: &prevNode,
	}

	changed, err := loader.HasServiceChanged(&response)
	require.Nil(t, err)
	require.True(t, changed)
}

func TestHasServiceChangedSetNodeErrorButPreviousNodeIsFine(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}

	node := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": 42}",
	}

	prevNode := client.Node{
		Dir:   false,
		Value: "{\"name\": \"heartOfGold\", \"service\": \"8.8.8.8\", \"tags\": [\"webapp\"]}",
	}

	response := client.Response{
		Action:   "set",
		Node:     &node,
		PrevNode: &prevNode,
	}

	_, err := loader.HasServiceChanged(&response)
	require.NotNil(t, err)
}

func TestIsServiceNodeWithoutNode(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	isService, err := loader.isServiceNode(nil)
	require.Nil(t, err)
	require.False(t, isService)
}

func TestIsServiceNodeWithoutValue(t *testing.T) {
	loader := &Loader{config: Configuration{Tag: "webapp"}}
	node := &client.Node{
		Dir: false,
	}

	isService, err := loader.isServiceNode(node)
	require.Nil(t, err)
	require.False(t, isService)
}
