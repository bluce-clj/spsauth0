/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package client

import (
	"fmt"
	"github.com/bluce-clj/spsauth0/internal/config"
	"github.com/spf13/viper"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var ClientCmd = &cobra.Command{
	Use:   "client",
	Short: "to do",
	Long: `to do`,
}

func init() {
	cobra.OnInitialize(InitRootConfig)

	ClientCmd.AddCommand(clientAddCmd)
	ClientCmd.AddCommand(clientListCmd)
	ClientCmd.AddCommand(clientTokenCmd)
}

// InitRootConfig initializes and loads the config for auth0 clients
func InitRootConfig() {
	// Initialize the config dir, here.
	cfgDir := viper.GetString(config.KeyRootCmdConfigDir)
	config.SyncInitConfigDir(cfgDir)
	errInit := config.SyncInitConfigDirErr()
	if errInit != nil {
		fmt.Printf("Error with config dir: %s: %v\n", cfgDir, errInit)
		os.Exit(1)
	}

	// Set up an client config file in the config dir
	cfgFile := path.Join(config.SyncGetConfigDir(), config.ClientConfigFile)

	_, err := config.LoadTenantConfig(cfgFile)
	if err != nil {
		fmt.Printf("Could not load aws config: %s: %v\n", cfgFile, err)
		os.Exit(1)
	}
}

