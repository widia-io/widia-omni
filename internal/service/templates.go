package service

import "strings"

type AreaTemplate struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Icon      string  `json:"icon"`
	Color     string  `json:"color"`
	Weight    float64 `json:"weight"`
	SortOrder int     `json:"sort_order"`
}

type GoalSuggestion struct {
	AreaSlug string `json:"area_slug"`
	Title    string `json:"title"`
	Period   string `json:"period"`
}

var areaTemplates = map[string][]AreaTemplate{
	"pt-BR": {
		{Name: "Saúde & Fitness", Slug: "saude-fitness", Icon: "dumbbell", Color: "green", Weight: 1, SortOrder: 1},
		{Name: "Carreira & Trabalho", Slug: "carreira-trabalho", Icon: "briefcase", Color: "orange", Weight: 1, SortOrder: 2},
		{Name: "Finanças", Slug: "financas", Icon: "dollar-sign", Color: "sand", Weight: 1, SortOrder: 3},
		{Name: "Relacionamentos", Slug: "relacionamentos", Icon: "users", Color: "rose", Weight: 1, SortOrder: 4},
		{Name: "Desenvolvimento Pessoal", Slug: "desenvolvimento-pessoal", Icon: "brain", Color: "sage", Weight: 1, SortOrder: 5},
		{Name: "Casa & Ambiente", Slug: "casa-ambiente", Icon: "home", Color: "blue", Weight: 1, SortOrder: 6},
		{Name: "Lazer & Diversão", Slug: "lazer-diversao", Icon: "gamepad-2", Color: "sky", Weight: 1, SortOrder: 7},
		{Name: "Espiritualidade", Slug: "espiritualidade", Icon: "sparkles", Color: "violet", Weight: 1, SortOrder: 8},
	},
	"en-US": {
		{Name: "Health & Fitness", Slug: "health-fitness", Icon: "dumbbell", Color: "green", Weight: 1, SortOrder: 1},
		{Name: "Career & Work", Slug: "career-work", Icon: "briefcase", Color: "orange", Weight: 1, SortOrder: 2},
		{Name: "Finances", Slug: "finances", Icon: "dollar-sign", Color: "sand", Weight: 1, SortOrder: 3},
		{Name: "Relationships", Slug: "relationships", Icon: "users", Color: "rose", Weight: 1, SortOrder: 4},
		{Name: "Personal Development", Slug: "personal-development", Icon: "brain", Color: "sage", Weight: 1, SortOrder: 5},
		{Name: "Home & Environment", Slug: "home-environment", Icon: "home", Color: "blue", Weight: 1, SortOrder: 6},
		{Name: "Leisure & Fun", Slug: "leisure-fun", Icon: "gamepad-2", Color: "sky", Weight: 1, SortOrder: 7},
		{Name: "Spirituality", Slug: "spirituality", Icon: "sparkles", Color: "violet", Weight: 1, SortOrder: 8},
	},
}

var goalSuggestions = map[string]map[string][]GoalSuggestion{
	"pt-BR": {
		"saude-fitness": {
			{AreaSlug: "saude-fitness", Title: "Treinar 4x por semana", Period: "quarterly"},
			{AreaSlug: "saude-fitness", Title: "Fazer check-up anual", Period: "yearly"},
		},
		"carreira-trabalho": {
			{AreaSlug: "carreira-trabalho", Title: "Completar certificação", Period: "quarterly"},
			{AreaSlug: "carreira-trabalho", Title: "Pedir promoção", Period: "yearly"},
		},
		"financas": {
			{AreaSlug: "financas", Title: "Montar reserva de emergência", Period: "yearly"},
			{AreaSlug: "financas", Title: "Economizar 20% do salário", Period: "monthly"},
		},
		"relacionamentos": {
			{AreaSlug: "relacionamentos", Title: "Jantar em família 1x por semana", Period: "quarterly"},
		},
		"desenvolvimento-pessoal": {
			{AreaSlug: "desenvolvimento-pessoal", Title: "Ler 12 livros no ano", Period: "yearly"},
			{AreaSlug: "desenvolvimento-pessoal", Title: "Meditar diariamente", Period: "quarterly"},
		},
		"casa-ambiente": {
			{AreaSlug: "casa-ambiente", Title: "Organizar um cômodo por mês", Period: "quarterly"},
		},
		"lazer-diversao": {
			{AreaSlug: "lazer-diversao", Title: "Viajar 2x no ano", Period: "yearly"},
		},
		"espiritualidade": {
			{AreaSlug: "espiritualidade", Title: "Praticar gratidão diária", Period: "quarterly"},
		},
	},
	"en-US": {
		"health-fitness": {
			{AreaSlug: "health-fitness", Title: "Work out 4x per week", Period: "quarterly"},
			{AreaSlug: "health-fitness", Title: "Get annual health check-up", Period: "yearly"},
		},
		"career-work": {
			{AreaSlug: "career-work", Title: "Complete a certification", Period: "quarterly"},
			{AreaSlug: "career-work", Title: "Ask for a promotion", Period: "yearly"},
		},
		"finances": {
			{AreaSlug: "finances", Title: "Build emergency fund", Period: "yearly"},
			{AreaSlug: "finances", Title: "Save 20% of salary", Period: "monthly"},
		},
		"relationships": {
			{AreaSlug: "relationships", Title: "Family dinner once a week", Period: "quarterly"},
		},
		"personal-development": {
			{AreaSlug: "personal-development", Title: "Read 12 books this year", Period: "yearly"},
			{AreaSlug: "personal-development", Title: "Meditate daily", Period: "quarterly"},
		},
		"home-environment": {
			{AreaSlug: "home-environment", Title: "Organize one room per month", Period: "quarterly"},
		},
		"leisure-fun": {
			{AreaSlug: "leisure-fun", Title: "Travel twice a year", Period: "yearly"},
		},
		"spirituality": {
			{AreaSlug: "spirituality", Title: "Practice daily gratitude", Period: "quarterly"},
		},
	},
}

func resolveLocale(locale string) string {
	if _, ok := areaTemplates[locale]; ok {
		return locale
	}
	lang := strings.SplitN(locale, "-", 2)[0]
	if strings.EqualFold(lang, "pt") {
		return "pt-BR"
	}
	return "en-US"
}

func GetAreaTemplates(locale string) []AreaTemplate {
	return areaTemplates[resolveLocale(locale)]
}

func GetGoalSuggestions(locale, areaSlug string) []GoalSuggestion {
	loc := resolveLocale(locale)
	if slugMap, ok := goalSuggestions[loc]; ok {
		if suggestions, ok := slugMap[areaSlug]; ok {
			return suggestions
		}
	}
	return []GoalSuggestion{}
}
