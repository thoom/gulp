package main

// GulpConfig contains configuration data
type GulpConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Display string            `json:"display"`
	Flags   map[string]string `json:"flags"`
}

// NewConfig creates a new instance of GulpConfig
func NewConfig() GulpConfig {
	flags := make(map[string]string)
	flags["use_color"] = "true"
	flags["verify_tls"] = "true"

	return GulpConfig{Flags: flags}
}

// UseColor adds a switch for whether or not to colorize the output
func (gc GulpConfig) UseColor() bool {
	return gc.Flags["use_color"] != "false"
}

// TLSVerify determines whether or not to verify that a TLS cert is valid
func (gc GulpConfig) TLSVerify() bool {
	return gc.Flags["verify_tls"] != "false"
}
