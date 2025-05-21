package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	translations    map[string]map[string]string
	defaultLanguage = "en"
)

// LoadTranslations loads translation files from the given directory.
func LoadTranslations(localeDir string) error {
	if translations == nil {
		translations = make(map[string]map[string]string)
	}

	files, err := filepath.Glob(filepath.Join(localeDir, "*.json"))
	if err != nil {
		return fmt.Errorf("error finding translation files: %w", err)
	}

	for _, file := range files {
		langCode := filepath.Base(file)
		langCode = langCode[:len(langCode)-len(filepath.Ext(langCode))] // Remove .json extension

		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading translation file %s: %v", file, err)
			continue // Skip this file
		}

		var langMap map[string]string
		if err := json.Unmarshal(data, &langMap); err != nil {
			log.Printf("Error unmarshalling translation file %s: %v", file, err)
			continue // Skip this file
		}
		translations[langCode] = langMap
	}
	return nil
}

// Localize returns the localized string for the given key and language.
// It falls back to the default language if the key is not found in the given language.
func Localize(langCode string, key string, args ...interface{}) string {
	langMap, ok := translations[langCode]
	if !ok {
		langMap = translations[defaultLanguage]
		if !ok {
			return fmt.Sprintf("[translation missing for lang: %s, key: %s]", langCode, key)
		}
	}

	formatString, ok := langMap[key]
	if !ok {
		// Try fallback to default language if not already using it
		if langCode != defaultLanguage {
			langMap = translations[defaultLanguage]
			if langMap != nil {
				formatString, ok = langMap[key]
			}
		}
		if !ok {
			return fmt.Sprintf("[translation missing for key: %s]", key)
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(formatString, args...)
	}
	return formatString
}

// AvailableLanguages returns a slice of all loaded language codes.
func AvailableLanguages() []string {
	if translations == nil {
		return []string{}
	}
	keys := make([]string, 0, len(translations))
	for k := range translations {
		keys = append(keys, k)
	}
	return keys
}

// ResetTranslationsForTest resets the translations map for testing purposes.
// This function should only be used in tests.
func ResetTranslationsForTest() {
	translations = make(map[string]map[string]string)
}
