package service

import (
	"html/template"
	"log"
	"os"
	"path"

	"github.com/pkg/errors"
)

// TemplateModel is the input for the target template
type TemplateModel struct {
	Maintenance string
	Services    Services
}

type Writer interface {
	WriteTemplate(services TemplateModel) error
}

// CommandWriter implements the writer interface and executes pre- and post-commands
type CommandWriter struct {
	config Configuration
}

// WriteServices transform the data with a golang template
func (c *CommandWriter) WriteTemplate(data TemplateModel) error {
	if c.config.PreCommand != "" {
		err := preCheck(c.config, data)
		if err != nil {
			return errors.Wrap(err, "pre check failed")
		}
	}

	err := write(c.config, data)
	if err != nil {
		return errors.Wrap(err, "failed to write data")
	}

	if c.config.PostCommand != "" {
		err = post(c.config.PostCommand)
		if err != nil {
			return errors.Wrap(err, "post command failed")
		}
	}
	return nil
}

func write(config Configuration, data interface{}) error {
	name := path.Base(config.Template)
	tmpl, err := template.New(name).ParseFiles(config.Template)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	file, err := os.Create(config.Target)
	if err != nil {
		return errors.Wrapf(err, "failed to create target file %s", config.Target)
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
