package config

import (
	"gopkg.in/yaml.v3"
	"os"

	"akapurgo/api/v1alpha1"
)

// Unmarshal TODO
func Unmarshal(bytes []byte) (config v1alpha1.ConfigSpec, err error) {
	err = yaml.Unmarshal(bytes, &config)
	return config, err
}

// ReadFile TODO
func ReadFile(filepath string) (config v1alpha1.ConfigSpec, err error) {
	var fileBytes []byte
	fileBytes, err = os.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	// Expand environment variables present in the config
	// This will cause expansion in the following way: field: "$FIELD" -> field: "value_of_field"
	fileExpandedEnv := os.ExpandEnv(string(fileBytes))
	config, err = Unmarshal([]byte(fileExpandedEnv))

	return config, err
}
