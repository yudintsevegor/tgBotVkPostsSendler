package tgBotVkPostSendler

import (
	"log"
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func (h *Handler) GetRecipients(input []string) {
	h.recipients = make(map[string]int64, len(input))
	for _, recipients := range input {
		h.recipients[recipients] = 0
	}
}

func (h *Handler) StartBot(in <-chan Message) error {
	if err := h.Telegram.createBot(); err != nil {
		return err
	}
	bot := h.Telegram.bot

	channelName := h.Telegram.ChannelName
	updates := bot.ListenForWebhook("/")

	for {
		select {
		case msg := <-in:
			h.handlePosts(msg, channelName)
		case <-time.After(h.TimeOut):
			h.handleFailedPosts(channelName)
		case update := <-updates:
			h.handleMessages(update.Message, channelName)
		}
	}
}

func (tg *Telegram) createBot() error {
	bot, err := tgbotapi.NewBotAPI(tg.BotToken)
	if err != nil {
		return err
	}

	if _, err := bot.SetWebhook(tgbotapi.NewWebhook(tg.WebHookURL)); err != nil {
		return err
	}
	tg.bot = bot

	log.Printf("Authorized on account %s\n", bot.Self.UserName)
	return nil
}
