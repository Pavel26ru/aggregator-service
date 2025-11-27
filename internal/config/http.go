package config

type HTTPConfig struct {
	Port string
}

func (c HTTPConfig) Addr() string {
	return ":" + c.Port
}
