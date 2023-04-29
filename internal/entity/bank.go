package entity

type Bank struct {
	Name string
	Code string
	ISPB string
}

func GetBanks() []Bank {
	banks := []Bank{
		{Name: "Banco Bradesco", Code: "237", ISPB: "60746948"},
		{Name: "Banco BTG Pactual", Code: "208", ISPB: "30306294"},
		{Name: "Banco Santander", Code: "33", ISPB: "90400888"},
		{Name: "Banco Daycoval", Code: "707", ISPB: "62232889"},
		{Name: "Banco do Brasil", Code: "1", ISPB: "0"},
		{Name: "Caixa Economica Federal", Code: "104", ISPB: "360305"},
		{Name: "Itaú Unibanco", Code: "341", ISPB: "60701190"},
		{Name: "Nubank", Code: "260", ISPB: "18236120"},
	}
	return banks
}
