package tenant

import (
	"fmt"
	"github.com/bluce-clj/spsauth0/common"
	"github.com/bluce-clj/spsauth0/internal/config"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (

	tenantUpdateCmd = &cobra.Command{
		Use:   "update <tenant name>",
		Short: "update a configured tenant",
		Long:  "to do",
		Args:  cobra.ExactArgs(1),
		Run:   tenantUpdateExecute,
	}
)

//func init() {
//}

func tenantUpdateExecute(cmd *cobra.Command, args []string)  {
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

	domain, err := common.PromptString("Tenant domain", tenant.Tenant.Domain, false)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
	}


	newTenant := &config.Tenant{
		Tenant: config.TenantProfile{
			Name:          tenantName,
			Domain:        domain,
			APIs: tenant.Tenant.APIs,
			DefaultClient: getDefaultClient(),
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
