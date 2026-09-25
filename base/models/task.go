package models

type Task struct {
	URL     string
	Depth   int
	RootURL string

	Parent *Page
}
