package dto

import "encoding/json"

type GCPServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

func (g *GCPServiceAccount) ToBytes() ([]byte, error) {
	bytes, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	return bytes, err
}
