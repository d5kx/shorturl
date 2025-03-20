package models

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

type RequestJSONBatch struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseSONBatch struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
