package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PulumiRunParameters struct {
	OrgName     string              // Name of the Pulumi Organisation for your stack
	ProjectName string              // Name of the Pulumi project to create/destroy
	StackName   string              // Name of your pulumi stack. For e.g. "dev" or "prod"
	Destroy     bool                // False to create stack. True to destroy your pulumi stack.
	Plugins     []map[string]string // Plugins required for your Pulumi program
	PulumiFn    pulumi.RunFunc      // Your pulumi program you want to run passed in as a function
}

func RunPulumi(ctx context.Context, params *PulumiRunParameters) {

	// Get run params
	//orgName := params.OrgName
	projectName := params.ProjectName
	stackName := params.StackName
	destroy := params.Destroy
	pulumiFn := params.PulumiFn

	// if orgName != "" {
	// 	stackName = auto.FullyQualifiedStackName(orgName, projectName, "dev")
	// }

	// Create stack
	s, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, pulumiFn)
	if err != nil {
		fmt.Printf("Failed to create or select stack: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created/Selected stack %q\n", stackName)

	// Install the plugins
	w := s.Workspace()
	for _, plugin := range params.Plugins {
		fmt.Printf("Installing plugin %s:%s \n", plugin["name"], plugin["version"])

		err = w.InstallPlugin(ctx, "azure-native", "v1.64.1")
		if err != nil {
			fmt.Printf("Failed to install program plugins: %v\n", err)
			os.Exit(1)
		}
	}

	// set stack configuration specifying the AWS region to deploy
	s.SetConfig(ctx, "azure-native:location", auto.ConfigValue{Value: "EastAsia"})

	fmt.Println("Successfully set config")
	fmt.Println("Starting refresh")

	_, err = s.Refresh(ctx)
	if err != nil {
		fmt.Printf("Failed to refresh stack: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Refresh succeeded!")

	if destroy {
		fmt.Println("Starting stack destroy")

		// wire up our destroy to stream progress to stdout
		stdoutStreamer := optdestroy.ProgressStreams(os.Stdout)

		// destroy our stack and exit early
		_, err := s.Destroy(ctx, stdoutStreamer)
		if err != nil {
			fmt.Printf("Failed to destroy stack: %v", err)
		}
		fmt.Println("Stack successfully destroyed")
		os.Exit(0)
	}

	fmt.Println("Starting update")

	// wire up our update to stream progress to stdout
	stdoutStreamer := optup.ProgressStreams(os.Stdout)

	// Run pulumi Up
	_, err = s.Up(ctx, stdoutStreamer)
	if err != nil {
		fmt.Printf("Failed to update stack: %v\n\n", err)
		os.Exit(1)
	}

	fmt.Println("Update succeeded!")

}
