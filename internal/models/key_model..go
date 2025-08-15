package models

type KeysPayload struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}
