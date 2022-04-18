package telegram

import (
	"log"

	"github.com/pinebit/eth-listener/config"
	"go.uber.org/atomic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram interface {
	Notify(message string)
}

type telegram struct {
	bot        *tgbotapi.BotAPI
	username   string
	subscribed atomic.Bool
	chatID     int64
}

func NewTelegram(cfg *config.Telegram) Telegram {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatal(err)
	}
	t := &telegram{
		bot:      bot,
		username: cfg.Username,
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
		if t.username != "" && update.Message.From.UserName != t.username {
			continue
		}
		if update.Message.Text == "/subscribe" {
			t.chatID = update.Message.Chat.ID
			t.subscribed.Store(true)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are subscribed!")
			t.bot.Send(msg)
		}
		if update.Message.Text == "/unsubscribe" {
			t.subscribed.Store(false)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are unsubscribed.")
			t.bot.Send(msg)
		}
	}

	log.Fatal("Telegram bot has been disconnected")
}
