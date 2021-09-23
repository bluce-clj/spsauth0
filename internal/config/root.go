package config

import (
	"fmt"
	"os"
	"sync"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

//TODO find a way to remove this.  Intention of sync went a little sideways

var (
	initRootOnce sync.Once
	configDir    string
	errInitRoot  error = fmt.Errorf("not initialized")
)

// SyncInitConfigDir initializes the config directory for spsauth0
func SyncInitConfigDir(cfgDir string) {
	initRootOnce.Do(func() {
		configDir, errInitRoot = InitConfigDir(cfgDir)
	})
}

// SyncInitConfigDirErr reports if there was an error initializing the config dir,
// or if initialization has not yet occurred.
func SyncInitConfigDirErr() error {
	return errInitRoot
}

// SyncGetConfigDir gets the config dir synchronously, if loaded.
func SyncGetConfigDir() string {
	if errInitRoot != nil {
		return ""
	}
	return configDir
}

// InitConfigDir ensures the spsauth0 config dir has been created
func InitConfigDir(cfgDir string) (string, error) {
	fullCfgDir, err := homedir.Expand(cfgDir)
	if err != nil {
		return "", err
	}

	fi, err := os.Stat(fullCfgDir)
	if err != nil {
		if os.IsNotExist(err) {
			if mkDirErr := os.MkdirAll(fullCfgDir, 0700); mkDirErr != nil {
				return "", mkDirErr
			}
			fi, err = os.Stat(fullCfgDir)
		}
	}

	// If err from 1st os.Stat() was not IsNotExist or err from 2nd os.Stat() was
	// set, then fail
	if err != nil {
		return "", err
	}

	// Double check that the config dir is, actually, a dir
	if !fi.IsDir() {
		errInitRoot = fmt.Errorf("config dir %s is not a directory", fullCfgDir)
	}

	return fullCfgDir, nil
}

// InitConfigDirWithViper ensures the spsauth0 config dir has been created
// and uses viper to load the path to the spsauth0 dir
func InitConfigDirWithViper() (string, error) {
	cfgDir := viper.GetString(KeyRootCmdConfigDir)
	return InitConfigDir(cfgDir)
}
