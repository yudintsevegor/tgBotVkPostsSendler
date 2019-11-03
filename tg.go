package tgBotVkPostSendler

import (
	"fmt"
	"log"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Caller struct {
	// Channelname is an unique identifier for the target chat or username
	// of the target channel (in the format @channelusername)
	ChannelName string
	// WebHookURL is a special URL which determines an address where telegram-bot is available.
	WebHookURL string
	Options    ReqOptions
	ErrChan    chan error
}

func (caller *Caller) CallBot(bot *tgbotapi.BotAPI, in <-chan string) error {
	channelName := caller.ChannelName
	webHookURL := caller.WebHookURL

	// bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	if _, err := bot.SetWebhook(tgbotapi.NewWebhook(webHookURL)); err != nil {
		caller.ErrChan <- err
	}

	updates := bot.ListenForWebhook("/")

	for {
		select {
		case text := <-in:
			if _, err := bot.Send(tgbotapi.NewMessageToChannel(channelName, text)); err != nil {
				log.Printf("Channel Name: %v, Error: %v", channelName, err)
			}
		case update := <-updates:
			log.Println(update.Message.Text)
			_, err := bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf("Bot is handler for %v channel", channelName),
			))
			if err != nil {
				log.Printf("Channel Name: %v, Error: %v", channelName, err)
			}
		}
	}
}
