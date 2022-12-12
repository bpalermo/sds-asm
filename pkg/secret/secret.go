package secret

import "encoding/json"

type SecretsManagerSecret struct {
	CertificateChain []byte `json:"certificateChain,omitempty"`
	PrivateKey       []byte `json:"PrivateKey,omitempty"`
}

func Unmarshal(payload *string) (*SecretsManagerSecret, error) {
	var sms *SecretsManagerSecret
	err := json.Unmarshal([]byte(*payload), &sms)
	if err != nil {
		return nil, err
	}
	return sms, nil
}
