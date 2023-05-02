package config

import "github.com/spf13/viper"

type Config struct {
	TokenUrl        string
	ClientId        string
	ClientSecret    string
	GrantType       string
	ParticipantsUrl string
	PaymentUrl      string
	ProductionUrl   string
	RedisDbName     string
	RedisServer     string
	RedisPassword   string
}

func GetConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	config := Config{
		TokenUrl:        viper.GetString("token_url"),
		ClientId:        viper.GetString("client_id"),
		ClientSecret:    viper.GetString("client_secret"),
		GrantType:       viper.GetString("grant_type"),
		ParticipantsUrl: viper.GetString("participants_url"),
		PaymentUrl:      viper.GetString("payment_url"),
		RedisDbName:     viper.GetString("redis_dbname"),
		RedisServer:     viper.GetString("redis_server"),
		RedisPassword:   viper.GetString("redis_password"),
	}
	return config
}
