package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

// CreateWatcher creates a new watcher for the configuration entry
func CreateWatcher(kapi client.KeysAPI, entry Entry) Watcher {
	watcher := Watcher{
		kapi:  kapi,
		entry: entry,
		key:   entry.Source,
	}

	switch entry.Type {
	case "warp":
		watcher.reader = WarpReader
	case "service":
		watcher.reader = ServiceReader
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
	key    string
	entry  Entry
	reader DataReader
	writer DataWriter
}

func (w *Watcher) post() error {
	log.Println("execute post command", w.entry.PostCommand)
	cmd := exec.Command(w.entry.PostCommand)
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (w *Watcher) preCheck(data interface{}) error {
	dir := filepath.Dir(w.entry.Target)
	prefix := filepath.Base(w.entry.Target)
	tmpFile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return err
	}

	defer os.Remove(tmpFile.Name())

	err = w.writer(w.entry, tmpFile.Name(), data)
	if err != nil {
		return err
	}

	log.Println("execute pre command", w.entry.PreCommand)
	cmd := exec.Command(w.entry.PreCommand)
	err = cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (w *Watcher) write(data interface{}) error {
	if w.entry.PreCommand != "" {
		err := w.preCheck(data)
		if err != nil {
			return err
		}
	}

	err := w.writer(w.entry, w.entry.Target, data)
	if err != nil {
		return err
	}

	if w.entry.PostCommand != "" {
		err = w.post()
	}
	return err
}

func (w *Watcher) run() {
	data, err := w.reader(w.kapi, w.key)
	if err != nil {
		log.Println("Error durring read", err)
	} else {
		err = w.write(data)
		if err != nil {
			log.Println("Error durring write", err)
		}
	}
}

// Watch starts watching for changes in etcd
func (w *Watcher) Watch() {
	w.run()
	watcherOpts := client.WatcherOptions{AfterIndex: 0, Recursive: true}
	ew := w.kapi.Watcher(w.key, &watcherOpts)
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
