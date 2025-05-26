package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/zintus/flowerss-bot/internal/bot"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/log"
	"github.com/zintus/flowerss-bot/internal/scheduler"
)

func main() {
	appCore := core.NewCoreFormConfig()
	if err := appCore.Init(); err != nil {
		log.Fatal(err)
	}

	// --- debug info for translation loading ---
	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Unable to get current working directory: %v", err)
	} else {
		log.Infof("Current working directory: %s", cwd)
	}

	// first try the path used inside the image
	localeDir := "/opt/flowerss/locales"
	// when running outside a container fall back to project-relative “locales”
	if _, err := os.Stat(localeDir); os.IsNotExist(err) {
		localeDir = "locales"
	}
	absLocaleDir, err := filepath.Abs(localeDir)
	if err != nil {
		log.Errorf("Unable to resolve absolute path for %q: %v", localeDir, err)
		absLocaleDir = localeDir // fall-back
	}
	if _, err := os.Stat(absLocaleDir); os.IsNotExist(err) {
		log.Infof("Locale directory DOES NOT exist: %s", absLocaleDir)
	} else if err != nil {
		log.Errorf("Error checking locale directory: %v", err)
	} else {
		log.Infof("Locale directory exists: %s", absLocaleDir)
	}

	if err := i18n.LoadTranslations(absLocaleDir); err != nil {
		log.Fatalf("Failed to load translations: %v", err)
	}
	log.Infof("Translations loaded.")

	go handleSignal()
	b := bot.NewBot(appCore)

	task := scheduler.NewRssTask(appCore)
	task.Register(b)
	task.Start()
	if err := b.Run(); err != nil {
		log.Fatalf("Failed to run bot: %v", err)
	}
}

func handleSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-c

	os.Exit(0)
}
