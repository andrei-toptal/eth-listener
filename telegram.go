package main

import (
	"log"

	"go.uber.org/atomic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram interface {
	Notify(message string)
}

type noopTelegram struct{}

func (noopTelegram) Notify(string) {}

type telegram struct {
	bot        *tgbotapi.BotAPI
	username   string
	subscribed atomic.Bool
	chatID     int64
}

func NewTelegram(config *Config) Telegram {
	if config.Telegram == nil {
		log.Println("Using no-op Telegram due to missing config")
		return noopTelegram{}
	}

	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}
	t := &telegram{
		bot:      bot,
		username: config.Telegram.Username,
	}
	go t.updatesLoop()
	return t
}

func (t telegram) Notify(message string) {
	if !t.subscribed.Load() {
		return
	}
	msg := tgbotapi.NewMessage(t.chatID, message)
	t.bot.Send(msg)
}

func (t *telegram) updatesLoop() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("Received message from %s: `%s`", update.Message.From.UserName, update.Message.Text)
		if t.username != "" && update.Message.From.UserName != t.username {
			continue
		}
		if update.Message.Text == "/subscribe" {
			t.chatID = update.Message.Chat.ID
			t.subscribed.Store(true)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are subscribed!")
			t.bot.Send(msg)

			log.Printf("Telegram Bot subscribed: %s", update.Message.From)
		}
		if update.Message.Text == "/unsubscribe" {
			t.subscribed.Store(false)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are unsubscribed.")
			t.bot.Send(msg)

			log.Printf("Telegram Bot unsubscribed: %s", update.Message.From)
		}
	}

	log.Fatal("Telegram bot has been disconnected")
}
