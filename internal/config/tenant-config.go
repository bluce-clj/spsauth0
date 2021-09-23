package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
)


type Auth0Config struct {
	Auth0Domain string `json:"AUTH0_DOMAIN"`
	Auth0ClientSecret string `json:"AUTH0_CLIENT_SECRET"`
	Auth0ClientId string `json:"AUTH0_CLIENT_ID"`
}

// Rename this vales to not have the type client in them.
type Client struct {
	ClientId string
	ClientSecret string
	ClientName   string
	ClientType   string
	TenantName	string
	Token 		string
	Audience string
}

type API struct {
	Name     string
	Audience string
}

type TenantProfile struct {
	Name string
	Domain string
	APIs   []API
	DefaultClient *Client
}

// Remove this struct. No reason to have this be a wrapper around TenantProfile
type Tenant struct {
	Tenant   TenantProfile
}

// TenantConfig represents auth0 Tenants that spsauth0 stores
type TenantConfig struct {
	data map[string]*Tenant
	v    *viper.Viper
}

// TenantConfig represents auth0 Tenants that spsauth0 stores
type A0deployConfig struct {
	data map[string]*Auth0Config
	v    *viper.Viper
}

type TenantProfileList []*TenantProfile

// move to common file
func ensureTenantConfig(cfgFile string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			if writeErr := v.WriteConfig(); writeErr != nil {
				return nil, writeErr
			}
		} else {
			return nil, err
		}
	}
	return v, nil
}

// LoadTenantConfig ensures the tenant config file exists and then loads it.  Once
// loaded, it can be set and written to multiple times within the process.
// If changes are made directly to the file, though, they will likely be
// overwritten while the process is running
func LoadTenantConfig(cfgFile string) (*TenantConfig, error) {
	v, err := ensureTenantConfig(cfgFile)
	if err != nil {
		return nil, err
	}

	c, err := cacheTenantConfig(v)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Rename to cacheTenantConfig
func cacheTenantConfig(v *viper.Viper) (*TenantConfig, error) {
	c := TenantConfig{
		data: make(map[string]*Tenant),
		v:    v,
	}
	err := v.Unmarshal(&c.data)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// LoadTenantConfigWithViper sets the path to aws cred config using the viper
// config (i.e. --configdir) and then ensures the AWS config file exists and
// loads it.
func LoadTenantConfigWithViper() (*TenantConfig, error) {
	rootConfigDir, err := InitConfigDirWithViper()
	if err != nil {
		return nil, err
	}

	configFile := path.Join(rootConfigDir, TenantConfigFile)

	return LoadTenantConfig(configFile)
}

// GetTenantConfig returns the tenantConfig for the specified name or nil if the
// tenant does not exist or config has not been loaded.
func (t *TenantConfig) GetTenantConfig(tenantName string) *Tenant {
	tenant, ok := t.data[tenantName]
	if !ok {
		return nil
	}
	return tenant
}

// SetAWSProfile sets the profile in the local cache and the store
func (a *TenantConfig) SetTenant(profileName string, profile *Tenant) {
	a.data[profileName] = profile
	a.v.Set(profileName, *profile)
}

// SaveTenantConfig writes the config back to disk, capturing any profile and
// session changes
func (a *TenantConfig) SaveTenantConfig() error {
	return a.v.WriteConfig()
}

// GetTenantProfileList returns a list of all AWSProfilesItems in the config store,
// or nil if there was an error loading the config.
func (a *TenantConfig) GetTenantProfileList() *TenantProfileList {
	list := make(TenantProfileList, 0, len(a.data))
	for _, v := range a.data {
		item := &TenantProfile{
			Name: v.Tenant.Name,
			Domain: v.Tenant.Domain,
			DefaultClient: v.Tenant.DefaultClient,
		}
		list = append(list, item)
	}
	listPtr := &list

	return listPtr
}

func(a *TenantConfig) GetTenantListNames() []string {
	list := make([]string, 0, len(a.data))
	for _, v := range a.data {
		list = append(list, v.Tenant.Name)
	}

	return list
}

func(a *TenantConfig) GetTenantAPINames(tenantAPIs []API) []string {
	fmt.Print(tenantAPIs)
	list := make([]string, 0, len(tenantAPIs))
	for _, v := range tenantAPIs {
		list = append(list, v.Name)
	}

	return list
}

// LoadTenantConfigWithViper sets the path to aws cred config using the viper
// config (i.e. --configdir) and then ensures the AWS config file exists and
// loads it.
func LoadA0DeployConfigWithViper() (*A0deployConfig, error) {
	rootConfigDir, err := InitConfigDirWithViper()
	if err != nil {
		return nil, err
	}

	configFile := path.Join(rootConfigDir, Auth0DeployConfigFile)

	return LoadA0deployConfig(configFile)
}

func LoadA0deployConfig(cfgFile string) (*A0deployConfig, error) {
	v, err := ensureTenantConfig(cfgFile)
	if err != nil {
		return nil, err
	}

	c, err := cacheA0deployConfig(v)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Rename to cacheTenantConfig
func cacheA0deployConfig(v *viper.Viper) (*A0deployConfig, error) {
	c := A0deployConfig{
		data: make(map[string]*Auth0Config),
		v:    v,
	}
	err := v.Unmarshal(&c.data)
	if err != nil {
		return nil, err
	}
	return &c, nil
}


// GetTenantConfig returns the tenantConfig for the specified name or nil if the
// tenant does not exist or config has not been loaded.
func (a0 *A0deployConfig) GetA0deployConfig(tenantName string) *Auth0Config {
	a0Config, ok := a0.data[tenantName]
	if !ok {
		return nil
	}
	return a0Config
}

// SetAWSProfile sets the profile in the local cache and the store
func (a0 *A0deployConfig) SetA0deployConfig(profileName string, profile *Auth0Config) {
	a0.data[profileName] = profile
	a0.v.Set(profileName, *profile)
}