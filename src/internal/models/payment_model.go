package models

type ResponsePaymentMethod struct {
	PMID      string `json:"pm_id"`
	CardLast4 string `json:"card_last4"`
	CardBrand string `json:"card_brand"`
	ExpMonth  int    `json:"exp_month"`
	ExpYear   int    `json:"exp_year"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
}
