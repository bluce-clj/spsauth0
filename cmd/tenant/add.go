package tenant

import (
	"fmt"
	"github.com/bluce-clj/spsauth0/common"
	"strings"

	"os"

	"github.com/bluce-clj/spsauth0/internal/config"
	"github.com/spf13/cobra"
)

var (

	tenantAddCmd = &cobra.Command{
		Use:   "add <tenant name>",
		Short: "Add a tenant",
		Long:  "tenant add",
		Args:  cobra.ExactArgs(1),
		Run:   tenantAddExecute,
	}
)

func init() {
	//awsCredAddCmd.Flags().String(config.FlagCmdAWSCloneProfile,
	//	"", "Clone an existing profile")
	//viper.BindPFlag(config.KeyCmdAWSCloneProfile,
	//	awsCredAddCmd.Flags().Lookup(config.FlagCmdAWSCloneProfile))
}

func tenantAddExecute(cmd *cobra.Command, args []string) {
	tenantName := getTenanteArg(args)
	tenantConfig, err := config.LoadTenantConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	tenant := tenantConfig.GetTenantConfig(tenantName)
	if tenant != nil {
		fmt.Printf("Tenant already exists with name %s, use 'spsauth0 tenant update %s instead'", tenantName, tenantName)
		os.Exit(1)
	}

	domain, err := common.PromptString("Tenant domain", "", false)
	if err != nil {
		fmt.Println(err.Error())
	}

	//var newTenant *config.Tenant
	newTenant := &config.Tenant{
		Tenant: config.TenantProfile{
			Name:          tenantName,
			Domain:        domain,
			DefaultClient: getDefaultClient(),
			APIs:  getTenantAPIs(),
		},
	}

	// Save tenant to config
	tenantConfig.SetTenant(tenantName, newTenant)

	err = tenantConfig.SaveTenantConfig()
	if err != nil {
		fmt.Println("Failed to save tenant configuration: ", err)
		os.Exit(1)
	}
}

func getTenantAPIs()  []config.API{

	apilist := make([]config.API, 0)
	for {
		apiName, err := common.PromptString("Tenant API name: ", "", false)
		if err != nil {
			fmt.Println("Failed to get tenant API configuration: ", err)
			os.Exit(1)
		}
		apiAudience, err := common.PromptString("Tenant API audience: ", "", false)
		if err != nil {
			fmt.Println("Failed to get tenant API configuration: ", err)
			os.Exit(1)
		}
		apilist = append(apilist, config.API{Name: apiName, Audience: apiAudience})
		_, addAnotherAPI, err := common.PromptSelect("Do you have more APIs to add to this tenant?", []string{"Yes", "No"})
		if addAnotherAPI == "No" {
			break
		}
	}

	return apilist
}

func getDefaultClient() *config.Client {
	// Get configured clients
	clientConfig, err := config.LoadClientConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	if len(*clientConfig.GetClientList("all")) == 0 {
		return nil
	}

	// Wrap this is a user select y/n if they want to set a default client.- Give a short blurb on how this is used
	_, selectedClient, err := common.PromptSelect("Select a Default Client to use with this tenant ", config.GetClientListNames(*clientConfig.GetClientList("all")))
	return clientConfig.GetClientConfig(strings.ToLower(selectedClient))
}