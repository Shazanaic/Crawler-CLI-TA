package models

type Result struct {
	Task  Task
	Page  *Page
	Error error
}
