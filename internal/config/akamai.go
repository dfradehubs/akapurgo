package config

import (
	"akapurgo/api/v1alpha1"
	"akapurgo/internal/commons"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
)

func CreateAkamaiConfigFile(ctx v1alpha1.Context) error {

	// Check if the file already exists
	_, err := os.Stat(commons.AkamaiConfigPath)
	if err == nil {
		// The file already exists, nothing to do
		fmt.Println("The configuration file already exists.")
		return nil
	}

	// If the file doesn't exist, create it
	cfg := ini.Empty()

	// Create the [default] section
	section, err := cfg.NewSection("default")
	if err != nil {
		return fmt.Errorf("could not create default section: %v", err)
	}

	// Set values for the [default] section
	section.Key("host").SetValue(ctx.Config.Akamai.Host)
	section.Key("client_secret").SetValue(ctx.Config.Akamai.ClientSecret)
	section.Key("client_token").SetValue(ctx.Config.Akamai.ClientToken)
	section.Key("access_token").SetValue(ctx.Config.Akamai.AccessToken)

	// Create the file
	err = cfg.SaveTo(commons.AkamaiConfigPath)
	if err != nil {
		return fmt.Errorf("could not create the configuration file: %v", err)
	}

	fmt.Println("Akamai configuration file created successfully.")
	return nil
}
