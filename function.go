package vessel

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

// Function ...
type Function struct {
	Name     string
	Path     string
	Next     string `yaml:"next"`
	Start    bool   `yaml:"start"`
	End      bool   `yaml:"end"`
	Count    int    `yaml:"count"`
	Parallel bool   `yaml:"parallel"`
	Type     NodeType
}

// NodeType ...
type NodeType int

//
const (
	Lambda NodeType = iota
	Task
)

// LambdaType ...
type LambdaType int

//
const (
	GoLambda LambdaType = iota
	NodeLambda
	PythonLambda
)

func (l LambdaType) String() string {
	switch l {
	case 0:
		return "Go Lambda"
	case 1:
		return "Node Lambda"
	case 2:
		return "Python Lambda"
	default:
		return ""
	}
}

// Init ...
func (f *Function) Init(name, functionPath string) error {
	f.Name = "vessel-" + name
	f.Type = Lambda
	f.Path = functionPath

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

// Install ...
func (f *Function) Install(lambdaSvc *lambda.Lambda) error {
	switch f.Type {
	case Lambda:
		files, err := ioutil.ReadDir(f.Path)
		if err != nil {
			return errors.Wrap(err, "read lambda files")
		}
		var lambdaType LambdaType
		for _, f := range files {
			filename := f.Name()
			if strings.HasSuffix(filename, ".go") {
				lambdaType = GoLambda
			}
			if strings.HasSuffix(filename, ".py") {
				lambdaType = PythonLambda
			}
			if strings.HasSuffix(filename, ".js") {
				lambdaType = NodeLambda
			}
		}
		switch lambdaType {
		case GoLambda:
			return installGoLambda(lambdaSvc, f)
		default:
			return fmt.Errorf("Lambda Type: '%v' is not supported", lambdaType)
		}
	case Task:
		fmt.Println("Installing task function:", f.Name)
	default:
		return fmt.Errorf("Unsupported function type: '%v'", f.Type)
	}
	return nil
}

// Functions ...
type Functions []*Function

// NewFunctions ...
func NewFunctions() Functions {
	return Functions{}
}

func installGoLambda(lambdaSvc *lambda.Lambda, function *Function) error {
	cmd := exec.Command("env", "GOOS=linux", "go", "build")
	cmd.Dir = function.Path
	if err := runCommand(cmd); err != nil {
		return errors.Wrap(err, "run go build")
	}

	dirName, err := filepath.Abs(function.Path)
	if err != nil {
		return errors.Wrap(err, "get function dir name")
	}

	binaryName := fmt.Sprintf("./%v", dirName)

	cmd = exec.Command("zip", "handler.zip", binaryName)
	cmd.Dir = function.Path
	if err := runCommand(cmd); err != nil {
		return errors.Wrap(err, "zip go binary")
	}

	zipFile, err := ioutil.ReadFile(filepath.Join(function.Path, "handler.zip"))
	if err != nil {
		return errors.Wrap(err, "read function zip file")
	}

	code := &lambda.FunctionCode{
		ZipFile: zipFile,
	}

	lambdaSvc.CreateFunction(&lambda.CreateFunctionInput{
		FunctionName: aws.String(function.Name),
		MemorySize:   aws.Int64(128),
		Runtime:      aws.String("go1.x"),
		Code:         code,
	})

	return nil
}
