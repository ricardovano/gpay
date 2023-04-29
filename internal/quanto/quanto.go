package quanto

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ricardovano/qpay/internal/entity"
)

func GetToken() string {
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
	return tokenObj.AccessToken
}

func GetParticipants(token string) (*entity.AuthorisationServers, error) {
	url := "https://api-quanto.com/opb-api/v1/participants/payments"
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
