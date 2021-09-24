package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port      int    `mapstructure:"PORT"`
	Env       string `mapstructure:"ENV"`
	DbURI     string `mapstructure:"DB_URI"`
	JwtSecret string `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	SMTP      struct {
		Host     string `mapstructure:"SMTP_HOST"`
		Port     int    `mapstructure:"SMTP_PORT"`
		Username string `mapstructure:"SMTP_USERNAME"`
		Password string `mapstructure:"SMTP_PASSWORD"`
		Sender   string `mapstructure:"SMTP_SENDER"`
	}
	CORS struct {
		TrustedOrigins []string `mapstructure:"CORS_TRUSTED_ORIGINS"`
	}
}

func LoadConfig(path string, name string) (cfg Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	err = viper.Unmarshal(&cfg.SMTP)
	err = viper.Unmarshal(&cfg.CORS)
	return
}
