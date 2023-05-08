package entity

import "time"

type Log struct {
	Beneficiary     Beneficiary
	Payer           Payer
	Amount          string
	TransactionDate time.Time
	Error           string
	Function        string
	Status          string
}

type LogDTO struct {
	Logs          []Log
	Beneficiaries []Beneficiary
}
