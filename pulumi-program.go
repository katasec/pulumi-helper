package main

import (
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/resources"
	"github.com/pulumi/pulumi-azure-native/sdk/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func pulumiFunc(ctx *pulumi.Context) error {
	// Create an Azure Resource Group
	resourceGroup, err := resources.NewResourceGroup(ctx, "resourceGroup", nil)
	if err != nil {
		return err
	}

	// Create an Azure resource (Storage Account)
	account, err := storage.NewStorageAccount(ctx, "sa", &storage.StorageAccountArgs{
		ResourceGroupName: resourceGroup.Name,
		AccessTier:        storage.AccessTierHot,
		Sku: &storage.SkuArgs{
			Name: pulumi.String(storage.SkuName_Standard_LRS),
		},
		Kind: pulumi.String(storage.KindStorageV2),
	})
	if err != nil {
		return err
	}

	// Pulumi.All gets the computed names for the inputs: resourcegroup name and account name
	// We use that to extract the storage key
	storageAccountKey := pulumi.All(resourceGroup.Name, account.Name).ApplyT(
		func(inputs []interface{}) (string, error) {
			return getStorageAccountKeys(ctx, inputs)
		},
	)

	// Export the storage key
	ctx.Export("primaryStorageKey", storageAccountKey)

	return nil
}

func getStorageAccountKeys(ctx *pulumi.Context, inputs []interface{}) (string, error) {
	resourceGroupName := inputs[0].(string)
	accountName := inputs[1].(string)
	accountKeys, err := storage.ListStorageAccountKeys(ctx, &storage.ListStorageAccountKeysArgs{
		ResourceGroupName: resourceGroupName,
		AccountName:       accountName,
	})
	if err != nil {
		return "", err
	}

	return accountKeys.Keys[0].Value, nil
}
