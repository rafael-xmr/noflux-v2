// SPDX-FileCopyrightText: Copyright The Noflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "github.com/fiatjaf/noflux/internal/locale"

import (
	"embed"
	"encoding/json"
	"fmt"
)

type translationDict map[string]interface{}
type catalog map[string]translationDict

var defaultCatalog = make(catalog, len(AvailableLanguages))

//go:embed translations/*.json
var translationFiles embed.FS

func GetTranslationDict(language string) (translationDict, error) {
	if _, ok := defaultCatalog[language]; !ok {
		var err error
		if defaultCatalog[language], err = loadTranslationFile(language); err != nil {
			return nil, err
		}
	}
	return defaultCatalog[language], nil
}

// LoadCatalogMessages loads and parses all translations encoded in JSON.
func LoadCatalogMessages() error {
	var err error

	for language := range AvailableLanguages {
		defaultCatalog[language], err = loadTranslationFile(language)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadTranslationFile(language string) (translationDict, error) {
	translationFileData, err := translationFiles.ReadFile(fmt.Sprintf("translations/%s.json", language))
	if err != nil {
		return nil, err
	}

	translationMessages, err := parseTranslationMessages(translationFileData)
	if err != nil {
		return nil, err
	}

	return translationMessages, nil
}

func parseTranslationMessages(data []byte) (translationDict, error) {
	var translationMessages translationDict
	if err := json.Unmarshal(data, &translationMessages); err != nil {
		return nil, fmt.Errorf(`invalid translation file: %w`, err)
	}
	return translationMessages, nil
}
