package main

// Sving notes for later:
// 	pulumi config set azure-native:location eastus

import (
	"context"
)

func main() {

	// Setup Pulumi run parameters
	args := &PulumiRunParameters{
		OrgName:     "qigroup",
		ProjectName: "helloazure",
		StackName:   "dev",
		Destroy:     false,
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
