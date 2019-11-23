package tgBotVkPostSendler

import (
	"errors"
	"net/url"
	"time"
)

const (
	version = "5.102"
	vkUrl   = "https://vk.com/"
	reqUrl  = "https://api.vk.com/method/wall.get?"
)

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

func (h *Handler) GetVkPosts(groupID, vkServiceKey string) <-chan Message {
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
	u.Set("access_token", vkServiceKey)

	u.Set("v", version)
	u.Set("extended", "1") // is it really important?

	path := reqUrl + u.Encode()
	go h.loop(groupID, path, out)

	return out
}

// restrictions:
// wall.get — 5000 requests per day. -> ~1 req per 20seconds
// https://vk.com/dev/data_limits
var timeout time.Duration = 30 * time.Second // TODO: tmp solution

func (h *Handler) loop(groupID, path string, out chan Message) {
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

		ids, err := h.DbWriter.SelectCompletedRows()
		if err != nil {
			out <- Message{Error: err}
			continue
		}

		posts := getDiffPosts(ids, body.Items)

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

func getDiffPosts(ids map[string]struct{}, input []data) []data {
	out := make([]data, 0, len(input))
	for _, v := range input {
		if _, ok := ids[string(v.ID)]; ok {
			continue
		}
		out = append(out, v)
	}

	return out
}
