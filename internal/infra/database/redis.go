package database

import (
	"fmt"

	"github.com/go-redis/redis"
)

const (
	dbPass    = ""
	dbName    = "apis"
	dbAddress = "localhost:6379"
)

func getClient() *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     dbAddress,
		Password: dbPass,
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
