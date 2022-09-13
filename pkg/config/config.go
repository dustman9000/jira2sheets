package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"gopkg.in/errgo.v2/errors"
)

type Config struct {
	Spreadsheets       []Spreadsheet      `json:"spreadsheets"`
	ActiveSprintsSheet ActiveSprintsSheet `json:"activeSprintsSheet,omitempty"`
}

type Spreadsheet struct {
	Url        string `json:"url"`
	SheetName  string `json:"sheetName"`
	JiraFilter string `json:"jiraFilter"`
}

type ActiveSprintsSheet struct {
	Url          string `json:"url"`
	SheetName    string `json:"sheetName"`
	JiraEndpoint string `json:"jiraEndpoint"`
}

func ReadConfig(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, errors.Because(err, nil, fmt.Sprintf("reading %s", path))
	}
	return &config, nil
}
