package tenant

import (
	"github.com/bluce-clj/spsauth0/internal/config"
	"github.com/spf13/viper"
)


func getTenanteArg(args []string) string {
	if len(args) != 1 {
		panic("Thought that profile was set as first arg...")
	}
	viper.Set(config.KeyCmdTenantName, args[0])
	return args[0]
}

func getTenantSearchArg(args []string) string {
	if len(args) != 1 {
		panic("must provide an argument as to what you want to search for.")
	}

	if args[0] != "clients" {
		panic("Currently tenant search only supports searching for clients")
	}
	return args[0]
}

