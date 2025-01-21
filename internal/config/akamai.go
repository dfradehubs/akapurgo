package config

import (
	"akapurgo/api/v1alpha1"
	"akapurgo/internal/commons"
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

// CreateAkamaiConfigFile creates the Akamai configuration file
func CreateAkamaiConfigFile(ctx v1alpha1.Context) error {
	// Check if the configuration file already exists
	_, err := os.Stat(commons.AkamaiConfigPath)
	if err == nil {
		// File already exists
		ctx.Logger.Info("The configuration file already exists.")
		return nil
	}

	// Create a new configuration file
	cfg := ini.Empty()

	// Create the default section
	section, err := cfg.NewSection("default")
	if err != nil {
		return fmt.Errorf("could not create default section: %v", err)
	}

	// Assign the values to the keys
	section.Key("host").SetValue(ctx.Config.Akamai.Host)
	section.Key("client_secret").SetValue(ctx.Config.Akamai.ClientSecret)
	section.Key("client_token").SetValue(ctx.Config.Akamai.ClientToken)
	section.Key("access_token").SetValue(ctx.Config.Akamai.AccessToken)

	// Save the configuration to the file
	err = cfg.SaveTo(commons.AkamaiConfigPath)
	if err != nil {
		return fmt.Errorf("could not create the configuration file: %v", err)
	}

	ctx.Logger.Info("Akamai configuration file created successfully.")
	return nil
}
