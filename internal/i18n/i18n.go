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
	log.Printf("[i18n] Loading translations from directory: %s", localeDir)
	if translations == nil {
		translations = make(map[string]map[string]string)
	}

	files, err := filepath.Glob(filepath.Join(localeDir, "*.json"))
	if err != nil {
		return fmt.Errorf("error finding translation files: %w", err)
	}
	log.Printf("[i18n] Found translation files: %v", files)

	for _, file := range files {
		log.Printf("[i18n] Processing translation file: %s", file)
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
	langMap, langOk := translations[langCode] // Use a different variable name for this 'ok'
	if !langOk { // If the specific language code is not found
		langMap = translations[defaultLanguage] // Fallback to default language map
		if langMap == nil { // Check if the default language map itself is missing (e.g., "en.json" not loaded)
			// Return a message indicating the original langCode and key, and that default was also not found.
			return fmt.Sprintf("[translation missing for lang: %s, key: %s (default lang '%s' not loaded/found)]", langCode, key, defaultLanguage)
		}
	}

	formatString, keyOk := langMap[key] // keyOk refers to key's presence in the current langMap
	if !keyOk { // If key is not in the current langMap (which might be the original or default)
		// If we weren't already using the default language, and the key was not found,
		// it means we tried the specific lang, it failed (either lang or key), and now we are effectively checking the default.
		// If langCode was already defaultLanguage, this block won't be entered, correctly.
		if langCode != defaultLanguage && translations[defaultLanguage] != nil {
			// Explicitly try fetching from default language map if we haven't already.
			// This handles the case where langCode existed, but the key was missing,
			// and we need to check the default language for the key.
			formatString, keyOk = translations[defaultLanguage][key]
		}

		if !keyOk { // If key is still not found (even after trying default if applicable)
			return fmt.Sprintf("[translation missing for key: %s (in lang '%s' and default '%s')]", key, langCode, defaultLanguage)
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
