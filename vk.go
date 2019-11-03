package tgBotVkPostSendler

import (
	"errors"
	"log"
	"net/url"
	"time"
)

// restrictions:
// wall.get — 5000 вызовов в сутки. -> ~1 req per 20seconds
// https://vk.com/dev/data_limits

const (
	version = "5.102"
	vkUrl   = "https://vk.com/"
	reqUrl  = "https://api.vk.com/method/wall.get?"
)

// Count is a number of records you want to retrieve. Maximum value: 100

// Offset is a required to select a specific subset of records.

// Filter determines what types of wall entries you want to retrieve. Possible value:
// suggestions-suggested posts on the community wall (only available when called with access_token);
// postponed-deferred records (available only when called with access_token pass);
// owner — the record owner of the wall;
// others-entries are not from the wall owner;
// all-all entries on the wall (owner + others).
// Default: all.

type ReqOptions struct {
	Count    string
	Offset   string
	Filter   string
	AllPosts bool
}

func (options *ReqOptions) GetVkPosts(groupID, serviceKey string) <-chan string {
	count := options.Count
	offset := options.Offset
	filter := options.Filter

	u := url.Values{}
	u.Set("count", count)
	u.Set("offset", offset)
	u.Set("filter", filter)

	u.Set("owner_id", groupID)
	u.Set("access_token", serviceKey)

	u.Set("v", version)
	u.Set("extended", "1") // is it really important?

	path := reqUrl + u.Encode()

	out := make(chan string)
	go func() {
		var isFirstReq = true
		var corner int
		var zeroLevel int

		for {
			body, err := getPosts(path)
			if err != nil {
				log.Println(err)
				continue
			}

			// only one groupID in request before
			if len(body.Groups) != 1 {
				log.Println(errors.New("empty info about group"))
				continue
			}

			corner = body.Count - zeroLevel
			if isFirstReq {
				isFirstReq = false
				zeroLevel = body.Count
				corner = len(body.Items)
			}

			// send posts from the latest to the earliest
			for i := corner - 1; i >= 0; i-- {
				out <- makeMessage(body.Items[i], groupID)
			}
			time.Sleep(20 * time.Second)
		}
	}()

	return out
}
