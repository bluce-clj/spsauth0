package tenant

import (
"fmt"
"os"

"spsauth0/internal/config"
"github.com/olekukonko/tablewriter"
"github.com/spf13/cobra"
)

var (
	tenantListCmd = &cobra.Command{
		Use:     "list",
		Short:   "List configured tenants",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Run:     tenantListExecute,
	}
)

func tenantListExecute(cmd *cobra.Command, args []string) {
	tenantConfig, err := config.LoadTenantConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	tenants := tenantConfig.GetTenantProfileList()
	rows := prepData(tenants)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Tenant Name", "Domain",
		"Default Client"})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetHeaderLine(false)
	table.AppendBulk(rows)
	table.Render()
}

func prepData(tenants *config.TenantProfileList) [][]string {
	rows := make([][]string, 0, len(*tenants))
	for _, tenant := range *tenants {
		name := tenant.Name
		domain := tenant.Domain
		defaultClient := prepClient(tenant.DefaultClient)
		rows = append(rows, []string{
			name,
			domain,
			defaultClient,
		})
	}
	return rows
}

func prepClient(client *config.Client) string {
	if client == nil {
		return "No client Configured"
	}
	return client.ClientName
}
