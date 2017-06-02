package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
  "time"
  "github.com/cloudogu/ces-confd/confd/warp"
  "github.com/cloudogu/ces-confd/confd/service"
)

var (
	// Version of the application
	Version string
)

// Configuration main configuration object
type Configuration struct {
  Endpoint string
  Warp warp.Configuration
  Service service.Configuration
}

// Application struct
type Application struct {
	Configuration *Configuration
}

func (app *Application) createEtcdClient() (client.KeysAPI, error) {
	cfg := client.Config{
		Endpoints: []string{app.Configuration.Endpoint},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}

	ec, err := client.New(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create etcd client")
	}

	return client.NewKeysAPI(ec), nil
}

func (app *Application) readConfiguration(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "could not read configuration at "+path)
	}

	err = yaml.Unmarshal(data, app.Configuration)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal configuration "+path)
	}

	return nil
}

func (app *Application) run(c *cli.Context) {
	err := app.readConfiguration(c.String("config"))
	if err != nil {
		log.Fatal(err)
	}

  kapi, err := app.createEtcdClient()
	if err != nil {
		log.Fatal(err)
	}

  go warp.Run(app.Configuration.Warp, kapi)
  go service.Run(app.Configuration.Service, kapi)

  time.Sleep(10*time.Second)
}

func main() {
	config := Configuration{}
	application := Application{
		Configuration: &config,
	}

	app := cli.NewApp()
	app.Name = "ces-confd"
	app.Version = Version
	app.Usage = "watches etcd for changes and writes config files"
	app.Action = application.run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "endpoint, e",
			Value:       "http://localhost:2379",
			Usage:       "etcd endpoint",
			Destination: &config.Endpoint,
		},
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/ces-confd/config.yaml",
			Usage: "configuration path",
		},
	}

	app.Run(os.Args)
}
