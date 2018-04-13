package service

import (
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path"

	"github.com/cloudogu/ces-confd/confd"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var modificationActions = []string{"create", "delete", "update", "set"}

// Services is a collection of service structs
type Services []*Service

// Service is a running service
type Service struct {
	Name string
	URL  string
}

// TemplateModel is the input for the target template
type TemplateModel struct {
	Maintenance string
	Services    Services
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
	PreCommand      string `yaml:"pre-command"`
	PostCommand     string `yaml:"post-command"`
	Order           confd.Order
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

func convertToServices(kapi client.KeysAPI, tag string, key string) (Services, error) {
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read service key %s from etcd", key)
	}

	return convertChildNodesToServices(resp.Node.Nodes, tag), nil
}

func convertChildNodesToServices(childNodes client.Nodes, tag string) Services {
	services := Services{}
	for _, child := range childNodes {
		service, err := convertToService(tag, child.Value)
		if err != nil {
			// do not fail, if a single service contains an invalid entry
			log.Printf("failed to convert node %s to service: %v", child.Key, err)
		} else if service != nil {
			services = append(services, service)
		}
	}

	return services
}

func convertToService(tag string, value string) (*Service, error) {
	raw := confd.RawData{}
	err := json.Unmarshal([]byte(value), &raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall service json")
	}

	if tag != "" {
		exists, err := hasTag(raw, tag)
		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, nil
		}
	}
	return createService(raw), nil
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

func createTemplateModel(configuration Configuration, kapi client.KeysAPI) (*TemplateModel, error) {
	maintenanceMode := ""
	resp, err := kapi.Get(context.Background(), configuration.MaintenanceMode, nil)
	if err != nil {
		if !client.IsKeyNotFound(err) {
			return nil, errors.Wrapf(err, "could not determine state of maintenance mode")
		}
	} else {
		maintenanceMode = resp.Node.Value
	}

	services, err := serviceReader(configuration.Source, configuration.Tag, kapi)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read service %s", configuration.Source.Path)
	}

	return &TemplateModel{maintenanceMode, services}, nil
}

// serviceReader reads from etcd and converts the keys and value to service
// struct, which can easily used for configuration templates
func serviceReader(source Source, tag string, kapi client.KeysAPI) (Services, error) {
	resp, err := kapi.Get(context.Background(), source.Path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root %s from etcd", source.Path)
	}

	services := Services{}
	for _, child := range resp.Node.Nodes {
		// convertToServices returns only an error, if the root key could not be read.
		// In this case we should return an too.
		serviceEntries, err := convertToServices(kapi, tag, child.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert node %s to service", child.Key)
		}
		for _, service := range serviceEntries {
			services = append(services, service)
		}
	}
	return services, nil
}

func readFromConfig(configuration Configuration, kapi client.KeysAPI) (interface{}, error) {
	return createTemplateModel(configuration, kapi)
}

// templateWriter transform the data with a golang template
func templateWriter(conf Configuration, data interface{}) error {
	if conf.PreCommand != "" {
		err := preCheck(conf, data)
		if err != nil {
			return errors.Wrap(err, "pre check failed")
		}
	}

	err := write(conf, data)
	if err != nil {
		return errors.Wrap(err, "failed to write data")
	}

	if conf.PostCommand != "" {
		err = post(conf.PostCommand)
		if err != nil {
			return errors.Wrap(err, "post command failed")
		}
	}
	return nil
}

func write(conf Configuration, data interface{}) error {
	name := path.Base(conf.Template)
	tmpl, err := template.New(name).ParseFiles(conf.Template)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	file, err := os.Create(conf.Target)
	if err != nil {
		return errors.Wrapf(err, "failed to create target file %s", conf.Target)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("failed to close file")
		}
	}()

	err = tmpl.Execute(file, data)
	if err != nil {
		return errors.Wrap(err, "failed to render template")
	}
	return nil
}

func reloadServices(conf Configuration, kapi client.KeysAPI) {
	log.Println("reload services from etcd")
	services, err := readFromConfig(conf, kapi)
	if err != nil {
		log.Printf("failed to reload services: %v", err)
	}

	log.Printf("write services to template: %v", services)

	if err := templateWriter(conf, services); err != nil {
		log.Printf("error on templateWriter: %s", err.Error())
	}
}

func watch(conf Configuration, kapi client.KeysAPI, etcdIndex uint64, execChannel chan string) {
	watcherOpts := client.WatcherOptions{AfterIndex: etcdIndex, Recursive: true}
	watcher := kapi.Watcher(conf.Source.Path, &watcherOpts)
	for {
		resp, err := watcher.Next(context.Background())
		if err != nil {
			// TODO: execute before watch start again? wait to reduce load, in case of unrecoverable error?
			// TODO: is recursive required? continue? do nothing?
			watch(conf, kapi, etcdIndex, execChannel)
		} else {
			key := resp.Node.Key

			changed, err := hasServiceChanged(conf, resp)
			if err != nil {
				log.Printf("failed to check if the change is responsible for a service: %v", err)
				execChannel <- key
				continue
			}

			action := resp.Action
			etcdIndex = resp.Index
			if changed {
				log.Printf("service %s changed, action=%s", key, action)
				execChannel <- key
			} else {
				log.Printf("ignoring change to non service key %s with action %s", key, action)
			}
		}
	}
}

func hasServiceChanged(conf Configuration, resp *client.Response) (bool, error) {
	if !isDirectory(resp.Node) && isModificationAction(resp.Action) {
		return isServiceResponse(resp, conf.Tag)
	}
	return false, nil
}

func isServiceResponse(resp *client.Response, tag string) (bool, error) {
	service, err := isServiceNode(resp.Node, tag)
	if err != nil {
		return false, err
	}

	if service {
		return true, nil
	}

	service, err = isServiceNode(resp.PrevNode, tag)
	if err != nil {
		return false, err
	}

	if service {
		return true, nil
	}

	return false, nil
}

func isDirectory(node *client.Node) bool {
	return node.Dir
}

func isModificationAction(action string) bool {
	return confd.ContainsString(modificationActions, action)
}

func isServiceNode(node *client.Node, tag string) (bool, error) {
	if node == nil {
		return false, nil
	}

	if node.Value == "" {
		return false, nil
	}

	service, err := convertToService(tag, node.Value)
	if err != nil {
		return false, errors.Wrapf(err, "failed to convert node to service")
	}

	return service != nil, nil
}

func watchForMaintenanceMode(conf Configuration, kapi client.KeysAPI, etcdIndex uint64, execChannel chan string) {
	watcherOpts := client.WatcherOptions{AfterIndex: etcdIndex, Recursive: false}
	watcher := kapi.Watcher(conf.MaintenanceMode, &watcherOpts)
	for {
		resp, err := watcher.Next(context.Background())
		if err != nil {
			// TODO: execute before watch start again? wait to reduce load, in case of unrecoverable error?
			// TODO: is recursive required? continue? do nothing?
			watchForMaintenanceMode(conf, kapi, etcdIndex, execChannel)
		} else {
			log.Println("Change in maintenance mode config:", resp.Node.Value)
			etcdIndex = resp.Index
			execChannel <- resp.Node.Key
		}
	}
}

// Run creates the configuration for the services and updates the configuration whenever a service changed
func Run(conf Configuration, kapi client.KeysAPI) {

	reloadServices(conf, kapi)

	execChannel := make(chan string)

	log.Println("starting service watcher")
	go watch(conf, kapi, 1, execChannel)

	log.Println("starting maintenance mode watcher")
	go watchForMaintenanceMode(conf, kapi, 1, execChannel)

	for key := range execChannel {
		log.Printf("reloading services, because key %s has changed", key)
		reloadServices(conf, kapi)
	}
}
