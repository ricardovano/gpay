package database

import (
	"encoding/json"

	"github.com/ricardovano/qpay/internal/entity"
)

func CreateBeneficiary(beneficiary entity.Beneficiary) string {

	jsonData, err := json.Marshal(beneficiary)
	if err != nil {
		panic(err)
	}

	data := string(jsonData)
	err = setData(beneficiary.Code, data)
	if err != nil {
		panic(err)
	}

	return beneficiary.Code
}

func GetBeneficiary(id string) entity.Beneficiary {
	data := getData(id)

	var beneficiary entity.Beneficiary
	err := json.Unmarshal([]byte(data), &beneficiary)
	if err != nil {
		panic(err)
	}
	return beneficiary
}
