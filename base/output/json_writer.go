package output

import (
	"encoding/json"
	"os"

	"Crawler-CLI-TA/base/models"
)

type JSONWriter struct{}

func NewJSONWriter() *JSONWriter {
	return &JSONWriter{}
}

func (w *JSONWriter) Write(pages []models.Page, path string) error {
	data, err := json.MarshalIndent(pages, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
