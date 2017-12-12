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
  "strings"
  "fmt"
  "regexp"
)

// Services is a collection of service structs
type Services []*Service

// Service is a running service
type Service struct {
	Name string
	URL  string
}

// Source of services path in etcd
type Source struct {
	Path string
}

// Configuration struct for the service part of ces-confd
type Configuration struct {
	Source      Source
	Target      string
	Template    string
	Tag         string
	PreCommand  string `yaml:"pre-command"`
	PostCommand string `yaml:"post-command"`
	Order       confd.Order
}

func convertToServices(kapi client.KeysAPI, tag string, key string) (Services, error) {
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read service key %s from etcd", key)
	}

	services := Services{}
	for _, child := range resp.Node.Nodes {
		service, err := ConvertToService(tag, child.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert node to service")
		} else if service != nil {
			services = append(services, service)
		}
	}

	return services, nil
}

// ConvertToService converts a json into a service if it contains the passed serviceTag
func ConvertToService(serviceTag string, value string) (*Service, error) {
	raw := confd.RawData{}
	err := json.Unmarshal([]byte(value), &raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall service json")
	}

	if serviceTag != "" {
		return createServiceIfNecessary(raw, serviceTag), nil
	} else {
    return createServiceFromRaw(raw), nil
  }
	return nil, nil
}

func createServiceIfNecessary(data map[string]interface{}, serviceTag string) (*Service) {
  if data["tags"] != nil {
    tags := data["tags"].([]interface{})
    for _, sliceItem := range tags {
      tag, castSuccessful := sliceItem.(string)
      if !castSuccessful {
        continue
      }
      service := checkForTag(tag, serviceTag, data)
      if service != nil {
        return service
      }
    }
  }
  return nil
}
func checkForTag(tag string, serviceTag string, data map[string]interface{} ) (*Service) {
  if tag == serviceTag {
    return createServiceFromRaw(data)
  }
  if strings.HasPrefix(tag, serviceTag) {
    port, err := findPortInTag(tag, serviceTag)
    if err != nil {
      log.Println(err)
      return nil
    }
    if port == fmt.Sprintf("%v",data["port"]) {
      name := strings.TrimSuffix(data["name"].(string), "-"+port)
      address := data["service"].(string)
      return createService(name, address)
    }
  }
  return nil
}

func findPortInTag(tag string, tagPrefix string) (string, error) {
  // searches for <port> in strings like 'webapp:port=<port>'
  re := regexp.MustCompile("^"+tagPrefix+".*:port=([0-9]+).*$")
  // returns array of first(and only) match of regex and the captured group
  match := re.FindStringSubmatch(tag)
  if len(match) <= 1 {
    return "", errors.New("Could not find port in tag "+ tag)
  }
  return match[1], nil
}

func createServiceFromRaw(raw confd.RawData) *Service {
  return createService(raw["name"].(string), raw["service"].(string))
}

func createService(name string, address string) *Service {
  return &Service{
    Name: name,
    URL:  "http://" + address,
  }
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

// naming
func execute(conf Configuration, kapi client.KeysAPI) {
	log.Println("read from etcd")
	services, err := readFromConfig(conf, kapi)
	if err != nil {
		log.Println("error durring read", err)
	}
	log.Printf("write services to template: %v", services)

	if err := templateWriter(conf, services); err != nil {
		log.Printf("error on templateWriter: %s", err.Error())
	}
}

func watch(conf Configuration, kapi client.KeysAPI) {
	watcherOpts := client.WatcherOptions{AfterIndex: 0, Recursive: true}
	watcher := kapi.Watcher(conf.Source.Path, &watcherOpts)
	for {
		resp, err := watcher.Next(context.Background())
		if err != nil {
			// TODO: execute before watch start again? wait to reduce load, in case of unrecoverable error?
			watch(conf, kapi)
		} else {
			action := resp.Action
			log.Printf("%s changed, action=%s", resp.Node.Key, action)
			execute(conf, kapi)
		}
	}
}

// Run creates the configuration for the services and updates the configuration whenever a service changed
func Run(conf Configuration, kapi client.KeysAPI) {
	execute(conf, kapi)
	log.Println("start service watcher")
	watch(conf, kapi)
}
