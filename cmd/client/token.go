package client

import (
	"fmt"
	"github.com/SPSCommerce/spsauth0/common"
	"github.com/SPSCommerce/spsauth0/internal/config"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	clientTokenCmd = &cobra.Command{
		Use:     "token",
		Short:   "Get a token for a configured client",
		Long: "If getting a token for a native client you must have a redirect of" +
			" http://localhost:1000 configured via DevCenter for authtool to work.",
		Aliases: []string{"tk"},
		Args:    cobra.NoArgs,
		Run:     clientTokenExecute,
	}
)

func clientTokenExecute(cmd *cobra.Command, args []string) {
	// Get configured clients
	clientConfig, err := config.LoadClientConfigWithViper()
	if err != nil {
		fmt.Printf("Error: could not load tenant config - %v\n", err)
		os.Exit(1)
	}

	_, selectedClient, err := common.PromptSelect("Clients", config.GetClientListNames(*clientConfig.GetClientList("all")) )
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	// Check that a client does not already exist in the config with the same name
	client := clientConfig.GetClientConfig(strings.ToLower(selectedClient))
	if client == nil {
		fmt.Printf("Error getting client configuration")
		os.Exit(1)
	}

	fmt.Println(common.GetTokenHandler(client))
}
