package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ricardovano/qpay/internal/entity"
)

func main() {
	http.HandleFunc("/", participantsHandler)
	http.HandleFunc("/pay", payHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/register", registerHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/global.css", fs)
	http.Handle("/script.js", fs)
	http.Handle("/cafe1.png", fs)
	http.Handle("/cafe2.png", fs)
	http.Handle("/cafe3.png", fs)
	http.Handle("/cafe4.png", fs)
	http.ListenAndServe(":8080", nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Register beneficiary
	var beneficiary entity.Beneficiary
	beneficiary.LocalInstrument = "MANU"
	beneficiary.CPFCNPJ = FormatCPF(r.FormValue("cpf"))
	beneficiary.Name = r.FormValue("name")
	beneficiary.ISPB = r.FormValue("ispb")
	beneficiary.Issuer = r.FormValue("issuer")
	beneficiary.Number = r.FormValue("number")
	beneficiary.AccountType = "CACC"
	beneficiary.Email = r.FormValue("email")

	//generate
	beneficiary.Code = "ri5vano7rh1"

	//SAVE ON DATABASE

	//SEND EMAIL WITH LINK TO THE FRIEND (FULL URL)

	//REDIRECT TO REGISTER COMPLETE SITE

}

func payHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	token := getToken()

	payment := entity.PaymentRequest{
		Type: "PIX",
		Payer: entity.Payer{
			Name:          r.FormValue("name"),
			CPF:           FormatCPF(r.FormValue("cpf")),
			Email:         r.FormValue("email"),
			ParticipantId: r.FormValue("bank"),
		},
		Beneficiary:           getBeneficiary(r.FormValue("code")),
		Amount:                FormatMoney(r.FormValue("amount")),
		ReturnUri:             "http://localhost:8080/status",
		WebhookUrl:            "http://localhost:8080/webhook",
		ReferenceCode:         "be5ec8c9-6974-4830-94e3-363d1cfeb975",
		TermsOfUseVersion:     "1",
		TermsOfPrivacyVersion: "1",
	}

	response := postPayment(payment, token)

	response.AuthenticationUri, err = ReplaceRedirectUri(response.AuthenticationUri)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, response.AuthenticationUri, http.StatusSeeOther)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {

	var statusResponse entity.StatusResponse

	if r.URL.Query().Get("success") == "true" {
		statusResponse.Success = "Concluída com sucesso!"
	}
	statusResponse.PaymentId = r.URL.Query().Get("paymentId")
	statusResponse.ReferenceCode = r.URL.Query().Get("referenceCode")
	statusResponse.Banks = getBanks()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseFiles(wd + "/static/payment_status.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, statusResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println("JSON data:", string(body))

}

func participantsHandler(w http.ResponseWriter, r *http.Request) {
	token := getToken()
	participants, err := getParticipants(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseFiles(wd + "/static/payment_list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	paymentDTO := entity.PaymentDTO{}
	paymentDTO.Data = participants.Data
	paymentDTO.Beneficiary = getBeneficiary(r.URL.Query().Get("code"))

	err = tmpl.Execute(w, paymentDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postPayment(payment entity.PaymentRequest, token string) entity.PaymentResponse {
	url := "https://api-quanto.com/opb-api/v1/payments"

	paymentBytes, err := json.Marshal(payment)
	if err != nil {
		panic(err)
	}

	payload := strings.NewReader(string(paymentBytes))
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	oauth := "Bearer " + token
	req.Header.Add("authorization", oauth)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	var response entity.PaymentResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		fmt.Println("Failed to decode JSON:", err)
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println("JSON data:", string(body))
	}
	return response
}

func getParticipants(token string) (*entity.AuthorisationServers, error) {
	url := "https://api-quanto.com/opb-api/v1/participants/payments"
	req, err := http.NewRequest("GET", url, nil)

	oauth := "Bearer " + token
	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", oauth)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payments: %v", err)
	}
	defer resp.Body.Close()

	var participants []entity.Participant
	if err := json.NewDecoder(resp.Body).Decode(&participants); err != nil {
		fmt.Println("Failed to decode JSON:", err)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("JSON data:", string(body))
	}

	var data []entity.AuthorisationServer

	for i, p := range participants {
		for j, a := range p.AuthorisationServers {
			if a.CustomerFriendlyName == "Nubank" {
				data = append(data, a)
			}
			if a.CustomerFriendlyName == "Mercado Pago" {
				data = append(data, a)
			}
			j++
		}
		i++
	}

	var authorizationServers entity.AuthorisationServers
	authorizationServers.Data = data
	return &authorizationServers, nil
}

func getToken() string {
	url := "https://api-quanto.com/v1/api/token"
	payload := strings.NewReader("client_id=6abc0c4b-9ea7-41ce-b4ce-7465a4db20d5&client_secret=0cAHIMTwrj0pW3TdyNBctTiaaSpufRnh&grant_type=client_credentials")
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	var tokenObj entity.Token
	json.NewDecoder(res.Body).Decode(&tokenObj)

	fmt.Println("getting token...")
	return tokenObj.AccessToken
}

func getBeneficiary(beneficiaryCode string) entity.Beneficiary {

	//getById(beneficiaryCode) on database
	//Code = ri5vano7rh1

	beneficiary := entity.Beneficiary{
		LocalInstrument: "MANU",
		CPFCNPJ:         "28047925873",
		Name:            "Ricardo Vano",
		ISPB:            "60701190",
		Issuer:          "6477",
		Number:          "142035",
		AccountType:     "CACC",
		Code:            "ri5vano7rh1",
	}
	return beneficiary
}

func ReplaceRedirectUri(inputStr string) (string, error) {

	inputURL, err := url.Parse(inputStr)
	if err != nil {
		fmt.Println("Invalid input string")
		return "", err
	}

	queryParams, err := url.ParseQuery(inputURL.RawQuery)
	if err != nil {
		fmt.Println("Invalid input string")
		return "", err
	}

	// replace the redirect_uri parameter with the new value
	queryParams.Set("redirect_uri", "http://localhost.com:8080/status")

	// construct the output URL
	outputURL := inputURL.Scheme + "://" + inputURL.Host + inputURL.Path + "?" + queryParams.Encode()

	// replace any occurrences of %2F with /
	outputURL = strings.ReplaceAll(outputURL, "%2F", "/")

	// print the output URL
	fmt.Println(outputURL)

	return outputURL, nil
}

func FormatMoney(amount string) string {
	return strings.ReplaceAll(amount, ",", ".")
}

func FormatCPF(cpf string) string {
	newCPF := strings.ReplaceAll(cpf, ".", "")
	newCPF = strings.ReplaceAll(newCPF, "-", "")
	return newCPF
}

func getBanks() []entity.Bank {
	banks := []entity.Bank{
		{Name: "Banco BRADESCO", Code: "237", ISPB: "60746948"},
		{Name: "Banco BTG PACTUAL", Code: "208", ISPB: "30306294"},
		{Name: "Bancoob", Code: "756", ISPB: "2038232"},
		{Name: "Banco Santander", Code: "33", ISPB: "90400888"},
		{Name: "Banco Daycoval", Code: "707", ISPB: "62232889"},
		{Name: "Banco do Brasil", Code: "1", ISPB: "0"},
		{Name: "Caixa Economica Federal", Code: "104", ISPB: "360305"},
		{Name: "Itaú Unibanco", Code: "341", ISPB: "60701190"},
		{Name: "Nubank", Code: "260", ISPB: "18236120"},
		{Name: "Banco Intermedium", Code: "77", ISPB: "416968"},
	}
	return banks
}
