package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path"
  "github.com/pkg/errors"
)

// TemplateWriter transform the data with a golang template
func TemplateWriter(entry Entry, target string, data interface{}) error {
	name := path.Base(entry.Template)
	tmpl, err := template.New(name).ParseFiles(entry.Template)
	if err != nil {
		return errors.Wrap(err, "failed to parse tempate")
	}

	file, err := os.Create(target)
	if err != nil {
		return errors.Wrapf(err, "failed to create target file %s", target)
	}
  defer file.Close()

	err = tmpl.Execute(file, data)
  if err != nil {
    return errors.Wrap(err, "failed to render template")
  }
  return nil
}

// JSONWriter converts the data to a json
func JSONWriter(entry Entry, target string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal data to json")
	}

	return ioutil.WriteFile(target, bytes, 0755)
}
