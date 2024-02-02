package main

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	_ "embed"
)

//go:embed check.go.tmpl
var checkTmpl string

const (
	configFile = "config.yml"
	checksDir  = "pkg/checks/"
)

func main() {
	config, err := os.ReadFile(configFile)
	if err != nil {
		logrus.WithError(err).Fatal("failed to read config file")
	}

	var data struct {
		Checks map[string]string `yaml:"checks"`
	}

	err = yaml.Unmarshal(config, &data)
	if err != nil {
		logrus.WithError(err).Fatal("failed to unmarshal config file")
	}

	err = filepath.Walk(checksDir, cleanChecksDir)
	if err != nil {
		logrus.WithError(err).Fatal("failed to clean checks directory")
	}

	tmpl, err := template.New("check.go.tmpl").Parse(checkTmpl)
	if err != nil {
		logrus.WithError(err).Fatal("failed to parse template")
	}

	for name, remote := range data.Checks {
		var cleanRemote string
		if strings.Contains(remote, "@") {
			cleanRemote = strings.Split(remote, "@")[1]
		} else {
			cleanRemote = remote
		}

		out, err := os.Create(filepath.Join(checksDir, name+".go"))
		if err != nil {
			logrus.WithError(err).Fatalf("failed to create check file: \"%s.go\"", name)
		}

		err = tmpl.Execute(out, struct {
			Name        string
			Remote      string
			CleanRemote string
		}{
			Name:        name,
			Remote:      remote,
			CleanRemote: cleanRemote,
		})
		if err != nil {
			logrus.WithError(err).Fatal("failed to execute template")
		}
	}
}

func cleanChecksDir(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.Name() != "main.go" && !info.IsDir() {
		return os.Remove(path)
	}

	return nil
}
