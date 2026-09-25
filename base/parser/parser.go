package parser

type ParseResult struct {
	Title string
	URLs  []string
}

type Parser interface {
	Parse(data []byte) (ParseResult, error)
}
