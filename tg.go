package tgBotVkPostSendler

import (
	"fmt"
	"log"
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Handler struct {
	// Channelname is an unique identifier for the target chat or username
	// of the target channel (in the format @channelusername)
	ChannelName string
	// WebHookURL is a special URL which determines an address where telegram-bot is available
	WebHookURL string
	// Options is a struct with request options to VK-API
	Options ReqOptions
	// ErrChan is a golang channel for sending errors
	ErrChan chan error
	// Recipients determine usernames which get logs in private channel
	recipients map[string]int64
	// TimeOut is a field which determins how often program asks database for getting old posts,
	// which are not published yet
	TimeOut time.Duration

	// Writer is a "client" for taking requests to DB
	DbWriter *DbWriter
}

func (h *Handler) GetRecipients(input []string) {
	h.recipients = make(map[string]int64, len(input))
	for _, recipients := range input {
		h.recipients[recipients] = 0
	}
}

func (h *Handler) StartBot(bot *tgbotapi.BotAPI, in <-chan Message) {
	channelName := h.ChannelName
	webHookURL := h.WebHookURL

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	if _, err := bot.SetWebhook(tgbotapi.NewWebhook(webHookURL)); err != nil {
		h.ErrChan <- err
		return
	}

	updates := bot.ListenForWebhook("/")
	w := h.DbWriter

	for {
		select {
		case mes := <-in:
			if mes.Error != nil {
				h.errorLogging(bot, mes.Error.Error())
			}

			if err := w.sendMessage(bot, channelName, mes.ID, mes.Text); err != nil {
				h.errorLogging(bot, fmt.Sprintf("[ERR] Channel Name: %v, Error: %v", channelName, err))
			}
		case <-time.After(h.TimeOut):
			messages, err := w.SelectOldRows()
			if err != nil {
				h.errorLogging(bot, fmt.Sprintf("[ERR] Trying to select old rows. Error: %v", err))
				continue
			}

			for _, message := range messages {
				if err := w.sendMessage(bot, channelName, message.ID, message.Text); err != nil {
					h.errorLogging(bot, fmt.Sprintf("[ERR] Channel Name: %v, Error: %v", channelName, err))
				}
			}
		case update := <-updates:
			_, err := bot.Send(
				tgbotapi.NewMessage(update.Message.Chat.ID,
					fmt.Sprintf("Bot is handler for %v channel", channelName),
				),
			)

			username := update.Message.Chat.UserName
			if _, ok := h.recipients[username]; ok {
				h.manageCommands(username, update.Message.Text, update.Message.Chat.ID)
			}

			if err != nil {
				h.errorLogging(bot, fmt.Sprintf("[ERR] Channel Name: %v, Error: %v", channelName, err))
				continue
			}
		}
	}
}

func (h *Handler) manageCommands(username, text string, chatId int64) {
	switch text {
	case "/unsetlog":
		h.recipients[username] = 0
	case "/setlog":
		h.recipients[username] = chatId
	}
}

func (h *Handler) errorLogging(bot *tgbotapi.BotAPI, text string) {
	log.Println(text)

	if len(h.recipients) == 0 {
		return
	}

	for _, chatId := range h.recipients {
		if chatId == 0 {
			continue
		}

		bot.Send(tgbotapi.NewMessage(chatId, text))
	}
}

func (w *DbWriter) sendMessage(bot *tgbotapi.BotAPI, channelName, id, text string) error {
	if _, err := bot.Send(tgbotapi.NewMessageToChannel(channelName, text)); err != nil {
		return err
	}

	if err := w.UpdateStatus(id); err != nil {
		return err
	}

	return nil
}
