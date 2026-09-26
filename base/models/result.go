package models

type Result struct {
	Task  Task
	Page  *Page
	Links []string
	Error error
}
