package entities

type Link struct {
	UUID        string `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	DeletedFlag bool   `json:"deleted_flag,omitempty"`
}
