package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Message struct {
	URL                 string `json:"url"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes"`
}

type shortlinkResp struct {
	ShortLink string `json:"shortLink"`
}

func main() {
	url1 := "http://localhost:8080/api/shorten"
	body := getData()
	shortLink := TestPost(url1, body)

	url2 := "http://localhost:8080/api/info?shortLink=" + shortLink
	TestGet(url2)

	url3 := "http://localhost:8080/" + shortLink
	TestGet(url3)
}

func getData() io.Reader {
	m := Message{URL: "http://www.baidu.com", ExpirationInMinutes: 30}
	jsonMessage, err := json.Marshal(m)
	if err != nil {
		log.Println("json format error:", err)
		return nil
	}

	body := bytes.NewBuffer(jsonMessage)
	return body
}

func TestPost(url string, data io.Reader) string {
	resp, err := http.Post(url, "application/json; charset=utf-8", data)

	if err != nil {
		log.Println(err)
	}

	body := resp.Body
	defer body.Close()

	ret, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(ret))

	m := shortlinkResp{}
	err1 := json.Unmarshal(ret, &m)
	if err1 != nil {
		log.Println(err1)
	}

	return m.ShortLink
}

func TestGet(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}

	body := resp.Body

	defer body.Close()

	ret, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(ret))
}
