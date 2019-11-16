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
	// ErrChan is a golang channel for sending error
	ErrChan chan error
	// TimeOut is a field which determins how often program asks database for get old posts,
	// which are not published yet
	TimeOut time.Duration

	// Writer is a "client" for taking requests to DB
	Writer *Writer
}

func (handler *Handler) StartBot(bot *tgbotapi.BotAPI, in <-chan Message) error {
	channelName := handler.ChannelName
	webHookURL := handler.WebHookURL

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	if _, err := bot.SetWebhook(tgbotapi.NewWebhook(webHookURL)); err != nil {
		handler.ErrChan <- err
	}

	updates := bot.ListenForWebhook("/")
	w := handler.Writer

	for {
		select {
		case mes := <-in:
			if err := w.sendMessage(bot, channelName, mes.ID, mes.Text); err != nil {
				log.Printf("Channel Name: %v, Error: %v", channelName, err)
			}
		case <-time.After(handler.TimeOut):
			messages, err := w.SelectOldRows()
			if err != nil {
				log.Println(err)
				continue
			}

			for _, message := range messages {
				if err := w.sendMessage(bot, channelName, message.ID, message.Text); err != nil {
					log.Printf("Channel Name: %v, Error: %v", channelName, err)
				}
			}
		case update := <-updates:
			_, err := bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf("Bot is handler for %v channel", channelName),
			))

			if err != nil {
				log.Printf("Channel Name: %v, Error: %v", channelName, err)
				continue
			}
		}
	}
}

func (w *Writer) sendMessage(bot *tgbotapi.BotAPI, channelName, id, text string) error {
	if _, err := bot.Send(tgbotapi.NewMessageToChannel(channelName, text)); err != nil {
		return err
	}

	if err := w.UpdateStatus(id); err != nil {
		return err
	}

	return nil
}
