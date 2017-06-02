package service

import (
	"encoding/json"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
  "github.com/pkg/errors"
  "path"
  "os"
  "html/template"
  . "github.com/cloudogu/ces-confd/confd"
  "log"
)

// Services is a collection of service structs
type Services []*Service

// Service is a running service
type Service struct {
	Name string
	URL  string
}

type Source struct {
  Path string
}

type Configuration struct {
  Source   Source
  Target   string
  Template string
  Tag      string
  PreCommand  string `yaml:"pre-command"`
  PostCommand string `yaml:"post-command"`
  Order Order
}

func createService(raw RawData) *Service {
	return &Service{
		Name: raw["name"].(string),
		URL:  "http://" + raw["service"].(string),
	}
}

func convertToServices(kapi client.KeysAPI, tag string, key string) (Services, error) {
  resp, err := kapi.Get(context.Background(), key, nil)
  if err != nil {
    return nil, errors.Wrapf(err, "failed to read service key %s from etcd", key)
  }

  services := Services{}
  for _, child := range resp.Node.Nodes {
    service, err := convertToService(tag, child.Value)
    if err != nil {
      return nil, errors.Wrap(err, "failed to convert node to service")
    } else if service != nil {
      services = append(services, service)
    }
  }

  return services, nil
}

func convertToService(tag string, value string) (*Service, error) {
	raw := RawData{}
	err := json.Unmarshal([]byte(value), &raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall service json")
	}

	if tag != "" {
		if raw["tags"] != nil {
			tags := raw["tags"].([]interface{})
			if Contains(tags, tag) {
				return createService(raw), err
			}
		} else {
			return createService(raw), err
		}
	}
	return nil, nil
}

// serviceReader reads from etcd and converts the keys and value to service
// struct, which can easily used for configuration templates
func serviceReader(source Source, tag string, kapi client.KeysAPI) (interface{}, error) {
	resp, err := kapi.Get(context.Background(), source.Path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read root %s from etcd", source.Path)
	}

	services := Services{}
	for _, child := range resp.Node.Nodes {
		serviceEntries, err := convertToServices(kapi, tag, child.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert node to service")
		}
		for _, service := range serviceEntries {
			services = append(services, service)
		}
	}

	return services, nil
}

func readFromConfig(configuration Configuration, kapi client.KeysAPI) (interface{}, error) {
  return serviceReader(configuration.Source, configuration.Tag, kapi)
}

// templateWriter transform the data with a golang template
func templateWriter(conf Configuration, data interface{}) error {
  name := path.Base(conf.Template)
  tmpl, err := template.New(name).ParseFiles(conf.Template)
  if err != nil {
    return errors.Wrap(err, "failed to parse template")
  }

  file, err := os.Create(conf.Target)
  if err != nil {
    return errors.Wrapf(err, "failed to create target file %s", conf.Target)
  }
  defer file.Close()

  err = tmpl.Execute(file, data)
  if err != nil {
    return errors.Wrap(err, "failed to render template")
  }
  return nil
}

func Run(conf Configuration, kapi client.KeysAPI) {
  println("in service Run()")
  services, err := readFromConfig(conf, kapi)
  if err != nil {
    log.Println("error durring read", err)
  }
  log.Printf("all found services: %i", services)
  templateWriter(conf, services)
}


