package models

type RequestJSON struct {
	URL string `json:"url"`
}

type UserLoginJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

type RequestJSONBatch struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseJSONBatch struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ResponseJSONGetUserUrls struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
