package service

type Tier struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	Description Info    `json:"description"`
	Included    []Info  `json:"included"`
}

type Info struct {
	En string `json:"en"`
	Pt string `json:"pt"`
}

type TierService struct{}

func NewTierService() *TierService {
	return &TierService{}
}

// DefaultTiers contains the available tiers. Defined as a variable (not inside a type).
var defaultTiers = []Tier{
	{
		ID:   "free",
		Name: "Free",
		Cost: 0,
		Description: Info{
			En: "Ideal for testing and small projects. Limited features. No credit card required.",
			Pt: "Ideal para testes e pequenos projetos. Recursos limitados. Nenhum cartão de crédito necessário.",
		},
		Included: []Info{
			{En: "5GB Storage", Pt: "5GB de Armazenamento"},
			{En: "Basic Support", Pt: "Suporte Básico"},
			{En: "1000 Requests per week", Pt: "1000 Requisições por semana"},
		},
	},
	{
		ID:   "basic",
		Name: "Basic",
		Cost: 35.99,
		Description: Info{
			En: "Perfect for small communities. Pay with Pix or Credit Card",
			Pt: "Perfeito para pequenas comunidades. Pague com Pix ou Cartão de Crédito",
		},
		Included: []Info{
			{En: "40GB Storage", Pt: "40GB de Armazenamento"},
			{En: "Priority Support", Pt: "Suporte Prioritário"},
			{En: "Unlimited Requests", Pt: "Requisições Ilimitadas"},
		},
	},
	{
		ID:   "pro",
		Name: "Professional",
		Cost: 59.99,
		Description: Info{
			En: "Best for growing communities. Advanced features and priority support.",
			Pt: "Melhor para comunidades em crescimento. Recursos avançados e suporte prioritário.",
		},
		Included: []Info{
			{En: "70GB Storage", Pt: "70GB de Armazenamento"},
			{En: "24/7 Support", Pt: "Suporte 24/7"},
			{En: "Unlimited Requests", Pt: "Requisições Ilimitadas"},
		},
	},
}

func (u *TierService) GetTierNameByID(tierID string) string {
	for _, tier := range defaultTiers {
		if tier.ID == tierID {
			return tier.Name
		}
	}
	return "Unknown"
}

func (u *TierService) GetAllTiers() []Tier {
	return defaultTiers
}
