// config/config.go

package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    DBHost     string `mapstructure:"DB_HOST"`
    DBPort     int    `mapstructure:"DB_PORT"`
    DBUser     string `mapstructure:"DB_USER"`
    DBPassword string `mapstructure:"DB_PASSWORD"`
    DBName     string `mapstructure:"DB_NAME"`
}

var AppConfig Config

func LoadConfig() error {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        return err
    }

    if err := viper.Unmarshal(&AppConfig); err != nil {
        return err
    }

    return nil
}