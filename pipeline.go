package vessel

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v1"
)

//
type Pipeline struct {
	Name  string
	Graph map[string]*Function
}

//
func NewPipeline(projectPath, filePath string) (*Pipeline, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "read pipeline file")
	}

	pipeline := &Pipeline{}
	if err := yaml.Unmarshal(file, &pipeline.Graph); err != nil {
		return nil, errors.Wrap(err, "unmarshal pipeline graph")
	}

	pipeline.Name = strings.TrimSuffix(filepath.Base(filePath), ".yaml")
	for functionName, function := range pipeline.Graph {
		functionPath := filepath.Join(projectPath, "functions", functionName)
		if err := function.Init(functionName, functionPath); err != nil {
			return nil, errors.Wrap(err, "init function")
		}
	}

	return pipeline, nil
}

//
type Pipelines []*Pipeline

//
func NewPipelines(projectPath string) (Pipelines, error) {
	pipelinesDir := filepath.Join(projectPath, "pipelines")
	files, err := ioutil.ReadDir(pipelinesDir)
	if err != nil {
		return nil, errors.Wrap(err, "read pipelines files")
	}

	pipelines := Pipelines{}
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, "yaml") {
			pipeline, err := NewPipeline(projectPath, filepath.Join(pipelinesDir, filename))
			if err != nil {
				return nil, errors.Wrap(err, "new pipeline")
			}
			pipelines = append(pipelines, pipeline)
		}
	}

	return pipelines, nil
}
