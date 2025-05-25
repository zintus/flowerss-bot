package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// setupTestLocales creates a temporary directory with locale files for testing.
// It takes a map where keys are filenames (e.g., "en.json") and values are the content to write.
func setupTestLocalesWithFiles(t *testing.T, files map[string][]byte) string {
	tmpDir, err := os.MkdirTemp("", "test_locales_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, content, 0644)
		if err != nil {
			// Clean up before failing
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to write to %s: %v", filePath, err)
		}
	}
	return tmpDir
}

// Helper to reset global state before each test requiring LoadTranslations
func resetGlobals() {
	translations = make(map[string]map[string]string)
	// defaultLanguage is already "en", no need to reset unless tests change it
}

func TestLoadTranslations(t *testing.T) {
	t.Run("ValidJSONFiles", func(t *testing.T) {
		resetGlobals()
		validEn := map[string]string{"greeting": "Hello", "farewell": "Goodbye"}
		validXx := map[string]string{"greeting": "Hallo"}
		enBytes, _ := json.Marshal(validEn)
		xxBytes, _ := json.Marshal(validXx)

		tmpDir := setupTestLocalesWithFiles(t, map[string][]byte{
			"en.json": enBytes,
			"xx.json": xxBytes,
		})
		defer os.RemoveAll(tmpDir)

		err := LoadTranslations(tmpDir)
		if err != nil {
			t.Fatalf("LoadTranslations failed: %v", err)
		}

		if len(translations) != 2 {
			t.Errorf("Expected 2 languages loaded, got %d", len(translations))
		}
		if val, ok := translations["en"]["greeting"]; !ok || val != "Hello" {
			t.Errorf("Expected en.greeting 'Hello', got '%s'", val)
		}
		if val, ok := translations["xx"]["greeting"]; !ok || val != "Hallo" {
			t.Errorf("Expected xx.greeting 'Hallo', got '%s'", val)
		}
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		resetGlobals()
		// LoadTranslations uses filepath.Glob which returns nil, nil for non-existent base dir if pattern includes it.
		// However, if the pattern is just the directory itself, it might depend on the Glob behavior.
		// The current implementation of LoadTranslations forms a pattern like "non_existent_dir/*.json".
		// filepath.Glob will return nil, nil in this case, which means no files found, and LoadTranslations will return nil.
		// To truly test a directory access error, one might need to make the directory unreadable,
		// or Glob would have to error on non-existent path, which it doesn't for patterns.
		// Let's assume the goal is to see no translations loaded and no error returned by LoadTranslations itself for this specific Glob behavior.
		err := LoadTranslations("non_existent_dir_for_sure")
		if err != nil {
			// Based on current LoadTranslations, an error from filepath.Glob for a non-existent path is not guaranteed.
			// If Glob returns an error, it's caught. If it returns nil, nil (no matches), LoadTranslations returns nil.
			// This test case might need refinement based on desired behavior for a non-existent directory.
			// For now, checking that no translations are loaded if the directory is empty/non-existent.
			t.Logf("LoadTranslations with non-existent directory returned error: %v (this might be acceptable depending on Glob)", err)
		}
		if len(translations) != 0 {
			t.Errorf("Expected 0 translations for non-existent directory, got %d", len(translations))
		}
	})

	t.Run("InvalidJSONFile", func(t *testing.T) {
		resetGlobals()
		validEn := map[string]string{"greeting": "Hello"}
		enBytes, _ := json.Marshal(validEn)
		invalidJsonBytes := []byte(`{"greeting": "Bad", "farewell":`) // Invalid JSON

		tmpDir := setupTestLocalesWithFiles(t, map[string][]byte{
			"en.json":       enBytes,
			"invalid.json":  invalidJsonBytes,
			"another.json":  enBytes, // to ensure one valid is loaded after skip
		})
		defer os.RemoveAll(tmpDir)

		// LoadTranslations logs errors for invalid JSON but doesn't return an error itself, it skips the file.
		err := LoadTranslations(tmpDir)
		if err != nil {
			t.Fatalf("LoadTranslations failed unexpectedly: %v", err)
		}

		if len(translations) != 2 { // en.json and another.json should load
			t.Errorf("Expected 2 languages loaded after skipping invalid JSON, got %d: %v", len(translations), translations)
		}
		if _, ok := translations["invalid"]; ok {
			t.Errorf("Expected 'invalid' language not to be loaded")
		}
		if val, ok := translations["en"]["greeting"]; !ok || val != "Hello" {
			t.Errorf("Expected en.greeting 'Hello', got '%s'", val)
		}
		if val, ok := translations["another"]["greeting"]; !ok || val != "Hello" {
			t.Errorf("Expected another.greeting 'Hello', got '%s'", val)
		}
	})

	t.Run("EmptyJSONFile", func(t *testing.T) {
		resetGlobals()
		emptyJsonBytes := []byte(`{}`)
		tmpDir := setupTestLocalesWithFiles(t, map[string][]byte{
			"empty.json": emptyJsonBytes,
		})
		defer os.RemoveAll(tmpDir)

		err := LoadTranslations(tmpDir)
		if err != nil {
			t.Fatalf("LoadTranslations failed for empty JSON: %v", err)
		}
		if langMap, ok := translations["empty"]; !ok {
			t.Errorf("Expected 'empty' language to be loaded")
		} else if len(langMap) != 0 {
			t.Errorf("Expected 'empty' language to have 0 translations, got %d", len(langMap))
		}
	})
}

func TestLocalize(t *testing.T) {
	resetGlobals() // Ensure clean state

	enContent := map[string]string{"hello": "Hello", "world_format": "World %s", "only_in_en": "Only English"}
	xxContent := map[string]string{"hello": "Hallo XX", "only_in_xx": "Only XX"}

	enBytes, _ := json.Marshal(enContent)
	xxBytes, _ := json.Marshal(xxContent)

	tmpDir := setupTestLocalesWithFiles(t, map[string][]byte{
		"en.json": enBytes, // en is default
		"xx.json": xxBytes,
	})
	defer os.RemoveAll(tmpDir)

	if err := LoadTranslations(tmpDir); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	// Test cases
	if Localize("xx", "hello") != "Hallo XX" {
		t.Errorf("Expected 'Hallo XX', got '%s'", Localize("xx", "hello"))
	}
	if Localize("xx", "world_format", "Test") != "World Test" { // Fallback to en
		t.Errorf("Expected 'World Test', got '%s'", Localize("xx", "world_format", "Test"))
	}
	if Localize("yy", "hello") != "Hello" { // Fallback to en for non-existent language "yy"
		t.Errorf("Expected 'Hello' (fallback for yy), got '%s'", Localize("yy", "hello"))
	}
	// For a key missing in "xx" but present in "en" (default)
	if Localize("xx", "only_in_en") != "Only English" { // Fallback to en
		t.Errorf("Expected 'Only English' (fallback for xx), got '%s'", Localize("xx", "only_in_en"))
	}
	// For a key missing in "xx" AND also missing in "en" (default)
	// The new error message format is "[translation missing for key: %s (in lang '%s' and default '%s')]"
	expectedMissingKeyInXxAndEn := "[translation missing for key: non_existent_key (in lang 'xx' and default 'en')]"
	if Localize("xx", "non_existent_key") != expectedMissingKeyInXxAndEn {
		t.Errorf("Expected '%s', got '%s'", expectedMissingKeyInXxAndEn, Localize("xx", "non_existent_key"))
	}
	// For a key missing in "yy" (non-existent lang) AND also missing in "en" (default)
	expectedMissingKeyInYyAndEn := "[translation missing for key: non_existent_key_in_yy (in lang 'yy' and default 'en')]"
	if Localize("yy", "non_existent_key_in_yy") != expectedMissingKeyInYyAndEn {
		t.Errorf("Expected '%s', got '%s'", expectedMissingKeyInYyAndEn, Localize("yy", "non_existent_key_in_yy"))
	}
	if Localize("en", "world_format", "Tester") != "World Tester" {
		t.Errorf("Expected 'World Tester', got '%s'", Localize("en", "world_format", "Tester"))
	}
	if Localize("xx", "only_in_xx") != "Only XX" {
		t.Errorf("Expected 'Only XX', got '%s'", Localize("xx", "only_in_xx"))
	}
	if Localize("xx", "only_in_en") != "Only English" { // Fallback to en
		t.Errorf("Expected 'Only English' (fallback for xx), got '%s'", Localize("xx", "only_in_en"))
	}
	// Test case for when default language itself is missing (simulated by clearing translations)
	translations = make(map[string]map[string]string) // Clear all, including default
	expectedMissingMsg := "[translation missing for lang: zz, key: any_key (default lang 'en' not loaded/found)]"
	if got := Localize("zz", "any_key"); got != expectedMissingMsg {
	    t.Errorf("Expected '%s' when default lang is missing, got '%s'", expectedMissingMsg, got)
	}

}

func TestAvailableLanguages(t *testing.T) {
	resetGlobals()

	enBytes, _ := json.Marshal(map[string]string{"k": "v"})
	xxBytes, _ := json.Marshal(map[string]string{"k": "v"})
	yyBytes, _ := json.Marshal(map[string]string{"k": "v"})

	tmpDir := setupTestLocalesWithFiles(t, map[string][]byte{
		"en.json": enBytes,
		"xx.json": xxBytes,
		"yy.json": yyBytes,
	})
	defer os.RemoveAll(tmpDir)

	if err := LoadTranslations(tmpDir); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	expectedLangs := []string{"en", "xx", "yy"}
	actualLangs := AvailableLanguages()
	sort.Strings(actualLangs) // Sort for consistent comparison
	sort.Strings(expectedLangs)

	if !reflect.DeepEqual(actualLangs, expectedLangs) {
		t.Errorf("Expected available languages %v, got %v", expectedLangs, actualLangs)
	}

	// Test with no translations loaded
	resetGlobals()
	if err := LoadTranslations("non_existent_dir_for_sure_again"); err != nil {
		t.Logf("LoadTranslations with non-existent directory returned error: %v", err)
	}
	if len(AvailableLanguages()) != 0 {
		t.Errorf("Expected 0 available languages, got %v", AvailableLanguages())
	}
}
