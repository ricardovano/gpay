package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ricardovano/qpay/internal/entity"
	"github.com/ricardovano/qpay/internal/infra/database"
	"github.com/ricardovano/qpay/internal/infra/quanto"
	"github.com/ricardovano/qpay/internal/infra/tools"
)

func main() {
	http.HandleFunc("/", participantsHandler)
	http.HandleFunc("/pay", payHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/log", logHandler)

	fs := http.FileServer(http.Dir("static"))
	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
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

	log.Default()

}

func logHandler(w http.ResponseWriter, r *http.Request) {

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.ParseFiles(wd + "/static/log.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var dto entity.LogDTO
	dto.Beneficiaries = database.GetAllBeneficiaries()
	dto.Logs = database.GetAllLogs()
	err = tmpl.Execute(w, dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		beneficiary.CPFCNPJ = tools.FormatCPF(r.FormValue("cpf"))
		beneficiary.Name = r.FormValue("name")
		beneficiary.ISPB = r.FormValue("ispb")
		beneficiary.Issuer = r.FormValue("issuer")
		beneficiary.Number = r.FormValue("number")
		beneficiary.AccountType = "CACC"
		beneficiary.Email = r.FormValue("email")

		//generate
		beneficiary.Code = tools.GetRandonString()

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
		dto.Banks = entity.GetBanks()

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

	token := quanto.GetToken()

	payment := entity.PaymentRequest{
		Type: "PIX",
		Payer: entity.Payer{
			Name:          r.FormValue("name"),
			CPF:           tools.FormatCPF(r.FormValue("cpf")),
			Email:         r.FormValue("email"),
			ParticipantId: r.FormValue("bank"),
		},
		Beneficiary:           beneficiary,
		Amount:                tools.FormatMoney(r.FormValue("amount")),
		ReturnUri:             tools.GetHost() + "status",
		WebhookUrl:            tools.GetHost() + "webhook",
		ReferenceCode:         beneficiary.Code,
		TermsOfUseVersion:     "1",
		TermsOfPrivacyVersion: "1",
	}

	response, err := quanto.PostPayment(payment, token)

	var log entity.Log
	log.Amount = payment.Amount
	log.Beneficiary = payment.Beneficiary
	log.Payer = payment.Payer
	if err != nil {
		log.Error = err.Error()
	}
	log.TransactionDate = time.Now()
	log.Status = response.Status
	database.CreateLog(log)

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
	statusResponse.Banks = entity.GetBanks()

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
		dto.Banks = entity.GetBanks()

		err = tmpl.Execute(w, dto)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		beneficiary := database.GetBeneficiary(code)

		token := quanto.GetToken()
		participants, err := quanto.GetParticipants(token)
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
