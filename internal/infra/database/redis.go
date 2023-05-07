package database

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/ricardovano/qpay/config"
)

func getClient(db int) *redis.Client {

	config := config.GetConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisServer,
		Password: config.RedisPassword,
		DB:       db,
	})
	return client
}

func getData(id string, db int) string {

	client := getClient(db)
	val, err := client.Get(id).Result()
	if err != nil {
		fmt.Println(err)
	}
	return string(val)
}

func setData(id string, obj string, db int) error {
	client := getClient(db)
	err := client.Set(id, obj, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func getAll(db int) []string {
	client := getClient(db)
	val, err := client.Keys("*").Result()
	if err != nil {
		fmt.Println(err)
	}
	return val
}
