package client

import (
	"fmt"
	"spsauth0/common"
	"os"

	"spsauth0/internal/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	clientListCmd = &cobra.Command{
		Use:     "list",
		Short:   "List configured client",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Run:     clientListExecute,
	}
)

func clientListExecute(cmd *cobra.Command, args []string) {
	clientConfig, err := config.LoadClientConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	// If list of clients is 0 error now

	tenantToSearch := getTenantToSearch()
	clients := clientConfig.GetClientList(tenantToSearch)

	rows := prepData(clients)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Client Name", "Client ID",
		"Client Secret"})
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

func prepData(clients *config.ClientList) [][]string {
	rows := make([][]string, 0, len(*clients))
	for _, client := range *clients {
		clientName := client.ClientName
		clientId := client.ClientId
		clientSecret := client.ClientSecret
		rows = append(rows, []string{
			clientName,
			clientId,
			clientSecret,
		})
	}
	return rows
}

func getTenantToSearch() string{

	tenantConfig, err := config.LoadTenantConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	tenantNames := tenantConfig.GetTenantListNames()
	if len(tenantNames) > 1 {
		_, tenantName, err := common.PromptSelect("Which tenant do you want to list clients for", tenantNames)
		if err != nil {
			fmt.Println("Err")
			os.Exit(1)
		}
		return tenantName
	}

	return tenantNames[0]

}