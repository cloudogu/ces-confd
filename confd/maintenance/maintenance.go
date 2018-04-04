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

// Source of maintenance path in etcd
type Source struct {
	Path string
}

// PageModel is the input to render the maintenance page template
type PageModel struct {
  Title string `json:"title"`
  Text  string `json:"text"`
}

func (p PageModel) String() string {
  return p.Title + " " + p.Text
}

// Configuration for the maintenance modul
type Configuration struct {
	Source   Source
	Target   string
	Template string
	Default  PageModel
}

func write(conf Configuration, pageModel PageModel) error {
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

	err = tmpl.Execute(file, &pageModel)
	if err != nil {
		return errors.Wrap(err, "failed to render template")
	}
	return nil
}

func renderTemplate(conf Configuration, value string) error {
  log.Println("render maintenance page:", value)

	var pageModel PageModel
	err := json.Unmarshal([]byte(value), &pageModel)
	if err != nil {
		return errors.Wrapf(err, "Could not parse JSON for maintenance page")
	}

	return write(conf, pageModel)
}

func renderDefault(conf Configuration) {
  log.Println("render default maintenance page")
  err := write(conf, conf.Default)
  if err != nil {
    log.Printf("failed to write template with default: %v", err)
  }
}

func readAndRender(conf Configuration, kapi client.KeysAPI) {
  resp, err := kapi.Get(context.Background(), conf.Source.Path, nil)
  if err != nil {
    if client.IsKeyNotFound(err) {
      renderDefault(conf)
      return
    }

    log.Printf("failed to read key %s: %v", conf.Source.Path, err)
    return
  }

  err = renderTemplate(conf, resp.Node.Value)
  if err != nil {
    log.Printf("failed to render template with model %s: %v", resp.Node.Value, err)
  }
}

func watchForMaintenanceMode(conf Configuration, kapi client.KeysAPI) {
	watcherOpts := client.WatcherOptions{AfterIndex: 0, Recursive: false}
	watcher := kapi.Watcher(conf.Source.Path, &watcherOpts)
	for {
		_, err := watcher.Next(context.Background())
		if err != nil {
			watchForMaintenanceMode(conf, kapi)
		} else {
			log.Println("Change in maintenance mode config")
			readAndRender(conf, kapi)
		}
	}
}

// Run renders the maintenance page and watches for changes
func Run(conf Configuration, kapi client.KeysAPI) {
  readAndRender(conf, kapi)

	log.Println("Starting maintenance mode watcher...")
	watchForMaintenanceMode(conf, kapi)
}
