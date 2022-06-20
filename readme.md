# Overview

This library contains convenience functions for using the [Pulumi Automation API](https://www.pulumi.com/docs/guides/automation-api/) to run inline Pulumi programs.

## Example

### Setup necessary arguments to run you Pulumi program

The organization, project and stackname are provided as parameters as per below. Set the `Destroy` field to `true` for a `pulumi destory` else, set it to `true` for a `pulumi up`.

The `Plugins` field is an array containing the name/versions of the pulumi plugins required by your pulumi program.

Finally, the `PulumiFn` field is a Go function that contains the pulumi program you want to run.

```
	// Setup Pulumi run parameters
	args := &pulumihelper.PulumiRunParameters{
		OrgName:     "acme",
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
```

Run your Pulumi program using the following code:


```
	// Run pulumi
	ctx := context.Background()
	pulumihelper.RunPulumi(ctx, args)
```

Check out [sample.go](./sample.go) for an example.