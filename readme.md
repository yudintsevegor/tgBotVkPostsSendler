# TelegramBot + VK-API
Easy util for getting posts from vk communities by them ID and send the posts to telegram channel by it ID.

## Install
`go get github.com/yudintsevegor/tgBotVkPostsSendler`

## Example of Usage
``` go
package main

import (
	sendler "github.com/yudintsevegor/tgBotVkPostsSendler"
	// other
)

func main() {
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		// error handler
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		// error handler
	}

	w := sendler.Writer{
		DB:        db,
		TableName: "tableName",
	}

	if _, err = w.CreateTable(); err != nil {
		// error handler
	}

	port := "8080"
	go http.ListenAndServe(":"+port, nil)
	fmt.Printf("start listen :%v\n", port)

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		// error handler
	}

	groupID := "groupID"
	channelName := "@channelName"
	webHookURL := "webHook"

	opt := sendler.ReqOptions{
		Count:  "10",
		Offset: "0",
		Filter: "owner",
	}

	handler := sendler.Handler{
		ChannelName: channelName,
		WebHookURL:  webHookURL,
		Options:     opt,
		ErrChan:     make(chan error),

		TimeOut: time.Hour * 24,
		Writer:  &w,
	}

	go handler.StartBot(bot, handler.GetVkPosts(groupID, ServiceKey))

	for err := range handler.ErrChan {
		// error handler
	}
}

```

## Packages
* [go-telegram-bot-api](gopkg.in/telegram-bot-api.v4)
* [vk-api](https://vk.com/dev/)
* standarts golang libs
