package entity

type PaymentRequest struct {
	Type                  string      `json:"type"`
	Payer                 Payer       `json:"payer"`
	Beneficiary           Beneficiary `json:"beneficiary"`
	Amount                string      `json:"amount"`
	ReturnUri             string      `json:"returnUri"`
	WebhookUrl            string      `json:"webhookUrl"`
	ReferenceCode         string      `json:"referenceCode"`
	TermsOfUseVersion     string      `json:"termsOfUseVersion"`
	TermsOfPrivacyVersion string      `json:"termsOfPrivacyVersion"`
}

type PaymentResponse struct {
	Type                  string      `json:"type"`
	PaymentId             string      `json:"paymentId"`
	Amount                string      `json:"amount"`
	InitiatedAt           string      `json:"initiatedAt"`
	UpdatedAt             string      `json:"updatedAt"`
	Status                string      `json:"status"`
	Payer                 Payer       `json:"payer"`
	Beneficiary           Beneficiary `json:"beneficiary"`
	ReturnUri             string      `json:"returnUri"`
	WebhookUrl            string      `json:"webhookUrl"`
	ReferenceCode         string      `json:"referenceCode"`
	TermsOfUseVersion     string      `json:"termsOfUseVersion"`
	TermsOfPrivacyVersion string      `json:"termsOfPrivacyVersion"`
	AuthenticationUri     string      `json:"authenticationUri"`
	EndToEndId            string      `json:"endToEndId"`
	TransactionId         string      `json:"transactionId"`
}

type Payer struct {
	Name          string `json:"name"`
	CPF           string `json:"cpf"`
	Email         string `json:"email"`
	ParticipantId string `json:"participantId"`
}

type Beneficiary struct {
	LocalInstrument string `json:"localInstrument"`
	CPFCNPJ         string `json:"cpfCnpj"`
	Name            string `json:"name"`
	ISPB            string `json:"ispb"`
	Issuer          string `json:"issuer"`
	Number          string `json:"number"`
	AccountType     string `json:"accountType"`
	Code            string `json:"code"`
	Email           string `json:"email"`
}

type Participant struct {
	OrganisationName     string `json:"organisationName"`
	AuthorisationServers []AuthorisationServer
}

type AuthorisationServer struct {
	Id                          string `json:"id"`
	CustomerFriendlyName        string `json:"customerFriendlyName"`
	CustomerFriendlyDescription string `json:"customerFriendlyDescription"`
	CustomerFriendlyLogoUri     string `json:"customerFriendlyLogoUri"`
	Status                      string `json:"status"`
}

type StatusResponse struct {
	Success       string `json:"success"`
	PaymentId     string `json:"paymentId"`
	ReferenceCode string `json:"referenceCode"`
	Banks         []Bank `json:"banks"`
}

type AuthorisationServers struct {
	Data []AuthorisationServer `json:"data"`
}

type PaymentDTO struct {
	Data        []AuthorisationServer `json:"data"`
	Beneficiary Beneficiary           `json:"beneficiary"`
}

type Bank struct {
	Name string
	Code string
	ISPB string
}
