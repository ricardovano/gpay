package tools

import (
	"crypto/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetHost() string {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	fullUrl := ""
	if host == "QT-MBP00093.local" {
		fullUrl = "http://localhost/"
	} else {
		fullUrl = "http://www.cafedoamigo.com.br/"
	}

	return fullUrl
}

func GetRandonString() string {

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

func FormatMoney(amount string) string {
	return strings.ReplaceAll(amount, ",", ".")
}

func FormatCPF(cpf string) string {
	newCPF := strings.ReplaceAll(cpf, ".", "")
	newCPF = strings.ReplaceAll(newCPF, "-", "")
	return newCPF
}

func ReplaceRedirectUri(inputStr string) (string, error) {

	inputURL, err := url.Parse(inputStr)
	if err != nil {
		return "", err
	}

	queryParams, err := url.ParseQuery(inputURL.RawQuery)
	if err != nil {
		return "", err
	}

	queryParams.Set("redirect_uri", GetHost()+"status")
	outputURL := inputURL.Scheme + "://" + inputURL.Host + inputURL.Path + "?" + queryParams.Encode()
	outputURL = strings.ReplaceAll(outputURL, "%2F", "/")

	return outputURL, nil
}
