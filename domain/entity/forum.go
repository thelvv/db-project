package entity

type Forum struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	User    string `json:"user"`
	Threads int    `json:"threads"`
	Posts   int    `json:"posts"`
}

type ForumInput struct {
	Slug   string `json:"slug"`
	Tittle string `json:"title"`
	User   string `json:"user"`
}
