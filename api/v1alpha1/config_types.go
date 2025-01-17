package v1alpha1

// Configuration struct
type ConfigSpec struct {
	Server struct {
		ListenAddress string `yaml:"listenAddress"`
	} `yaml:"server"`
	Akamai struct {
		Host         string `yaml:"host"`
		ClientSecret string `yaml:"clientSecret"`
		ClientToken  string `yaml:"clientToken"`
		AccessToken  string `yaml:"accessToken"`
	} `yaml:"akamai"`
}
