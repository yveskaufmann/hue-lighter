package config

type Config struct {
	Meta struct {
		Version     string `yaml:"version"`
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	} `yaml:"meta"`
	Location struct {
		Latitude  float64 `yaml:"latitude"`
		Longitude float64 `yaml:"longitude"`
	} `yaml:"location"`
	Lights []struct {
		ID   *string `yaml:"id"`
		Name *string `yaml:"name"`
	} `yaml:"lights"`
}
