package config

import (
	"github.com/spf13/viper"
	"path"
	"strings"
)

type ClientList []*Client


type ClientConfig struct {
	data map[string]*Client
	v    *viper.Viper}


// LoadClientConfigWithViper sets the path to aws cred config using the viper
// config (i.e. --configdir) and then ensures the AWS config file exists and
// loads it.
func LoadClientConfigWithViper() (*ClientConfig, error) {
	rootConfigDir, err := InitConfigDirWithViper()
	if err != nil {
		return nil, err
	}

	configFile := path.Join(rootConfigDir, ClientConfigFile)

	return LoadClientConfig(configFile)
}

// LoadClientConfig  ensures the tenant config file exists and then loads it.  Once
// loaded, it can be set and written to multiple times within the process.
// If changes are made directly to the file, though, they will likely be
// overwritten while the process is running
func LoadClientConfig(cfgFile string) (*ClientConfig, error) {
	v, err := ensureTenantConfig(cfgFile)
	if err != nil {
		return nil, err
	}

	c, err := cacheClientConfig(v)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func cacheClientConfig(v *viper.Viper) (*ClientConfig, error) {
	c := ClientConfig{
		data: make(map[string]*Client),
		v:    v,
	}
	err := v.Unmarshal(&c.data)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func GetSupportedClientTypes() []string {
	return []string{"Web Service Application", "Native Application", "Single-Page Application (SPA)",
		"Machine-to-Machine Application"}
}

// GetTenantConfig returns the tenantConfig for the specified name or nil if the
// tenant does not exist or config has not been loaded.
func (c *ClientConfig) GetClientConfig(clientName string) *Client {
	client, ok := c.data[clientName]
	if !ok {
		return nil
	}
	return client
}

// SetAWSProfile sets the profile in the local cache and the store
func (c *ClientConfig) SetClient(client *Client) {
	c.v.Set(client.ClientName, client)
}

// SaveTenantConfig writes the config back to disk, capturing any profile and
// session changes
func (c *ClientConfig) SaveClientConfig() error {
	return c.v.WriteConfig()
}

// GetTenantProfileList returns a list of all AWSProfilesItems in the config store,
// or nil if there was an error loading the config.
func (c *ClientConfig) GetClientList(tenantName string) *ClientList {
	list := make(ClientList, 0, len(c.data))
	for _, v := range c.data {
		if strings.ToLower(v.TenantName) == tenantName || tenantName == "all"{
			item := &Client{
				ClientName: v.ClientName,
				ClientId: v.ClientId,
				ClientSecret: v.ClientSecret,
			}
			list = append(list, item)
		}
	}

	return &list
}

func GetClientListNames(clients ClientList) []string {
	list := make([]string, 0, len(ClientList{}))
	for _, v := range clients {
		list = append(list, v.ClientName)
	}

	return list
}

//func GetConfiguredClient(clients *ClientList, clientConfig ClientConfig)  *Client{
//	if len(*clients) == 0 {
//		return nil
//	}
//	_, selectedClient, err := common.PromptSelect("Clients", GetClientListNames(*clients) )
//	if err != nil {
//		fmt.Printf(err.Error())
//		os.Exit(1)
//	}
//	return clientConfig.GetClientConfig(selectedClient)
//}