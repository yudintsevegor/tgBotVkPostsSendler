# TelegramBot + VK-API
Easy util for getting posts from vk communities by them ID and send the posts to telegram channel by it ID.

## Install
`go get github.com/yudintsevegor/tgBotVkPostsSendler`

## Packages
* [go-telegram-bot-api](gopkg.in/telegram-bot-api.v4)
* [vk-api](https://vk.com/dev/)

## Restrictions
* Only PostgresSQL is supporting now

## Usage
##### 1. Creation telegram-bot using @BotFather and getting BotToken
Firstly, you need to find @BotFather in telegram and create telgram bot. On final step ypu will see:
```
Done! Congratulations on your new bot. You will find it at t.me/BotName. You can now add a description,
about section and profile picture for your bot, see /help for a list of commands. By the way, when you've
finished creating your cool bot, ping our Bot Support if you want a better username for it. Just make sure
the bot is fully operational before you do this.

Use this token to access the HTTP API:
`BotToken`
Keep your token secure and store it safely, it can be used by anyone to control your bot.

For a description of the Bot API, see this page: https://core.telegram.org/bots/api
```
BotToken is one of the ssential requirements for using package.

##### 2. Creation vk application and getting VkServiceKey
Secondly, you need to get VkServiceKey using [vk-dev]https://vk.com/editapp?act=create). In `Settings` after creation you can find `Service token`. Get it and it is a VkServiceKey.
##### 3. Find vk-community and getting its ID.
Thirdly, you need to visit a vk-community, find a random post and follow the link with time, how long time ago post was created. In URL you can see 
```
https://vk.com/ustami_msu?w=wallGroupID_46014
```
GroupID is what you need in format `-numbers`.
##### 4. The last steps
Finaly, create in telegram a channel and add a bot in this channel. Using channel name in format `@channelName`, set settings. Example of usage below.

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

	groupID := "groupID"
	channelName := "@channelName"
	webHookURL := "webHook"

	telegram := sendler.Telegram{
		ChannelName: channelName,
		WebHookURL:  webHookURL,
		BotToken:    BotToken,
	}

	opt := sendler.ReqOptions{
		Count:  "10",
		Offset: "0",
		Filter: "owner",
	}

	handler := sendler.Handler{
		Telegram: telegram,
		Options:  opt,
		ErrChan:  make(chan error),

		TimeOut:  time.Hour * 24,
		DbWriter: &w,
	}

	recipients := []string{"telegramUserName"}
	handler.GetRecipients(recipients)

	go handler.StartBot(handler.GetVkPosts(groupID, VkServiceKey))

	for err := range handler.ErrChan {
		// error handler
	}
}

```