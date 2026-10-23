package config

type Config struct {
	Meta struct {
		Version     string `yaml:"version"`
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	} `yaml:"meta"`
	Location Location `yaml:"location"`
	Lights   []Light  `yaml:"lights"`
}

type Location struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
}

type Light struct {
	ID   *string `yaml:"id"`
	Name *string `yaml:"name"`
}

// Clone returns a deep copy of the config, so it can be handed to another
// goroutine without sharing the lights.
func (c *Config) Clone() *Config {
	clone := *c
	clone.Lights = make([]Light, len(c.Lights))
	for i, light := range c.Lights {
		clone.Lights[i] = Light{ID: cloneString(light.ID), Name: cloneString(light.Name)}
	}
	return &clone
}

func cloneString(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}
