package service

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	. "github.com/cloudogu/ces-confd/confd"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
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
	Source      Source
	Target      string
	Template    string
	Tag         string
	PreCommand  string `yaml:"pre-command"`
	PostCommand string `yaml:"post-command"`
	Order       Order
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
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		return errors.Wrap(err, "failed to render template")
	}
	return nil
}

func executeCommand(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	err := cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "failed to execute command: \"%s\"", command)
	}

	return cmd.Wait()
}

func post(command string) error {
	log.Println("execute post command", command)
	err := executeCommand(command)
	if err != nil {
		return errors.Wrap(err, "failed to execute post command")
	}
	return nil
}

func preCheck(conf Configuration, data interface{}) error {
	dir := filepath.Dir(conf.Target)
	prefix := filepath.Base(conf.Target)
	tmpFile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return errors.Wrap(err, "failed to create temp file for pre check")
	}

	defer os.Remove(tmpFile.Name())
	conf.Target = tmpFile.Name()
	err = write(conf, data)
	if err != nil {
		return errors.Wrap(err, "failed to write to temp file for pre check")
	}
	log.Println("execute pre command", conf.PreCommand)
	err = executeCommand(conf.PreCommand)
	if err != nil {
		return errors.Wrap(err, "pre check command failed")
	}
	return err
}

// naming
func execute(conf Configuration, kapi client.KeysAPI) {
	log.Println("read from etcd")
	services, err := readFromConfig(conf, kapi)
	if err != nil {
		log.Println("error durring read", err)
	}
	log.Printf("write services to template: %i", services)

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
			readAndWrite(conf, kapi)
		}
	}
}

func Run(conf Configuration, kapi client.KeysAPI) {
	execute(conf, kapi)
	log.Println("start service watcher")
	watch(conf, kapi)
}
