package database

import (
	"encoding/json"

	"github.com/ricardovano/qpay/internal/entity"
)

func CreateLog(log entity.Log) entity.Log {

	jsonData, err := json.Marshal(log)
	if err != nil {
		panic(err)
	}

	data := string(jsonData)
	err = setData(log.TransactionDate.GoString(), data, 1)
	if err != nil {
		panic(err)
	}

	return log
}

func GetLog(id string) entity.Log {
	data := getData(id, 1)

	var log entity.Log
	err := json.Unmarshal([]byte(data), &log)
	if err != nil {
		panic(err)
	}
	return log
}

func GetAllLogs() []entity.Log {
	data := getAll(1)
	var log entity.Log
	var logs []entity.Log
	for _, s := range data {
		log = GetLog(s)
		logs = append(logs, log)
	}
	return logs
}
