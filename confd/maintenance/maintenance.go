package maintenance

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"os"
	"path"

	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
)

const MaintenanceModePath = "/config/_global/maintenance"

type Source struct {
	path string
}

type Configuration struct {
	Source   Source
	Target   string
	Template string
}

type PageModel struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func (p PageModel) String() string {
	return p.Title + " " + p.Text
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

func renderTemplate(conf Configuration, etcdJson string) error {
	var pageModel PageModel
	err := json.Unmarshal([]byte(etcdJson), &pageModel)

	if err != nil {
		return errors.Wrapf(err, "Could not parse JSON for maintenance page")
	}

	return write(conf, pageModel)
}

func watchForMaintenanceMode(conf Configuration, kapi client.KeysAPI) {
	watcherOpts := client.WatcherOptions{AfterIndex: 0, Recursive: true}
	watcher := kapi.Watcher(MaintenanceModePath, &watcherOpts)
	for {
		resp, err := watcher.Next(context.Background())
		if err != nil {
			watchForMaintenanceMode(conf, kapi)
		} else {
			log.Println("Change in maintenance mode config: " + resp.Node.Value)
			if resp.Node.Value != "" {
				log.Println("Rendering maintenance mode template")
				err = renderTemplate(conf, resp.Node.Value)
				if err != nil {
					log.Printf("Error rendering template: %v", err)
				}
			}
		}
	}
}

func Run(conf Configuration, kapi client.KeysAPI) {
	log.Println("Starting maintenance mode watcher...")
	watchForMaintenanceMode(conf, kapi)
}
