package tenant

import (
	"encoding/json"
	"fmt"
	"github.com/SPSCommerce/spsauth0/common"
	"github.com/SPSCommerce/spsauth0/internal/config"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

var (

	tenantExportCmd = &cobra.Command{
		Use:   "export <tenant name>",
		Short: "Export a tenants configuration",
		Long:  "Export a tenants configuration",
		Args:  cobra.ExactArgs(1),
		Run:   tenantExportExecute,
	}
)

// Write config file using viper
// output string of a0deploy cmd to use pointing to generated config file and pointing to tmp folder

func init() {
	//awsCredAddCmd.Flags().String(config.FlagCmdAWSCloneProfile,
	//	"", "Clone an existing profile")
	//viper.BindPFlag(config.KeyCmdAWSCloneProfile,
	//	awsCredAddCmd.Flags().Lookup(config.FlagCmdAWSCloneProfile))
}

func tenantExportExecute(cmd *cobra.Command, args []string)  {

	tenantName := getTenanteArg(args)
	tenantConfig, err := config.LoadTenantConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	tenant := tenantConfig.GetTenantConfig(strings.ToLower(tenantName))
	if tenant == nil {
		fmt.Printf("Tenant %s does not exist, use 'spsauth0 tenant list' to list configured tenants.", tenantName)
		os.Exit(1)
	}

	//a0deployConfig, err := config.LoadA0DeployConfigWithViper()
	//if err != nil {
	//	fmt.Printf("Error: could not load tenant config - %v\n", err)
	//	os.Exit(1)
	//}

	client := getClientForExport(tenant)

	cfg := &config.Auth0Config{
		Auth0Domain: tenant.Tenant.Domain,
		Auth0ClientId: client.ClientId,
		Auth0ClientSecret: client.ClientSecret,
	}
	cfgDir := viper.GetString(config.KeyRootCmdConfigDir)
	fullCfgDir, err := homedir.Expand(cfgDir)
	fmt.Printf(fullCfgDir)

	data, err := json.Marshal(cfg)

	//a0deployConfig.
	//
	//dir, err := ioutil.TempDir(os.TempDir(), "prefix-")
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//defer os.RemoveAll(dir)
	a0fileName := viper.GetString(config.Auth0DeployConfigFile)
	fmt.Printf("here")

	ioutil.WriteFile(fmt.Sprint(fullCfgDir+"/"+a0fileName), data, 0644)

}

func getClientForExport(tenant *config.Tenant) *config.Client {

	if tenant.Tenant.DefaultClient != nil {
		_, useDefaultClient, _ := common.PromptSelect(fmt.Sprintf("Do you want to use the defaultClient %s set on the tenant?", tenant.Tenant.DefaultClient.ClientName), []string{"Yes", "No"})
		if useDefaultClient == "Yes"{
			return tenant.Tenant.DefaultClient
		}
	}

	// Get configured clients
	clientConfig, err := config.LoadClientConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load client config - %v\n", err)
		os.Exit(1)
	}

	if len(*clientConfig.GetClientList(tenant.Tenant.Name)) == 0 {
		fmt.Printf("You have no configured clients for the %s tenant", tenant.Tenant.Name)
		os.Exit(1)
	}

	// Wrap this is a user select y/n if they want to set a default client.- Give a short blurb on how this is used
	_, selectedClient, err := common.PromptSelect("Select a Default Client to use with this tenant ", config.GetClientListNames(*clientConfig.GetClientList(tenant.Tenant.Name)))
	return clientConfig.GetClientConfig(strings.ToLower(selectedClient))
}
