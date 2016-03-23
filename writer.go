package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path"
)

// TemplateWriter transform the data with a golang template
func TemplateWriter(entry Entry, target string, data interface{}) error {
	name := path.Base(entry.Template)
	tmpl, err := template.New(name).ParseFiles(entry.Template)
	if err != nil {
		return err
	}

	file, err := os.Create(target)
	defer file.Close()
	if err != nil {
		return err
	}
	return tmpl.Execute(file, data)
}

// JSONWriter converts the data to a json
func JSONWriter(entry Entry, target string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(target, bytes, 0755)
}
