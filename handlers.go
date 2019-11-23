package tgBotVkPostSendler

import (
	"fmt"
	"log"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func (h *Handler) handlePosts(msg Message, channelName string) {
	if msg.Error != nil {
		h.errorLogging(msg.Error.Error())
		return
	}

	if err := h.sendMsgToChannel(channelName, msg); err != nil {
		h.errorLogging(fmt.Sprintf("[ERR] Channel Name: %v, Error: %v", channelName, err))
	}
}

func (h *Handler) handleFailedPosts(channelName string) {
	messages, err := h.DbWriter.SelectFailedRows()
	if err != nil {
		h.errorLogging(fmt.Sprintf("[ERR] Trying to select old rows. Error: %v", err))
		return
	}

	for _, message := range messages {
		if err := h.sendMsgToChannel(channelName, message); err != nil {
			h.errorLogging(fmt.Sprintf("[ERR] Channel Name: %v, Error: %v", channelName, err))
		}
	}
}

func (h *Handler) handleMessages(message *tgbotapi.Message, channelName string) {
	username := message.Chat.UserName
	chatId := message.Chat.ID
	if _, ok := h.recipients[username]; ok {
		text := h.manageCommands(username, message.Text, chatId)
		h.Telegram.bot.Send(tgbotapi.NewMessage(chatId, text))
		return
	}

	if _, err := h.Telegram.bot.Send(
		tgbotapi.NewMessage(chatId, fmt.Sprintf("Bot is handler for %v channel", channelName)),
	); err != nil {
		h.errorLogging(fmt.Sprintf("[ERR] Channel Name: %v, Error: %v", channelName, err))
	}
}

const (
	unsetlogs = "/unsetlogs"
	setlogs   = "/setlogs"
)

func (h *Handler) manageCommands(username, text string, chatId int64) string {
	switch text {
	case unsetlogs:
		h.recipients[username] = 0
		return fmt.Sprintf("%s for user %s", unsetlogs, username)
	case setlogs:
		h.recipients[username] = chatId
		return fmt.Sprintf("%s for user %s", setlogs, username)
	default:
		return "unknown command"
	}
}

func (h *Handler) errorLogging(text string) {
	log.Println(text)

	if len(h.recipients) == 0 {
		return
	}

	// sending msgs only for users with chatId
	for _, chatId := range h.recipients {
		if chatId == 0 {
			continue
		}

		h.Telegram.bot.Send(tgbotapi.NewMessage(chatId, text))
	}
}

func (h *Handler) sendMsgToChannel(channelName string, msg Message) error {
	if _, err := h.Telegram.bot.Send(tgbotapi.NewMessageToChannel(channelName, msg.Text)); err != nil {
		return err
	}

	if err := h.DbWriter.UpdateStatus(msg.ID); err != nil {
		return err
	}

	return nil
}
