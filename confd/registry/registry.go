package registry

import "github.com/coreos/etcd/client"

// Config represents the configuration of a Registry
type Config struct {
	Endpoints []string
}

// Event represents a watchable event
type Event struct {
	Action string
	Value  string
}

// Registry manages a config registry (e.g. etcd)
type Registry interface {
	//NewRegistry(config Config) (*Registry, error)
	Get(key string) (*client.Response, error)
	Watch(key string, eventChannel chan Event)
}
