package pulumihelper

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PulumiRunRemoteParameters struct {
	OrgName      string              // Name of the Pulumi Organisation for your stack
	ProjectName  string              // Name of the Pulumi project to create/destroy
	GitURL       string              // For example github.com/katasec/project.git
	ProjectPath  string              // A sub folder under  github.com/katasec/project.git for e.g. folder1
	StackName    string              // Name of your pulumi stack. For e.g. "dev" or "prod"
	Destroy      bool                // False to create stack. True to destroy your pulumi stack.
	Plugins      []map[string]string // Plugins required for your Pulumi program, Specified as "name" and "version" in string map
	OutputStream *io.PipeWriter
	Config       []map[string]string // Config for your pulumi program, specified as "name" and "value" in string map
}

type PulumiRunParameters struct {
	OrgName      string              // Name of the Pulumi Organisation for your stack
	ProjectName  string              // Name of the Pulumi project to create/destroy
	StackName    string              // Name of your pulumi stack. For e.g. "dev" or "prod"
	Destroy      bool                // False to create stack. True to destroy your pulumi stack.
	Plugins      []map[string]string // Plugins required for your Pulumi program, Specified as "name" and "version" in string map
	PulumiFn     pulumi.RunFunc      // Your pulumi program you want to run, passed in as a function
	OutputStream *io.PipeWriter
	Config       []map[string]string // Config for your pulumi program, specified as "name" and "value" in string map
}

//func CreateLocalStack(ctx context.Context, params *PulumiRunParameters) (auto.Stack, error) {
//	// Create stack
//	s, err := auto.UpsertStackInlineSource(ctx, params.StackName, params.ProjectName, params.PulumiFn)
//	if err != nil {
//		fmt.Printf("Failed to create or select stack: %v\n", err)
//		return s, err
//	}
//	fmt.Printf("Created/Selected stack %q\n", params.StackName)
//
//	// Return stack
//	return s, nil
//}

func CreateRemoteStack(ctx context.Context, params *PulumiRunRemoteParameters) (auto.Stack, error) {

	// arguments used to set up the remote Pulumi program
	repo := auto.GitRepo{
		URL:         params.GitURL,
		ProjectPath: params.ProjectPath,
	}

	// Define project
	project, _ := defaultInlineProject(params.ProjectName)
	options := auto.Project(project)

	// Create stack
	s, err := auto.UpsertStackRemoteSource(ctx, params.StackName, repo, options)
	if err != nil {
		fmt.Printf("Failed to create or select stack: %v\n", err)
		return s, err
	}
	fmt.Printf("Created/Selected stack %q\n", params.StackName)

	// Return stack
	return s, nil
}

func setConfig(ctx context.Context, s auto.Stack, config []map[string]string) (auto.Stack, error) {
	// Set stack config if specified:
	if config != nil {
		// set stack configuration specifying the AWS region to deploy
		for _, key := range config {
			err := s.SetConfig(ctx, key["name"], auto.ConfigValue{Value: key["value"]})
			if err != nil {
				return s, err
			}
		}

		fmt.Println("Successfully set config")
	}

	return s, nil
}

func installPlugins(ctx context.Context, s auto.Stack, plugins []map[string]string) (auto.Stack, error) {
	// Install the plugins if specified
	w := s.Workspace()
	if plugins != nil {
		for _, key := range plugins {
			fmt.Printf("Start Installing plugin %s:%s \n", key["name"], key["version"])

			err := w.InstallPlugin(ctx, "azure-native", "v1.64.1")
			if err != nil {
				fmt.Printf("Failed to install program plugins: %v\n", err)
				return s, err
			}

			fmt.Printf("End Installing plugin %s:%s \n", key["name"], key["version"])
		}
		fmt.Printf("All plugin installed! \n")
	}
	return s, nil
}

func refreshStack(ctx context.Context, s auto.Stack) error {
	fmt.Println("Starting refresh")

	result, err := s.Refresh(ctx)
	if err != nil {
		fmt.Printf("Failed to refresh stack: %v\n", err)
		return err
	}

	fmt.Printf("Refresh succeeded!, Result:%s \n", result.Summary.Result)

	return nil
}

func pulumiDestroy(ctx context.Context, s auto.Stack, outputStream *io.PipeWriter) error {
	fmt.Println("Starting stack destroy")

	// Add extra output stream if specified
	var stdoutStreamer optdestroy.Option
	if outputStream != nil {
		stdoutStreamer = optdestroy.ProgressStreams(os.Stdout, outputStream)
	} else {
		stdoutStreamer = optdestroy.ProgressStreams(os.Stdout)
	}

	// destroy our stack and exit early
	_, err := s.Destroy(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to destroy stack: %v", err)
		return err
	}

	if outputStream != nil {
		outputStream.Close()
	}

	fmt.Println("Stack successfully destroyed")
	return nil
}

func pulumiUp(ctx context.Context, s auto.Stack, outputStream *io.PipeWriter) error {
	fmt.Println("Starting update")

	// Add extra output stream if specified
	var stdoutStreamer optup.Option
	if outputStream != nil {
		stdoutStreamer = optup.ProgressStreams(os.Stdout, outputStream)
	} else {
		stdoutStreamer = optup.ProgressStreams(os.Stdout)
	}

	// Run pulumi Up
	_, err := s.Up(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to update stack: %v\n\n", err)
		return err
	}
	if outputStream != nil {
		outputStream.Close()
	}

	fmt.Println("Update succeeded!")
	return nil
}

func RunPulumi(ctx context.Context, params *PulumiRunParameters) error {

	// Get run params
	//orgName := params.OrgName
	projectName := params.ProjectName
	stackName := params.StackName
	destroy := params.Destroy
	pulumiFn := params.PulumiFn
	outputStream := params.OutputStream

	// Create stack
	s, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, pulumiFn)
	if err != nil {
		fmt.Printf("Failed to create or select stack: %v\n", err)
		return err
	}
	fmt.Printf("Created/Selected stack %q\n", stackName)

	// Install the plugins if specified
	s, _ = installPlugins(ctx, s, params.Plugins)

	// Set stack config if specified:
	s, _ = setConfig(ctx, s, params.Config)

	// Always refresh stack before update
	refreshStack(ctx, s)

	// Run pulumi
	if destroy {
		err := pulumiDestroy(ctx, s, outputStream)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}
	} else {
		pulumiUp(ctx, s, outputStream)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}
	}

	return nil
}

func RunPulumiRemote(ctx context.Context, params *PulumiRunRemoteParameters) error {

	// Get run params
	stackName := params.StackName
	destroy := params.Destroy
	outputStream := params.OutputStream

	// Create stack
	s, err := CreateRemoteStack(ctx, params)
	if err != nil {
		fmt.Printf("Failed to create or select stack: %v\n", err)
		return err
	}
	fmt.Printf("Created/Selected stack %q\n", stackName)

	// Install the plugins if specified
	s, _ = installPlugins(ctx, s, params.Plugins)

	// Set stack config if specified:
	s, _ = setConfig(ctx, s, params.Config)

	// Always refresh stack before update
	refreshStack(ctx, s)

	// Run pulumi
	if destroy {
		err := pulumiDestroy(ctx, s, outputStream)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}
	} else {
		pulumiUp(ctx, s, outputStream)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}
	}

	return nil
}

func defaultInlineProject(projectName string) (workspace.Project, error) {
	var proj workspace.Project
	cwd, err := os.Getwd()
	if err != nil {
		return proj, err
	}
	proj = workspace.Project{
		Name:    tokens.PackageName(projectName),
		Runtime: workspace.NewProjectRuntimeInfo("go", nil),
		Main:    cwd,
	}

	return proj, nil
}
