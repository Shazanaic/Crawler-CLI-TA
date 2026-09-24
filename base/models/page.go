package models

type Page struct {
	Resource string `json:"resource"`
	Title    string `json:"title"`
	Links    []Page `json:"links"`
}
