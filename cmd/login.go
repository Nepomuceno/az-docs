package cmd

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func Login() *azidentity.DefaultAzureCredential {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Authentication failure: %+v", err)
	}
	return cred
}
