package service

import (
	"log"
	"os"
	"os/exec"

  "github.com/pkg/errors"
  "github.com/satori/go.uuid"
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
  _, statErr := os.Stat(conf.Target)
  if statErr == nil {
    err := executePreCheckWithExistingConfiguration(conf, data)
    if err != nil {
      return errors.Wrap(err, "failed to execute pre check with existing configuration")
    }

    return nil
  } else if !os.IsNotExist(statErr) {
    return errors.Wrapf(statErr, "failed to stat file %s", conf.Target)
  }

  err := executePreCheckWithNewConfiguration(conf, data)
  if err != nil {
    return errors.Wrap(err, "failed to execute pre check with new configuration")
  }

  return nil
}

func executePreCheckWithExistingConfiguration(conf Configuration, data interface{}) error {
  tmpPath := createTempPath(conf)

  log.Printf("move configuration %s to tempoarary location %s for pre check", conf.Target, tmpPath)

  err := os.Rename(conf.Target, tmpPath)
  if err != nil {
    return errors.Wrapf(err, "failed to rename %s to temp file %s", conf.Target, tmpPath)
  }

  defer func(){
    log.Printf("move temporary configuration %s back to original location %s", tmpPath, conf.Target)
    err := os.Rename(tmpPath, conf.Target)
    if err != nil {
      log.Printf("failed to rename temporary configuration %s back to original location %s: %v", tmpPath, conf.Target, err)
    }
  }()

  err = executePreCheck(conf, data)
  if err != nil {
    return err
  }

  return nil
}

func createTempPath(conf Configuration) string {
  return conf.Target + ".ces-confd-" + uuid.NewV4().String()
}

func executePreCheckWithNewConfiguration(conf Configuration, data interface{}) error {
  defer func(){
    err := os.Remove(conf.Target)
    if err != nil {
      log.Println("failed to remove temporary configuration", conf.Target)
    }
  }()

  err := executePreCheck(conf, data)
  if err != nil {
    return err
  }

  return nil
}

func executePreCheck(conf Configuration, data interface{}) error {
  err := write(conf, data)
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
