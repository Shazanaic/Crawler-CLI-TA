package output

import (
	"Crawler-CLI-TA/base/models"
)

type Writer interface {
	Write(pages []models.Page) error
}
