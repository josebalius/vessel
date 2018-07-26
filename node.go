package vessel

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

//
type Node struct {
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
func (n *Node) Init(name, nodePath string) error {
	n.Name = name
	n.Type = Lambda

	files, err := ioutil.ReadDir(nodePath)
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
		n.Type = Task
	}

	return nil
}
