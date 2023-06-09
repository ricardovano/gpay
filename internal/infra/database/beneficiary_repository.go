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
	err = setData(beneficiary.Code, data, 0)
	if err != nil {
		panic(err)
	}

	return beneficiary.Code
}

func GetBeneficiary(id string) entity.Beneficiary {
	data := getData(id, 0)

	var beneficiary entity.Beneficiary
	err := json.Unmarshal([]byte(data), &beneficiary)
	if err != nil {
		panic(err)
	}
	return beneficiary
}

func GetAllBeneficiaries() []entity.Beneficiary {
	data := getAll(0)
	var beneficiary entity.Beneficiary
	var beneficiaries []entity.Beneficiary
	for _, s := range data {
		beneficiary = GetBeneficiary(s)
		beneficiaries = append(beneficiaries, beneficiary)
	}
	return beneficiaries
}
