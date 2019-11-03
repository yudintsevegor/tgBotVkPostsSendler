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
	port := "8080"
	go http.ListenAndServe(":"+port, nil)
	fmt.Printf("start listen :%v\n", port)

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
	}

	groupID := "groupID"
	channelName := "@channelName"
	webHookURL := "WebHook"

	opt := sendler.ReqOptions{
		Count:  "5",
		Offset: "0",
		Filter: "owner",
	}

	caller := sendler.Caller{
		ChannelName: channelName,
		WebHookURL:  webHookURL,
		Options:     opt,
		ErrChan:     make(chan error),
	}


    // if you want to use VK-API,
	// you must get ServiceKey for your application
	
	go caller.CallBot(bot, caller.GetVkPosts(groupID, ServiceKey))

	for err := range caller.ErrChan {
		log.Fatal(err)
	}
}
```

## Packages
* [go-telegram-bot-api](gopkg.in/telegram-bot-api.v4)
* [vk-api](https://vk.com/dev/)
* standarts golang libs
