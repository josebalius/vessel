package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

// Input ...
type Input struct {
	Value string
}

// Output ...
type Output struct {
	Value string
}

// Handler ...
func Handler(input Input) (Output, error) {
	fmt.Println(input)
	return Output{Value: "output"}, nil
}

func main() {
	lambda.Start(Handler)
}
