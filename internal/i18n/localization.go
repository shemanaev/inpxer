package i18n

import (
	"embed"
	"io/fs"

	"github.com/vorlif/spreak"
	"golang.org/x/text/language"
)

const (
	DefaultDomain = "inpxer"
	GenresDomain  = "genres"
)

//go:embed locale/*
var localeFiles embed.FS

func GetLocalizer(lang string) (*spreak.Localizer, error) {
	files, _ := fs.Sub(localeFiles, "locale")

	bundle, err := spreak.NewBundle(
		//spreak.WithSourceLanguage(language.English), // this will break genres loading in english
		spreak.WithDefaultDomain(DefaultDomain),
		spreak.WithDomainFs(DefaultDomain, files),
		spreak.WithDomainFs(GenresDomain, files),
		spreak.WithLanguage(language.English, language.Russian),
	)
	if err != nil {
		return nil, err
	}

	t := spreak.NewLocalizer(bundle, lang)
	return t, nil
}
