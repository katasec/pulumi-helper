package pulumihelper

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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

func RunPulumi(ctx context.Context, params *PulumiRunParameters) error {

	// Get run params
	//orgName := params.OrgName
	projectName := params.ProjectName
	stackName := params.StackName
	destroy := params.Destroy
	pulumiFn := params.PulumiFn
	outputStream := params.OutputStream

	// if orgName != "" {
	// 	stackName = auto.FullyQualifiedStackName(orgName, projectName, "dev")
	// }

	// Create stack
	s, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, pulumiFn)
	if err != nil {
		fmt.Printf("Failed to create or select stack: %v\n", err)
		// os.Exit(1)
		return err
	}
	fmt.Printf("Created/Selected stack %q\n", stackName)

	// Install the plugins if specified
	w := s.Workspace()
	if params.Plugins != nil {
		for _, plugin := range params.Plugins {
			fmt.Printf("Start Installing plugin %s:%s \n", plugin["name"], plugin["version"])

			err = w.InstallPlugin(ctx, "azure-native", "v1.64.1")
			if err != nil {
				fmt.Printf("Failed to install program plugins: %v\n", err)
				return err
			}

			fmt.Printf("End Installing plugin %s:%s \n", plugin["name"], plugin["version"])
		}

		fmt.Printf("All plugin installed! \n")
	}

	// Set stack config if specified:
	if params.Config != nil {
		// set stack configuration specifying the AWS region to deploy
		for _, key := range params.Config {
			s.SetConfig(ctx, key["name"], auto.ConfigValue{Value: key["value"]})
		}

		fmt.Println("Successfully set config")
	}

	// Always refresh stack before update
	fmt.Println("Starting refresh")
	_, err = s.Refresh(ctx)
	if err != nil {
		fmt.Printf("Failed to refresh stack: %v\n", err)
		//os.Exit(1)
		return err
	}
	fmt.Println("Refresh succeeded!")

	if destroy {
		fmt.Println("Starting stack destroy")

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
		outputStream.Close()
		fmt.Println("Stack successfully destroyed")
		return nil
	}

	fmt.Println("Starting update")

	// wire up our update to stream progress to stdout
	var stdoutStreamer optup.Option
	if outputStream != nil {
		stdoutStreamer = optup.ProgressStreams(os.Stdout, outputStream)
	} else {
		stdoutStreamer = optup.ProgressStreams(os.Stdout)
	}

	// Run pulumi Up
	_, err = s.Up(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to update stack: %v\n\n", err)
		return err
	}
	outputStream.Close()

	fmt.Println("Update succeeded!")

	return nil
}
