package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudogu/ces-confd/confd"
	configRegistry "github.com/cloudogu/ces-confd/confd/registry"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
)

var modificationActions = []string{"create", "delete", "update", "set"}

// Services is a collection of service structs
type Services []*Service

// Rewrite is a rewrite rule for a service.
type Rewrite struct {
	Pattern string `json:"pattern"`
	Rewrite string `json:"rewrite"`
}

// Service is a running service
type Service struct {
	Name         string   `json:"name"`
	URL          string   `json:"url"`
	HealthStatus string   `json:"healthStatus"`
	Location     string   `json:"location"`
	Rewrite      *Rewrite `json:"rewrite,omitempty"`
}

// String returns a string representation of a service
func (service *Service) String() string {
	return fmt.Sprintf("{name=%s, URL=%s, HealthStatus=%s, Location=%s, Rewrite=%+v}", service.Name, service.URL, service.HealthStatus, service.Location, service.Rewrite)
}

// Source of services path in etcd
type Source struct {
	Path string
}

// Configuration struct for the service part of ces-confd
type Configuration struct {
	Source          Source
	MaintenanceMode string `yaml:"maintenance-mode"`
	Target          string
	Template        string
	Tag             string
	PreCommand      string `yaml:"pre-command"`
	PostCommand     string `yaml:"post-command"`
	Order           confd.Order
	IgnoreHealth    bool `yaml:"ignore-health"`
}

func createService(raw confd.RawData) (*Service, error) {
	service := raw.GetStringValue("service")
	if service == "" {
		return nil, nil
	}

	name := raw.GetStringValue("name")
	if name == "" {
		return nil, nil
	}

	// an empty healthStatus is ok since maybe an old version of registrator is used
	healthStatus := raw.GetStringValue("healthStatus")

	location := raw.GetAttributeValue("location")
	if location == "" {
		location = name
	}

	rewriteRule := raw.GetAttributeValue("rewrite")
	rule := &Rewrite{}
	if rewriteRule != "" {
		err := json.Unmarshal([]byte(rewriteRule), rule)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rewrite rule: %w", err)
		}
	}

	return &Service{
		Name:         name,
		URL:          "http://" + service,
		HealthStatus: healthStatus,
		Location:     location,
		Rewrite:      rule,
	}, nil
}

func hasTag(raw confd.RawData, tag string) (bool, error) {
	tagsInterface, ok := raw["tags"]
	if !ok {
		return false, nil
	}

	tags, ok := tagsInterface.([]interface{})
	if !ok {
		return false, errors.New("tags must be an slice of strings")
	}

	return confd.Contains(tags, tag), nil
}

func isDirectory(node *client.Node) bool {
	return node.Dir
}

func isModificationAction(action string) bool {
	return confd.ContainsString(modificationActions, action)
}

func reloadServicesIfNecessary(loader *Loader, resp *client.Response) {
	key := resp.Node.Key
	changed, err := loader.HasServiceChanged(resp)
	if err != nil {
		log.Printf("failed to check if the change is responsible for a service: %v", err)
		loader.ReloadServices()
		return
	}

	action := resp.Action

	if changed {
		log.Printf("service %s changed, action=%s", key, action)
		loader.ReloadServices()
	} else {
		log.Printf("ignoring change to non service key %s with action %s", key, action)
	}
}

// Run creates the configuration for the services and updates the configuration whenever a service changed
func Run(conf Configuration, registry configRegistry.Registry) {
	serviceChannel := make(chan *client.Response)
	maintenanceChannel := make(chan *client.Response)
	loader := &Loader{
		registry: registry,
		config:   conf,
		writer:   &CommandWriter{config: conf},
	}

	log.Println("starting service watcher")
	go func() {
		for {
			loader.ReloadServices()
			registry.Watch(conf.Source.Path, true, serviceChannel)
		}
	}()
	log.Println("starting maintenance mode watcher")

	go func() {
		for {
			// TODO: necessary?
			loader.ReloadServices()
			registry.Watch(conf.MaintenanceMode, false, maintenanceChannel)
		}
	}()
	for {
		select {
		case <-maintenanceChannel:
			loader.ReloadServices()
		case resp := <-serviceChannel:
			reloadServicesIfNecessary(loader, resp)
		}
	}
}
