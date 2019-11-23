package tgBotVkPostSendler

import (
	"errors"
	"net/url"
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

var mapFilter = map[string]struct{}{
	"suggestions": struct{}{},
	"postponed":   struct{}{},
	"owner":       struct{}{},
	"others":      struct{}{},
	"all":         struct{}{},
}

type Message struct {
	ID    string
	Text  string
	Error error
}

func (h *Handler) GetVkPosts(groupID, serviceKey string) <-chan Message {
	out := make(chan Message)
	if _, ok := mapFilter[h.Options.Filter]; !ok {
		h.ErrChan <- errors.New("unexpected vk-filter in the request")
		return out
	}

	u := url.Values{}
	u.Set("count", h.Options.Count)
	u.Set("offset", h.Options.Offset)
	u.Set("filter", h.Options.Filter)

	u.Set("owner_id", groupID)
	u.Set("access_token", serviceKey)

	u.Set("v", version)
	u.Set("extended", "1") // is it really important?

	const day = 24 // TODO: FIX IT
	path := reqUrl + u.Encode()
	go h.loop(h.TimeOut/day, groupID, path, out)

	return out
}

var timeout time.Duration = 30 * time.Second

func (h *Handler) loop(sleep time.Duration, groupID, path string, out chan Message) {
	var isFirstReq = true

	for {
		if !isFirstReq {
			time.Sleep(timeout)
		}
		if isFirstReq {
			isFirstReq = false
		}

		body, err := getPosts(path)
		if err != nil {
			out <- Message{Error: err}
			continue
		}

		// only one groupID in request before
		if len(body.Groups) != 1 {
			out <- Message{Error: errors.New("empty info about group")}
			continue
		}

		ids, err := h.DbWriter.SelectRows()
		if err != nil {
			out <- Message{Error: err}
			continue
		}

		posts := getDifPosts(ids, body.Items)

		// send posts from the latest to the earliest
		for i := len(posts) - 1; i >= 0; i-- {
			h.DbWriter.id = string(posts[i].ID)
			h.DbWriter.text = makeMessage(posts[i], groupID)

			if err := h.DbWriter.InsertToDb(); err != nil {
				out <- Message{Error: err}
				continue
			}

			out <- Message{
				ID:   h.DbWriter.id,
				Text: h.DbWriter.text,
			}
		}
	}
}

func getDifPosts(ids map[string]struct{}, input []data) []data {
	out := make([]data, 0, len(input))
	for _, v := range input {
		if _, ok := ids[string(v.ID)]; ok {
			continue
		}
		out = append(out, v)
	}

	return out
}
