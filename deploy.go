package vessel

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

//
func Deploy() (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "get current directory")
	}

	config, err := NewConfig(wd)
	if err != nil {
		return errors.Wrap(err, "get config")
	}

	awsSession, err := session.NewSessionWithOptions(session.Options{
		Profile: config.Profile,
	})
	if err != nil {
		return errors.Wrap(err, "get aws session")
	}

	awsConfig := aws.NewConfig().WithRegion(config.Region)

	stsSvc := sts.New(awsSession)
	identity, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return errors.Wrap(err, "get caller identity")
	}

	// init registry
	if err := initRegistry(awsSession, awsConfig, identity, config); err != nil {
		return errors.Wrap(err, "init regisry")
	}

	// init pipelines
	pipelines, err := NewPipelines(wd)
	if err != nil {
		return errors.Wrap(err, "load pipelines")
	}

	// ensure functions exist
	lambdaSvc := lambda.New(awsSession, awsConfig)

	listFunctionsOutput, err := lambdaSvc.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		return errors.Wrap(err, "list existing functions")
	}

	existingFunctions := make(map[string]bool)
	for _, functionConfiguration := range listFunctionsOutput.Functions {
		existingFunctions[*functionConfiguration.FunctionName] = true
	}

	missingFunctions := NewFunctions()
	for _, pipeline := range pipelines {
		for _, function := range pipeline.Graph {
			if _, exists := existingFunctions[function.Name]; !exists {
				missingFunctions = append(missingFunctions, function)
			}
		}
	}

	log.Println("Missing Functions", missingFunctions)

	return
}

func initRegistry(awsSession *session.Session, awsConfig *aws.Config, identity *sts.GetCallerIdentityOutput, config *Config) error {
	ecrSvc := ecr.New(awsSession, awsConfig)
	vesselRepoName := aws.String("vessel")

	output, err := ecrSvc.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		RegistryId: identity.Account,
	})
	if err != nil {
		return errors.Wrap(err, "describe ecr repositories")
	}

	var foundVesselRepo bool
	for _, repo := range output.Repositories {
		if *repo.RepositoryName == *vesselRepoName {
			foundVesselRepo = true
		}
	}

	if !foundVesselRepo {
		createRepoOutput, err := ecrSvc.CreateRepository(&ecr.CreateRepositoryInput{
			RepositoryName: vesselRepoName,
		})
		if err != nil {
			if !strings.Contains(err.Error(), ecr.ErrCodeRepositoryAlreadyExistsException) {
				return errors.Wrap(err, "create ecr repository")
			}
		}
		config.ECR.RegistryID = *createRepoOutput.Repository.RegistryId
	} else {
		config.ECR.RegistryID = *identity.Account
	}

	return nil
}
