package config

type GRPCConfig struct {
	Port string
}

func (c GRPCConfig) Addr() string {
	return ":" + c.Port
}
