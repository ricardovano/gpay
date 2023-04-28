package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ricardovano/qpay/internal/entity"
	"github.com/ricardovano/qpay/internal/infra/database"
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
	err := http.ListenAndServe(":80", nil)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Server running on port 80")
	}

}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {

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
		beneficiary.Code = getRandonString()

		//SAVE ON DATABASE
		database.CreateBeneficiary(beneficiary)

		println("Created beneficiary with code " + beneficiary.Code)

		//SEND EMAIL WITH LINK TO THE FRIEND (FULL URL)

		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		tmpl, err := template.ParseFiles(wd + "/static/payment_registered.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, beneficiary)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.Method == "GET" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		tmpl, err := template.ParseFiles(wd + "/static/payment_register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var dto entity.StatusResponse
		dto.Banks = getBanks()

		err = tmpl.Execute(w, dto)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func payHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	beneficiary := database.GetBeneficiary(code)

	token := getToken()

	payment := entity.PaymentRequest{
		Type: "PIX",
		Payer: entity.Payer{
			Name:          r.FormValue("name"),
			CPF:           FormatCPF(r.FormValue("cpf")),
			Email:         r.FormValue("email"),
			ParticipantId: r.FormValue("bank"),
		},
		Beneficiary:           beneficiary,
		Amount:                FormatMoney(r.FormValue("amount")),
		ReturnUri:             getHost() + "status",
		WebhookUrl:            getHost() + "webhook",
		ReferenceCode:         beneficiary.Code, //TODO: PERSIST TRANSACTION IN DATABASE
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
		statusResponse.Success = "Solicitação concluída com sucesso!"
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

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		tmpl, err := template.ParseFiles(wd + "/static/payment_register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var dto entity.StatusResponse
		dto.Banks = getBanks()

		err = tmpl.Execute(w, dto)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		beneficiary := database.GetBeneficiary(code)

		token := getToken()
		participants, err := getParticipants(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles(wd + "/static/payment_list.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		paymentDTO := entity.PaymentDTO{}
		paymentDTO.Data = participants.Data
		paymentDTO.Beneficiary = beneficiary

		err = tmpl.Execute(w, paymentDTO)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
			println(a.CustomerFriendlyName)
			if a.CustomerFriendlyName == "Nubank" {
				data = append(data, a)
			}
			if a.CustomerFriendlyName == "Mercado Pago" {
				data = append(data, a)
			}
			if strings.Contains(a.CustomerFriendlyName, "CAIXA") {
				a.CustomerFriendlyName = "Caixa"
				data = append(data, a)
			}
			if strings.Contains(a.CustomerFriendlyName, "Bradesco Pessoa Física") {
				a.CustomerFriendlyName = "Bradesco"
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
	queryParams.Set("redirect_uri", getHost()+"status")

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

func getRandonString() string {

	timestamp := time.Now().Unix()
	var chars = "abcdefghijklmnopqrstuvwxyz"

	length := 4

	ll := len(chars)
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%ll]
	}

	result := string(b) + strconv.FormatInt(timestamp, 10)
	return result
}

func getHost() string {
	host, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
	}

	fullUrl := ""
	if host == "QT-MBP00093.local" {
		fullUrl = "http://localhost/"
	} else {
		fullUrl = "http://www.cafedoamigo.com.br/"
	}

	return fullUrl
}
