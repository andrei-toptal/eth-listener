package main

import (
	"log"
	"sync"

	"go.uber.org/atomic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram interface {
	Notify(message string)
	Close()
}

type noopTelegram struct{}

func (noopTelegram) Notify(string) {}
func (noopTelegram) Close()        {}

type telegram struct {
	bot        *tgbotapi.BotAPI
	stopCh     chan struct{}
	waitStop   sync.WaitGroup
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
		stopCh:   make(chan struct{}),
		username: config.Telegram.Username,
	}
	t.waitStop.Add(1)
	go t.updatesLoop()
	return t
}

func (t *telegram) Notify(message string) {
	if !t.subscribed.Load() {
		return
	}
	msg := tgbotapi.NewMessage(t.chatID, message)
	t.bot.Send(msg)
}

func (t *telegram) updatesLoop() {
	defer t.waitStop.Done()

	updates := t.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for {
		select {
		case <-t.stopCh:
			return
		case update := <-updates:
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
	}
}

func (t *telegram) Close() {
	close(t.stopCh)
	t.waitStop.Wait()
}
