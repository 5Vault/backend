package AbacatePay

import "os"

type AbacatePay struct{}

func NewAbacatePay() *AbacatePay {
	return &AbacatePay{}
}

func (a *AbacatePay) ValidateSecretWebhook(secret string) bool {
	return secret == os.Getenv("SECRET_WEBHOOK_URL")
}
