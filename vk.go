package tgBotVkPostSendler

import (
	"errors"
	"log"
	"net/url"
	"strconv"
	"time"
)

// restrictions:
// wall.get — 5000 requests per day. -> ~1 req per 20seconds
// https://vk.com/dev/data_limits

const (
	version = "5.102"
	vkUrl   = "https://vk.com/"
	reqUrl  = "https://api.vk.com/method/wall.get?"
)

// from VK API: https://vk.com/dev/wall.get
type ReqOptions struct {
	// Count is a number of records you want to retrieve. Maximum value: 100
	Count string
	// Offset is a required to select a specific subset of records.
	Offset string
	// Filter determines what types of wall entries you want to retrieve. Possible value:
	// suggestions-suggested posts on the community wall (only available when called with access_token);
	// postponed-deferred records (available only when called with access_token pass);
	// owner — the record owner of the wall;
	// others-entries are not from the wall owner;
	// all-all entries on the wall (owner + others).
	// Default: all.
	Filter string
}

var mapFilter = map[string]struct{}{
	"suggestions": struct{}{},
	"postponed":   struct{}{},
	"owner":       struct{}{},
	"others":      struct{}{},
	"all":         struct{}{},
}

type Message struct {
	ID   string
	Text string
}

func (caller *Caller) GetVkPosts(groupID, serviceKey string) <-chan Message {
	count, _, err := caller.Options.validateOptions()
	if err != nil {
		caller.ErrChan <- err
	}

	u := url.Values{}
	u.Set("count", caller.Options.Count)
	u.Set("offset", caller.Options.Offset)
	u.Set("filter", caller.Options.Filter)

	u.Set("owner_id", groupID)
	u.Set("access_token", serviceKey)

	u.Set("v", version)
	u.Set("extended", "1") // is it really important?

	out := make(chan Message)
	const day = 24
	go caller.Writer.loop(caller.TimeOut/day, count, groupID, u, out)

	return out
}

func (opt *ReqOptions) validateOptions() (int, int, error) {
	count, err := strconv.Atoi(opt.Count)
	if err != nil {
		return 0, 0, err
	}

	offset, err := strconv.Atoi(opt.Offset)
	if err != nil {
		return 0, 0, err
	}

	if _, ok := mapFilter[opt.Filter]; !ok {
		return 0, 0, errors.New("unexpected filter in the request")
	}

	return count, offset, nil
}

func (w *Writer) loop(sleep time.Duration, count int, groupID string, u url.Values, out chan Message) {
	var isFirstReq = true

	path := reqUrl + u.Encode()

	for {
		if !isFirstReq {
			time.Sleep(sleep)
		}
		if isFirstReq {
			isFirstReq = false
		}

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

		ids, err := w.SelectRows()
		if err != nil {
			log.Println(err)
			continue
		}

		posts := getDifPosts(ids, body.Items)
		// send posts from the latest to the earliest
		for i := len(posts) - 1; i >= 0; i-- {
			w.id = string(posts[i].ID)
			w.text = makeMessage(posts[i], groupID)

			if err := w.InsertToDb(); err != nil {
				log.Println(err)
				continue
			}

			out <- Message{
				ID:   w.id,
				Text: w.text,
			}
		}
	}
}

func getDifPosts(ids map[string]struct{}, in []data) []data {
	out := make([]data, 0, len(in))
	for _, v := range in {
		if _, ok := ids[string(v.ID)]; ok {
			continue
		}
		out = append(out, v)
	}

	return out
}
