package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cloudogu/ces-confd/confd"
	configRegistry "github.com/cloudogu/ces-confd/confd/registry"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
)

type Loader struct {
	registry configRegistry.Registry
	config   Configuration
	writer   Writer
}

func (l *Loader) ReloadServices() {
	log.Println("reload services from etcd")
	templateModel, err := l.createTemplateModel()
	if err != nil {
		log.Printf("failed to reload services: %v", err)
		return
	}

	log.Printf("write services to template: %v", templateModel)

	if err := l.writer.WriteTemplate(templateModel); err != nil {
		log.Printf("error on writeTemplate: %s", err.Error())
	}
}

func (l *Loader) HasServiceChanged(resp *client.Response) (bool, error) {
	if !isDirectory(resp.Node) && isModificationAction(resp.Action) {
		return l.isServiceResponse(resp)
	}
	return false, nil
}

func (l *Loader) convertToServices(key string) (Services, error) {
	resp, err := l.registry.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read service key %s from etcd", key)
	}

	services := l.convertChildNodesToServices(resp.Node.Nodes)

	if !l.config.IgnoreState {
		l.getStates(services)
	}

	return services, nil
}

func (l *Loader) getStates(services Services) {
	for _, service := range services {
		res, err := l.registry.Get(fmt.Sprintf("%s/%s", l.config.State.Source, service.StateNode))
		if err != nil {
			log.Printf("could not get state of %s:%s", service.StateNode, err)
			log.Printf("continue and set state state of %s to %s", service.Name, "not ready")
			service.State = "not ready"
			continue
		}
		service.State = res.Node.Value
	}
}

func (l *Loader) convertChildNodesToServices(childNodes client.Nodes) Services {
	services := Services{}
	for _, child := range childNodes {
		service, err := l.convertToService(child.Value)
		if err != nil {
			// do not fail, if a single service contains an invalid entry
			log.Printf("failed to convert node %s to service: %v", child.Key, err)
		} else if service != nil {
			if service.StateNode == "" && !l.config.IgnoreState {
				// For now the stateNode is derived from the service key, but it can also be set in registry.
				stateNode, err := l.getStateNodeFromKey(child.Key)
				if err != nil {
					log.Printf("failed to get stateNode from key %s with error: %v", child.Key, err)
					log.Printf("skip service %s", service.Name)
					continue
				}
				service.StateNode = stateNode
			}
			services = append(services, service)
		}
	}

	return services
}

func (l *Loader) getStateNodeFromKey(key string) (string, error) {
	expectedPrefix := l.config.Source.Path + "/"
	if !strings.HasPrefix(key, expectedPrefix) {
		return "", fmt.Errorf("key %s does not match expected format since it does not begin with %s", key, expectedPrefix)
	}
	serviceSuffix := strings.TrimPrefix(key, expectedPrefix)
	serviceNode := strings.Split(serviceSuffix, "/")
	stateNode := strings.Split(serviceNode[0], "-")
	return stateNode[0], nil
}

func (l *Loader) convertToService(value string) (*Service, error) {
	raw := confd.RawData{}
	err := json.Unmarshal([]byte(value), &raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall service json")
	}

	if l.config.Tag != "" {
		exists, err := hasTag(raw, l.config.Tag)
		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, nil
		}
	}
	return createService(raw), nil
}

func (l *Loader) createTemplateModel() (TemplateModel, error) {
	maintenanceMode := ""
	resp, err := l.registry.Get(l.config.MaintenanceMode)

	if err != nil {
		if !client.IsKeyNotFound(err) {
			return TemplateModel{}, errors.Wrapf(err, "could not determine state of maintenance mode")
		}
	} else {
		log.Printf("Maintenance mode resp: %v", resp)
		maintenanceMode = resp.Node.Value
	}

	services, err := l.serviceReader()
	if err != nil {
		return TemplateModel{}, errors.Wrapf(err, "Could not read service %s", l.config.Source.Path)
	}

	return TemplateModel{maintenanceMode, services}, nil
}

// serviceReader reads from etcd and converts the keys and value to service
// struct, which can easily used for configuration templates
func (l *Loader) serviceReader() (Services, error) {
	resp, err := l.registry.Get(l.config.Source.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root %s from etcd", l.config.Source.Path)
	}

	services := Services{}
	for _, child := range resp.Node.Nodes {
		// convertToServices returns only an error, if the root key could not be read.
		// In this case we should return an too.
		serviceEntries, err := l.convertToServices(child.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert node %s to service", child.Key)
		}
		for _, service := range serviceEntries {
			services = append(services, service)
		}
	}
	return services, nil
}

func (l *Loader) isServiceResponse(resp *client.Response) (bool, error) {
	service, err := l.isServiceNode(resp.Node)
	if err != nil {
		return false, err
	}

	if service {
		return true, nil
	}

	service, err = l.isServiceNode(resp.PrevNode)
	if err != nil {
		return false, err
	}

	if service {
		return true, nil
	}

	return false, nil
}

func (l *Loader) isServiceNode(node *client.Node) (bool, error) {
	if node == nil {
		return false, nil
	}

	if node.Value == "" {
		return false, nil
	}

	service, err := l.convertToService(node.Value)
	if err != nil {
		return false, errors.Wrap(err, "failed to convert node to service")
	}

	return service != nil, nil
}
