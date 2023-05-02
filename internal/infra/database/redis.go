package database

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/ricardovano/qpay/config"
)

func getClient() *redis.Client {

	config := config.GetConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisServer,
		Password: config.RedisPassword,
		DB:       0,
	})
	return client
}

func getData(id string) string {

	client := getClient()
	val, err := client.Get(id).Result()
	if err != nil {
		fmt.Println(err)
	}
	return string(val)
}

func setData(id string, obj string) error {
	client := getClient()
	err := client.Set(id, obj, 0).Err()
	if err != nil {
		return err
	}

	return nil
}
