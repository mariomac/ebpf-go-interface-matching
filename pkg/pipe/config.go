package pipe

type Config struct {
	Exec     string `env:"EXECUTABLE_NAME"`
	FuncName string `env:"INSTRUMENT_FUNC_NAME" envDefault:"main.DispatchMessage"`
}

type ConfigError string

func (e ConfigError) Error() string {
	return string(e)
}

func (c *Config) Validate() error {
	if c.Exec == "" {
		return ConfigError("missing EXECUTABLE_NAME property")
	}
	if c.FuncName == "" {
		return ConfigError("missing INSTRUMENT_FUNC_NAME property")
	}
	return nil
}
