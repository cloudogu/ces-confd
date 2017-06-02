package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
  "github.com/pkg/errors"
)

// CreateWatcher creates a new watcher for the configuration entry
func CreateWatcher(kapi client.KeysAPI, entry Entry) Watcher {
	watcher := Watcher{
		kapi:  kapi,
		entry: entry,
	}

  log.Printf("CreateWatcher from source %i", entry.Source)
  for _, source := range entry.Source {
    log.Printf("found source %i", source)
    var read func(kapi client.KeysAPI, entry Entry, root string) (interface{}, error)
    log.Printf("reader for source %s", source.Reader)
    switch source.Reader {
    case "dogus":
      read = DogusWarpReader
    case "externals":
      read = ExternalsWarpReader
    case "service":
      read = ServiceReader
    }
    if read != nil {
      newReader := DataReader{rootPath:source.Path, read:read}
      watcher.reader = append(watcher.reader, newReader)
    }
  }

	if entry.Template != "" {
		watcher.writer = TemplateWriter
	} else {
		watcher.writer = JSONWriter
	}

	return watcher
}

// Watcher watches etcd and writes configuration files
type Watcher struct {
	kapi   client.KeysAPI
	entry  Entry
	reader []DataReader
	writer DataWriter
}

func execute(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	err := cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "failed to execute command: \"%s\"", command)
	}

	return cmd.Wait()
}

func (w *Watcher) post() error {
	log.Println("execute post command", w.entry.PostCommand)
	err := execute(w.entry.PostCommand)
  if err != nil {
    return errors.Wrap(err, "failed to execute post command")
  }
  return nil
}

func (w *Watcher) preCheck(data interface{}) error {
	dir := filepath.Dir(w.entry.Target[0].Path)
	prefix := filepath.Base(w.entry.Target[0].Path)
	tmpFile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return errors.Wrap(err, "failed to create temp file for pre check")
	}

	defer os.Remove(tmpFile.Name())

	err = w.writer(w.entry, tmpFile.Name(), data)
	if err != nil {
		return errors.Wrap(err, "failed to write to temp file for pre check")
	}

	log.Println("execute pre command", w.entry.PreCommand)
	err = execute(w.entry.PreCommand)
  if err != nil {
    return errors.Wrap(err, "pre check command failed")
  }
  return err
}

func (w *Watcher) write(data interface{}) error {
	if w.entry.PreCommand != "" {
		err := w.preCheck(data)
		if err != nil {
			return errors.Wrap(err, "pre check failed")
		}
	}

	err := w.writer(w.entry, w.entry.Target[0].Path, data)
	if err != nil {
		return errors.Wrap(err, "failed to write data")
	}

	if w.entry.PostCommand != "" {
		err = w.post()
    if err != nil {
      return errors.Wrap(err, "post command failed")
    }
	}
	return nil
}

func (w *Watcher) run() {
  var data []interface{}
  for _, reader := range w.reader {
    readData, err := reader.read(w.kapi, w.entry, reader.rootPath)
    log.Printf("readData %i",readData)
    if err != nil {
      log.Println("Error durring read", err)
    }
    data = append(data, readData.. )
  }
  log.Printf("found data %i", data)

  if len(data)>0 {
    err := w.write(data)
    if err != nil {
      log.Println("Error durring write", err)
    }
  }
}

// Watch starts watching for changes in etcd
func (w *Watcher) Watch() {
	w.run()
	watcherOpts := client.WatcherOptions{AfterIndex: 0, Recursive: true}
  log.Printf("register watcher for %s", w.entry.Source[0].Path)
	ew := w.kapi.Watcher(w.entry.Source[0].Path, &watcherOpts) // TODO: parallel watch für alle reader
	for {
		resp, err := ew.Next(context.Background())
		if err != nil {
			log.Println("Error durring watch", err)
		} else {
			action := resp.Action
			log.Printf("%s changed, action=%s", resp.Node.Key, action)
			w.run()
		}
	}
}
