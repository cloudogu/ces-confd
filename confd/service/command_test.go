package service

import (
  "testing"
  "io/ioutil"
  "os"
  "path"
  "github.com/stretchr/testify/require"
)

func TestPreCheck(t *testing.T) {
  directory, err := ioutil.TempDir("", "")
  require.Nil(t, err)
  defer func() {
    os.RemoveAll(directory)
  }()

  template := path.Join(directory, "template")
  err = ioutil.WriteFile(template, []byte("{{.Name}}"), 0644)
  require.Nil(t, err)

  target := path.Join(directory, "target")
  err = ioutil.WriteFile(target, []byte("trillian"), 0644)
  require.Nil(t, err)

  conf := Configuration{
    Target: target,
    Template: template,
    PreCommand: "grep slarti " + target,
  }

  err = preCheck(conf, model{"slarti"})
  require.Nil(t, err)

  // be sure old file gets restored
  _, err = os.Stat(target)
  require.Nil(t, err)

  content, err := ioutil.ReadFile(target)
  require.Nil(t, err)

  require.Equal(t, "trillian", string(content))
}

func TestPreCheckFailed(t *testing.T) {
  directory, err := ioutil.TempDir("", "")
  require.Nil(t, err)
  defer func() {
    os.RemoveAll(directory)
  }()

  template := path.Join(directory, "template")
  err = ioutil.WriteFile(template, []byte("{{.Name}}"), 0644)
  require.Nil(t, err)

  target := path.Join(directory, "target")
  err = ioutil.WriteFile(target, []byte("trillian"), 0644)
  require.Nil(t, err)

  // grep trillian should fail, because the pre check writes slarti to the target

  conf := Configuration{
    Target: target,
    Template: template,
    PreCommand: "grep trillian " + target,
  }

  err = preCheck(conf, model{"slarti"})
  require.NotNil(t, err)

  // be sure old file gets restored
  content, err := ioutil.ReadFile(target)
  require.Nil(t, err)

  require.Equal(t, "trillian", string(content))
}

func TestPreCheckWithoutExistingConfiguration(t *testing.T) {
  directory, err := ioutil.TempDir("", "")
  require.Nil(t, err)
  defer func() {
    os.RemoveAll(directory)
  }()

  template := path.Join(directory, "template")
  err = ioutil.WriteFile(template, []byte("{{.Name}}"), 0644)
  require.Nil(t, err)

  target := path.Join(directory, "target")

  conf := Configuration{
    Target: target,
    Template: template,
    PreCommand: "grep trillian " + target,
  }

  err = preCheck(conf, model{"slarti"})
  require.NotNil(t, err)

  // be sure target does not exists
  _, err = os.Stat(target)
  require.True(t, os.IsNotExist(err))
}

type model struct {
  Name string
}
