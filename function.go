package vessel

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

//
type Function struct {
	Name     string
	Next     string `yaml:"next"`
	Start    bool   `yaml:"start"`
	End      bool   `yaml:"end"`
	Count    int    `yaml:"count"`
	Parallel bool   `yaml:"parallel"`
	Type     NodeType
}

//
type NodeType int

//
const (
	Lambda NodeType = iota
	Task
)

//
func (f *Function) Init(name, functionPath string) error {
	f.Name = name
	f.Type = Lambda

	files, err := ioutil.ReadDir(functionPath)
	if err != nil {
		return errors.Wrap(err, "read stage dir")
	}

	var isTask bool
	for _, file := range files {
		if strings.Contains(file.Name(), "Dockerfile") {
			isTask = true
		}
	}

	if isTask {
		f.Type = Task
	}

	return nil
}
