package tgBotVkPostSendler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type response struct {
	Body body `json:"response"`
}

type body struct {
	Count  int     `json:"count"`
	Groups []group `json:"groups"`
	Items  []data  `json:"items"`
}

type group struct {
	ID         int    `json:"id"`
	ScreenName string `json:"screen_name"`
}

type data struct {
	ID   ID     `json:"id"`
	Text string `json:"text"`
	Date Time   `json:"date"`
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

type Time time.Time

func (t *Time) UnmarshalJSON(b []byte) error {
	var i int64
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}

	*t = Time(time.Unix(i, 0))

	return nil
}

func getPosts(path string) (body, error) {
	resp, err := http.Get(path)
	if err != nil {
		return body{}, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body{}, err
	}

	r := new(response)
	if err := json.Unmarshal(b, r); err != nil {
		return body{}, err
	}

	return r.Body, nil
}

func makeMessage(d data, groupID string) string {
	return d.Text + "\n\n" + makeLink(string(d.ID), groupID)
}

func makeLink(id, groupID string) string {
	return vkUrl + "wall" + groupID + "_" + id
}
