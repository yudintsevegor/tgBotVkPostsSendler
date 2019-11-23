package tgBotVkPostSendler

import (
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Handler struct {
	Telegram Telegram

	// Options is a struct with request options to VK-API
	Options ReqOptions

	// Writer is a "client" for taking requests to DB
	DbWriter *DbWriter

	// TimeOut is a field which determins how often program asks database for getting old posts,
	// which are not published yet
	TimeOut time.Duration

	// ErrChan is a golang channel for sending errors
	ErrChan chan error

	// recipients determine usernames which get logs in private channel
	recipients map[string]int64
}

type Telegram struct {
	// Channelname is an unique identifier for the target chat or username
	// of the target channel (in the format @channelusername)
	ChannelName string
	// WebHookURL is a special URL which determines an address where telegram-bot is available
	WebHookURL string
	// BotToken is an unique identifier for bot, which gets from @BotFather.
	BotToken string

	bot *tgbotapi.BotAPI
}

// from VK API: https://vk.com/dev/wall.get
type ReqOptions struct {
	// Count is a number of records that you want to retrieve. Maximum value: 100
	Count string
	// Offset is a required to select a specific subset of records.
	Offset string
	// Filter determines what types of wall entries you want to retrieve.
	// Possible value:
	// suggestions	-suggested posts on the community wall (only available when called with access_token);
	// postponed	-deferred records (available only when called with access_token pass);
	// owner		— the record owner of the wall;
	// others		-entries are not from the wall owner;
	// all			-all entries on the wall (owner + others).
	// Default: all.
	Filter string
}
