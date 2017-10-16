package scraper

import (
	"github.com/jinzhu/gorm"
	// "github.com/qor/l10n"
)

type Group struct {
	gorm.Model
	Name string
	// l10n.LocaleCreatable
}
