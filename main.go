package main

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"
)

var (
	// Version of the application
	Version string
)

// RawData is a map of raw data, it can be used to unmarshal json data
type RawData map[string]interface{}

// DataReader fetches data from etcd
type DataReader func(kapi client.KeysAPI, entry Entry, root string) (interface{}, error)

// DataWriter writes data to disk
type DataWriter func(entry Entry, target string, data interface{}) error

// Configuration main configuration object
type Configuration struct {
	Endpoint string
	Entries  []Entry
}

// Entry is a configuration entry
type Entry struct {
	Source      string
	Target      string
	Type        string
	Template    string
	Tag         string
	PreCommand  string `yaml:"pre-command"`
	PostCommand string `yaml:"post-command"`
	Order       map[string]int
}

// Application struct
type Application struct {
	Configuration *Configuration
}

func contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (app *Application) startWatch(kapi client.KeysAPI, wg sync.WaitGroup, entry Entry) {
	watcher := CreateWatcher(kapi, entry)
	watcher.Watch()
	wg.Done()
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
		return nil, err
	}

	return client.NewKeysAPI(ec), nil
}

func (app *Application) readConfiguration(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, app.Configuration)
	if err != nil {
		return err
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

	var wg sync.WaitGroup
	for _, entry := range app.Configuration.Entries {
		wg.Add(1)
		go app.startWatch(kapi, wg, entry)
	}

	wg.Wait()
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
