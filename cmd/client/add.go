package client

import (
"fmt"
"github.com/SPSCommerce/spsauth0/common"

//"github.com/manifoldco/promptui"
"os"
//"strconv"

//"strconv"

"github.com/SPSCommerce/spsauth0/internal/config"
"github.com/spf13/cobra"
)

var (

	clientAddCmd = &cobra.Command{
		Use:   "add <client name>",
		Short: "Add a auth0 client/ devcenter application",
		Long:  "client add",
		//Args:  cobra.ExactArgs(1),
		Run:   clientAddExecute,
	}
)

//func init() {
//}

func clientAddExecute(cmd *cobra.Command, args []string) {
	// Grab Client name from viper root and ensure it does not exist yet
	clientConfig, err := config.LoadClientConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load client config - %v\n", err)
		os.Exit(1)
	}

	tenantConfig, err := config.LoadTenantConfigWithViper()
	if len(*tenantConfig.GetTenantProfileList()) == 0 {
		fmt.Printf("You must first configure a tenant before adding a client \n run `spsauth0 tenant add <tenant name>` ")
		os.Exit(1)
	}

	// Prompt to get client name
	clientName, err := common.PromptString("Client Name", "", false)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Check that a client does not already exist in the config with the same name
	client := clientConfig.GetClientConfig(clientName)
	if client != nil {
		fmt.Printf("Client already exists with name %s, use 'spsauth0 client update %s instead'", clientName, clientName)
		os.Exit(1)
	}

	// Prompt to get clientId
	clientId, err := common.PromptString("Client Id", "", false)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Add some type of Validation to clientId

	clientSecret, err := common.PromptString("Client Secret", "", false)
	if err != nil {
		fmt.Println(err.Error())
	}

	_, clientType, err := common.PromptSelect("Client Type", config.GetSupportedClientTypes())
	if err != nil {
		fmt.Println(err.Error())
	}

	_, tenant, err := common.PromptSelect("Tenant", tenantConfig.GetTenantListNames())
	if err != nil {
		fmt.Println(err.Error())
	}

	newClient := &config.Client{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		ClientName:   clientName,
		ClientType:   clientType,
		TenantName: tenant,
	}

	// Save Client to config
	clientConfig.SetClient(newClient)

	err = clientConfig.SaveClientConfig()
	if err != nil {
		fmt.Println("Failed to save client configuration: ", err)
		os.Exit(1)
	}
}
