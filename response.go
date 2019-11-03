package tgBotVkPostSendler

import (
	"encoding/json"
	"strconv"
)

type Response struct {
	Body Body `json:"response"`
}

type Body struct {
	Count  int     `json:count`
	Items  []Data  `json:"items"`
	Groups []Group `json:"groups`
}

type Data struct {
	ID ID `json:"id"`

	Text string `json::text`
}

type Group struct {
	ID         int    `json:"id"`
	ScreenName string `json:screen_name`
}

type ID string

func (id *ID) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}

	*id = ID(strconv.Itoa(i))

	return nil
}
