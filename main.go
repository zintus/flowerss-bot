package main

import (
	"os"
	"os/signal"
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

	if err := i18n.LoadTranslations("locales"); err != nil {
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
