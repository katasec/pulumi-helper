package main

import (
	"context"
	"os"
	"strings"
)

// Sving notes for later:
// 	pulumi config set azure-native:location eastus

func main() {

	var destroy bool

	// Do pulumi destroy if passed in as an arg
	if len(os.Args) > 1 {
		if strings.ToLower(os.Args[1]) == "destroy" {
			destroy = true
		}
	} else {
		destroy = false
	}

	// Setup Pulumi run parameters
	args := &PulumiRunParameters{
		OrgName:     "qigroup",
		ProjectName: "helloazure",
		StackName:   "dev",
		Destroy:     destroy,
		Plugins: []map[string]string{
			{
				"name":    "azure-native",
				"version": "v1.64.1",
			},
		},
		PulumiFn: pulumiFunc,
	}

	// Run pulumi
	ctx := context.Background()
	RunPulumi(ctx, args)

}
