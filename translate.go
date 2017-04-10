package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type TranslateRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Text   string `json:"text"`
}

func Translate(source, target, text, user, pass string) string {
	translateRequest := TranslateRequest{
		Source: source,
		Target: target,
		Text:   text,
	}
	requestData, _ := json.Marshal(translateRequest)

	translateUrl := constructUrl(source, target)
	req, err := http.NewRequest("POST", translateUrl, bytes.NewBuffer(requestData))
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return ""
	}

	req.SetBasicAuth(user, pass)
	req.Header.Add("content-type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Copy: ", err)
		return ""
	}
	return string(body)
}

func constructUrl(source, target string) string {
	url := "https://gateway.watsonplatform.net/language-translator/api/v2/translate"
	return url
}
