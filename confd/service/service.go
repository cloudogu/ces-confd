package service

import (
	"log"

	"github.com/cloudogu/ces-confd/confd"
	configRegistry "github.com/cloudogu/ces-confd/confd/registry"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
)

var modificationActions = []string{"create", "delete", "update", "set"}

// Services is a collection of service structs
type Services []*Service

// Service is a running service
type Service struct {
	Name string
	URL  string
	// State represents the current state of the service
	State string
	// StateNode is the name of the node which holds the state information of the service.
	// For dogus it's usually the name of the dogu.
	StateNode string
}

// String returns a string representation of a service
func (service *Service) String() string {
	return "{name=" + service.Name + ", URL=" + service.URL + "}"
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
	PreCommand      string             `yaml:"pre-command"`
	PostCommand     string             `yaml:"post-command"`
	IgnoreState     bool               `yaml:"ignore-state"`
	State           stateConfiguration `yaml:"state"`
	Order           confd.Order
}

type stateConfiguration struct {
	Source    string `yaml:"source"`
	SaneValue string `yaml:"sane-value"`
}

func createService(raw confd.RawData) *Service {
	service := raw.GetStringValue("service")
	if service == "" {
		return nil
	}

	name := raw.GetStringValue("name")
	if name == "" {
		return nil
	}

	return &Service{
		Name: name,
		URL:  "http://" + service,
	}
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
	stateChannel := make(chan *client.Response)
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

	if !conf.IgnoreState {
		log.Println("starting state watcher")
		go func() {
			for {
				// TODO: necessary?
				loader.ReloadServices()
				registry.Watch(conf.State.Source, true, stateChannel)
			}
		}()
	}
	for {
		select {
		case <-maintenanceChannel:
			loader.ReloadServices()
		case resp := <-serviceChannel:
			reloadServicesIfNecessary(loader, resp)
		case resp := <-stateChannel:
			// TODO: Here is room for improvement. We could also check if actions changed a state of a service
			if isModificationAction(resp.Action) {
				loader.ReloadServices()
			}
		}
	}
}
