package output

import (
	"os"
	"strings"
	"testing"

	"Crawler-CLI-TA/base/models"
)

func TestJSONWriter_Write(t *testing.T) {
	file, err := os.CreateTemp("", "writer-test.json")
	if err != nil {
		t.Fatal(err)
	}
	path := file.Name()

	defer os.Remove(path)
	file.Close()

	pages := []models.Page{
		{
			Resource: "https://example.com",
			Title:    "Example",
			Links: []models.Page{
				{
					Resource: "https://example.com/about",
					Title:    "About",
					Links:    []models.Page{},
				},
			},
		},
	}

	writer := NewJSONWriter()

	err = writer.Write(pages, path)
	if err != nil {
		t.Fatalf("writing error:%v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading written file:%v", err)
	}

	content := string(data)

	if content == "" {
		t.Fatal("output is empty string")
	}

	if !strings.Contains(content, "https://example.com") {
		t.Fatal("output does not contain resource URL")
	}

	if !strings.Contains(content, "Example") {
		t.Fatal("output does not contain page title")
	}
}
