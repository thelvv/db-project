package entity

import "github.com/go-openapi/strfmt"

type Post struct {
	ID       int             `json:"id"`
	Author   string          `json:"author"`
	Message  string          `json:"message"`
	Parent   int             `json:"parent,omitempty"`
	Forum    string          `json:"forum"`
	Thread   int             `json:"thread"`
	Created  strfmt.DateTime `json:"created,omitempty"`
	IsEdited bool            `json:"isEdited"`
}

type PostOutput struct {
	Post   *Post   `json:"post"`
	Author *User   `json:"author,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
}
