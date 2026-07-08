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

// Planos competitivos com o Fivemanage ($8/mês ≈ R$40 para 50GB).
// 5Keepr oferece mais storage pelo mesmo preço ou mais barato.
var defaultTiers = []Tier{
	{
		ID:   "free",
		Name: "Gratuito",
		Cost: 0,
		Description: Info{
			Pt: "Para testar a plataforma. Sem cartão de crédito.",
			En: "To test the platform. No credit card required.",
		},
		Included: []Info{
			{Pt: "1GB de Armazenamento", En: "1GB Storage"},
			{Pt: "1 Chave de API", En: "1 API Key"},
			{Pt: "5000 Requisições/semana", En: "5000 Requests/week"},
			{Pt: "Subdomínio compartilhado", En: "Shared subdomain"},
		},
	},
	{
		ID:   "starter",
		Name: "Iniciante",
		Cost: 19.90,
		Description: Info{
			Pt: "Ideal para comunidades pequenas.",
			En: "Ideal for small communities.",
		},
		Included: []Info{
			{Pt: "60GB de Armazenamento", En: "60GB Storage"},
			{Pt: "3 Chaves de API", En: "3 API Keys"},
			{Pt: "Requisições Ilimitadas", En: "Unlimited Requests"},
			{Pt: "Subdomínio compartilhado", En: "Shared subdomain"},
			{Pt: "Suporte por e-mail", En: "Email support"},
		},
	},
	{
		ID:   "pro",
		Name: "Profissional",
		Cost: 34.90,
		Description: Info{
			Pt: "Para comunidades em crescimento. Domínio próprio e backup automático.",
			En: "For growing communities. Custom domain and automatic backup.",
		},
		Included: []Info{
			{Pt: "150GB de Armazenamento", En: "150GB Storage"},
			{Pt: "10 Chaves de API com permissões", En: "10 API Keys with permissions"},
			{Pt: "Requisições Ilimitadas", En: "Unlimited Requests"},
			{Pt: "Domínio próprio", En: "Custom domain"},
			{Pt: "Backup automático diário", En: "Daily automatic backup"},
			{Pt: "Suporte prioritário 24/7", En: "Priority 24/7 support"},
		},
	},
	{
		ID:   "enterprise",
		Name: "Empresarial",
		Cost: 49.90,
		Description: Info{
			Pt: "Para grandes operações. SLA garantido.",
			En: "For large operations. Guaranteed SLA.",
		},
		Included: []Info{
			{Pt: "250GB de Armazenamento", En: "250GB Storage"},
			{Pt: "Chaves de API ilimitadas", En: "Unlimited API Keys"},
			{Pt: "Requisições Ilimitadas", En: "Unlimited Requests"},
			{Pt: "Domínio próprio", En: "Custom domain"},
			{Pt: "Backup automático diário", En: "Daily backup"},
			{Pt: "Até 5 Storages", En: "Up to 5 Storages"},
			{Pt: "SLA 99,9% de uptime", En: "99.9% uptime SLA"},
			{Pt: "Suporte dedicado", En: "Dedicated support"},
		},
	},
}

func (u *TierService) GetTierNameByID(tierID string) string {
	for _, tier := range defaultTiers {
		if tier.ID == tierID {
			return tier.Name
		}
	}
	return "Desconhecido"
}

func (u *TierService) GetAllTiers() []Tier {
	return defaultTiers
}

func (u *TierService) GetTierCostByID(tierID string) float64 {
	for _, tier := range defaultTiers {
		if tier.ID == tierID {
			return tier.Cost
		}
	}
	return 0
}
