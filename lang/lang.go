// Package lang introduces a translation and localisation API for Fyne applications
//
// Since 2.5
package lang

import (
	"embed"
	"encoding/json"
	"log"
	"strings"
	"text/template"

	"fyne.io/fyne/v2"
	"github.com/fyne-io/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"golang.org/x/text/language"
)

var (
	// L is a shortcut to localize a string, similar to the gettext "_" function.
	// More info available on the `Localize` function.
	L = Localize

	// X is a shortcut to get the locaization of a string with specified key.
	// More info available on the `LocalizeKey` function.
	X = LocalizeKey

	// N is a shortcut to localize a string with plural forms, similar to the ngettext function.
	// More info available on the `LocalizePlural` function.
	N = LocalizePlural

	bundle    *i18n.Bundle
	localizer *i18n.Localizer

	//go:embed translations
	translations embed.FS
	translated   []language.Tag
)

// Localize asks the translation engine to translate a string, this behaves like the gettext "_" function.
// The string can be templated and the template data can be passed as a struct with exported fields,
// or as a map of string keys to any suitable value.
func Localize(in string, data ...any) string {
	return LocalizeKey(in, in, data...)
}

// LocalizeKey asks the translation engine for the translation with specific ID.
// If it cannot be found then the fallback will be used.
// The string can be templated and the template data can be passed as a struct with exported fields,
// or as a map of string keys to any suitable value.
func LocalizeKey(key, fallback string, data ...any) string {
	var d0 any
	if len(data) > 0 {
		d0 = data[0]
	}

	ret, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    key,
			Other: fallback,
		},
		TemplateData: d0,
	})

	if err != nil {
		fyne.LogError("Translation failure", err)
		return fallbackWithData(key, fallback, d0)
	}
	return ret
}

// LocalizePlural asks the translation engine to translate a string from one of a number of plural forms.
// This behaves like the ngettext function, with the `count` parameter determining the plurality looked up.
// The string can be templated and the template data can be passed as a struct with exported fields,
// or as a map of string keys to any suitable value.
func LocalizePlural(in string, count int, data ...any) string {
	var d0 any
	if len(data) > 0 {
		d0 = data[0]
	}

	ret, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    in,
			Other: in,
		},
		PluralCount:  count,
		TemplateData: d0,
	})

	if err != nil {
		fyne.LogError("Translation failure", err)
		return fallbackWithData(in, in, d0)
	}
	return ret
}

// AddTranslations allows an app to load a bundle of translations.
// The language that this relates to will be inferred from the resource name, for example "fr.json".
// The data should be in json format.
func AddTranslations(r fyne.Resource) error {
	return addLanguage(r.Content(), r.Name())
}

// AddTranslationsForLocale allows an app to load a bundle of translations for a specified locale.
// The data should be in json format.
func AddTranslationsForLocale(data []byte, l fyne.Locale) error {
	return addLanguage(data, l.String()+".json")
}

func addLanguage(data []byte, name string) error {
	_, err := bundle.ParseMessageFileBytes(data, name)
	return err
}

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	loadTranslationsFromFS(translations, "translations")

	// Find the closest translation from the user's locale list and set it up
	all, err := locale.GetLocales()
	if err != nil {
		fyne.LogError("Failed to load user locales", err)
		all = []string{"en"}
	}
	str := closestSupportedLocale(all).LanguageString()
	setupLang(str)
	localizer = i18n.NewLocalizer(bundle, str)
}

func fallbackWithData(key, fallback string, data any) string {
	t, err := template.New(key).Parse(fallback)
	if err != nil {
		log.Println("Could not parse fallback template")
		return fallback
	}
	str := &strings.Builder{}
	_ = t.Execute(str, data)
	return str.String()
}

func loadTranslationsFromFS(fs embed.FS, dir string) {
	files, err := fs.ReadDir(dir)
	if err != nil {
		fyne.LogError("failed to read bundled translations", err)
		return
	}

	for _, f := range files {
		data, err := fs.ReadFile(dir + "/" + f.Name())
		if err != nil {
			fyne.LogError("failed to read bundled translation", err)
			continue
		}
		bundle.MustParseMessageFileBytes(data, f.Name())

		name := "en"
		if !strings.Contains(f.Name(), "template") {
			name = f.Name()[5 : len(f.Name())-5]
		}
		tag := language.Make(name)
		translated = append(translated, tag)
	}
}

// A utility for setting up languages - available to unit tests for overriding system
func setupLang(lang string) {
	localizer = i18n.NewLocalizer(bundle, lang)
}