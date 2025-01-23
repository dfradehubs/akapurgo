package v1alpha1

// Configuration struct
type ConfigSpec struct {
	Server struct {
		ListenAddress string `yaml:"listen_address"`
		Config        struct {
			ReadBufferSize int `yaml:"read_buffer_size"`
		} `yaml:"config"`
	} `yaml:"server"`
	Akamai struct {
		Host         string `yaml:"host"`
		ClientSecret string `yaml:"client_secret"`
		ClientToken  string `yaml:"client_token"`
		AccessToken  string `yaml:"access_token"`
	} `yaml:"akamai"`
	Logs struct {
		ShowAccessLogs bool `yaml:"show_access_logs"`
		JwtUser        struct {
			Enabled  bool   `yaml:"enabled"`
			Header   string `yaml:"header"`
			JwtField string `yaml:"jwt_field"`
		} `yaml:"jwt_user"`
		AccessLogsFields []string `yaml:"access_logs_fields"`
	} `yaml:"logs"`
}
