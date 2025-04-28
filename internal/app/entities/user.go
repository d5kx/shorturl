package entities

type User struct {
	UUID       string `json:"uuid,omitempty"`
	Login      string `json:"login,omitempty"`
	PasswdHash string `json:"password,omitempty"`
}
