package tgBotVkPostSendler

import (
	"fmt"
	"log"
	"net/http"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// Channelname os an unique identifier for the target chat or username
// of the target channel (in the format @channelusername)

// WebHookURL is a special URL which determines a address where telegram-bot is available.
type Params struct {
	ChannelName string
	WebHookURL  string
}

func (params *Params) CallBot(bot *tgbotapi.BotAPI, in <-chan string) error {
	channelName := params.ChannelName
	webHookURL := params.WebHookURL

	// bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err := bot.SetWebhook(tgbotapi.NewWebhook(webHookURL))
	if err != nil {
		return err
	}

	updates := bot.ListenForWebhook("/")

	port := "8080"
	go http.ListenAndServe(":"+port, nil)
	fmt.Printf("start listen :%v", port)

	go func() {
		for {
			select {
			case text := <-in:
				if _, err := bot.Send(tgbotapi.NewMessageToChannel(channelName, text)); err != nil {
					log.Println(err)
				}
			case update := <-updates:
				_, err = bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					fmt.Sprintf("Bot is handler for %v channel", channelName),
				))
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	return nil
}
