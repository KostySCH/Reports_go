package config

type MinioConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
}

func Load() *MinioConfig {
	return &MinioConfig{
		Endpoint:  "localhost:9001",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		UseSSL:    false,
	}
}
