package v1alpha1

// Configuration struct
type ConfigSpec struct {
	Server struct {
		ListenAddress string `yaml:"listen_address"`
	} `yaml:"server"`
	Akamai struct {
		Host         string `yaml:"host"`
		ClientSecret string `yaml:"client_secret"`
		ClientToken  string `yaml:"client_token"`
		AccessToken  string `yaml:"access_token"`
	} `yaml:"akamai"`
	Logs struct {
		ShowAccessLogs   bool     `yaml:"show_access_logs"`
		AccessLogsFields []string `yaml:"access_logs_fields"`
	} `yaml:"logs"`
}
