package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
)

func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	pemString := string(pem.EncodeToMemory(privateKeyBlock))

	creds := map[string]string{
		"type":                        "service_account",
		"project_id":                  "dummy",
		"private_key_id":              "dummy_id",
		"private_key":                 pemString,
		"client_email":                "dummy@dummy.iam.gserviceaccount.com",
		"client_id":                   "12345",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/dummy%40dummy.iam.gserviceaccount.com",
		"universe_domain":             "googleapis.com",
	}

	b, _ := json.Marshal(creds)
	os.WriteFile("creds.json", b, 0644)
}
