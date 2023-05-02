package quanto

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ricardovano/qpay/config"
	"github.com/ricardovano/qpay/internal/entity"
)

func GetToken() string {

	config := config.GetConfig()
	url := config.TokenUrl

	payload := strings.NewReader("client_id=" + config.ClientId + "&client_secret=" + config.ClientSecret + "&grant_type=" + config.GrantType)
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
	return tokenObj.AccessToken
}

func GetParticipants(token string) (*entity.AuthorisationServers, error) {

	config := config.GetConfig()
	url := config.ParticipantsUrl

	req, err := http.NewRequest("GET", url, nil)

	oauth := "Bearer " + token
	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", oauth)

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var participants []entity.Participant
	err = json.NewDecoder(resp.Body).Decode(&participants)
	if err != nil {
		return nil, err
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
			if strings.Contains(a.CustomerFriendlyName, "Bradesco Pessoa Física") {
				a.CustomerFriendlyName = "Bradesco"
				data = append(data, a)
			}
			if strings.Contains(a.CustomerFriendlyName, "Banco do Brasil") {
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

func PostPayment(payment entity.PaymentRequest, token string) entity.PaymentResponse {

	config := config.GetConfig()
	url := config.PaymentUrl

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
		panic(err)
	}
	return response
}
