package service

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

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
