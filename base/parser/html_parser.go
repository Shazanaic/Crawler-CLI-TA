package parser

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
)

type HTMLParser struct{}

func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

func (p *HTMLParser) Parse(data []byte) (ParseResult, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return ParseResult{}, err
	}

	result := ParseResult{}

	result.Title = doc.Find("title").First().Text()

	urls := make([]string, 0)

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists && href != "" {
			urls = append(urls, href)
		}
	})
	result.URLs = urls

	return result, nil
}
