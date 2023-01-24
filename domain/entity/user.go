package entity

type User struct {
	ID       int    `json:"-"`
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname,omitempty"`
	Email    string `json:"email,omitempty"`
	About    string `json:"about,omitempty"`
}
