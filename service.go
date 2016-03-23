package main

import (
	"encoding/json"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

// Services is a collection of service structs
type Services []*Service

// Service is a running service
type Service struct {
	Name string
	URL  string
}

func contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func convertToService(value string) (*Service, error) {
	raw := RawData{}
	err := json.Unmarshal([]byte(value), &raw)
	if err != nil {
		return nil, err
	}

	tags := raw["tags"].([]interface{})
	if contains(tags, "webapp") {
		service := Service{
			Name: raw["name"].(string),
			URL:  "http://" + raw["service"].(string),
		}
		return &service, err
	}
	return nil, nil
}

func convertToServices(kapi client.KeysAPI, key string) (Services, error) {
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		return nil, err
	}

	services := Services{}
	for _, child := range resp.Node.Nodes {
		service, err := convertToService(child.Value)
		if err != nil {
			return nil, err
		} else if service != nil {
			services = append(services, service)
		}
	}

	return services, nil
}

// ServiceReader reads from etcd and converts the keys and value to service
// struct, which can easily used for configuration templates
func ServiceReader(kapi client.KeysAPI, root string) (interface{}, error) {
	resp, err := kapi.Get(context.Background(), root, nil)
	if err != nil {
		return nil, err
	}

	services := Services{}
	for _, child := range resp.Node.Nodes {
		serviceEntries, err := convertToServices(kapi, child.Key)
		if err != nil {
			return nil, err
		}
		for _, service := range serviceEntries {
			services = append(services, service)
		}
	}

	return services, nil
}
