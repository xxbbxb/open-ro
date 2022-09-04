package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Update struct {
	Id  int64   `json:"update_id"`
	Msg Message `json:"message"`
}

type Message struct {
	Id   int64  `json:"message_id"`
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	User User   `json:"from,omitempty"` // empty for channels
}

type Chat struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

type User struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"` // also optional
	IsBot     bool   `json:"is_bot"`
}

type WebhookInfo struct {
	Url            string   `json:"url"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
	SecretToken    string   `json:"secret_token,omitempty"`
}

type MessageToSend struct {
	ChatId                int64  `json:"chat_id"`
	Text                  string `json:"text"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
	ParseMode             string `json:"parse_mode"`
}

type FileToUpload struct {
	Name   string
	Reader io.Reader
}

type TelegramBot struct {
	client        *http.Client
	token         string
	apiBaseUrl    string
	webhookSecret string
}

func NewBot(u string, token string, webhookUrl string) (*TelegramBot, error) {
	if webhookUrl == "" {
		return nil, fmt.Errorf("webhook url not specified")
	}
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("telegram token not specified")
	}
	secretBytes := sha256.Sum256([]byte(token))
	bot := &TelegramBot{
		client: &http.Client{
			Transport: &http.Transport{},
		},
		token:         token,
		apiBaseUrl:    u,
		webhookSecret: hex.EncodeToString(secretBytes[:]),
	}
	h, err := bot.GetWebhookInfo()
	if err != nil {
		panic(err)
	}
	if h.Url == "" {
		err = bot.SetWebhook(WebhookInfo{
			Url:            webhookUrl,
			AllowedUpdates: []string{"message"},
			SecretToken:    bot.webhookSecret,
		})
		if err != nil {
			panic(err)
		}
	}
	return bot, nil
}

func (t *TelegramBot) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Telegram-Bot-Api-Secret-Token") == t.webhookSecret {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
		}
		u := Update{}
		err = json.Unmarshal(body, &u)
		if err != nil {
			http.Error(w, "can't parse json", http.StatusBadRequest)
		}
		fmt.Println(string(body))
		err = t.SendMessage(MessageToSend{
			ChatId:                u.Msg.Chat.Id,
			Text:                  fmt.Sprintf("%d, %s: ```\n%v\n```", u.Msg.Chat.Id, u.Msg.Text, u.Msg.User),
			DisableWebPagePreview: true,
			ParseMode:             "MarkdownV2",
		})
		fmt.Println(err)
	} else {
		http.Error(w, "missing X-Telegram-Bot-Api-Secret-Token token", http.StatusBadRequest)
		return
	}
	w.Write([]byte("200 Ok"))
}

func (t *TelegramBot) SendMessage(m MessageToSend) error {
	const request = "sendMessage"
	_, err := t.Do("POST", request, m, nil)
	return err
}

func (t *TelegramBot) GetWebhookInfo() (*WebhookInfo, error) {
	const request = "getWebhookInfo"

	b, err := t.Do("GET", request, nil, nil)
	if err != nil {
		return nil, err
	}
	var h WebhookInfo
	err = json.Unmarshal(b, &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (t *TelegramBot) SetWebhook(h WebhookInfo) error {
	const request = "setWebhook"
	/*if certReader != nil {
		files := []FileToUpload{
			{
				Name:   "certificate",
				Reader: certReader,
			},
		}
		_, err := t.Do("POST", request, h, files)
		return err
	} else {
	}
	*/
	_, err := t.Do("POST", request, h, nil)
	return err
}

func (t *TelegramBot) Do(method string, apiMethod string, payload interface{}, files []FileToUpload) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var payloadBytes []byte
	var contentType string
	if len(files) > 0 {
		var buf bytes.Buffer
		fw := multipart.NewWriter(&buf)
		for f, v := range getTagValues(payload) {
			err := fw.WriteField(f, v)
			if err != nil {
				return nil, fmt.Errorf("unable to add field %s to multipart/form-data", f)
			}
		}
		for _, f := range files {
			w, err := fw.CreateFormFile(f.Name, f.Name)
			if err != nil {
				return nil, fmt.Errorf("unable to add file %s to multipart/form-data", f.Name)
			}
			io.Copy(w, f.Reader)
		}
		fw.Close()
		contentType = fw.FormDataContentType()
		payloadBytes = buf.Bytes()
	} else {
		contentType = "application/json"
		if payload != nil {
			payloadBytes, _ = json.Marshal(payload)
		}
	}
	fmt.Println(string(payloadBytes))
	r, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/bot%s/%s", t.apiBaseUrl, t.token, apiMethod), bytes.NewReader(payloadBytes))

	if err != nil {
		return nil, fmt.Errorf("failed to prepare %s /%s request", method, apiMethod)
	}
	r.Header.Add("accept", "application/json")
	r.Header.Add("content-type", contentType)

	resp, err := t.client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to do %s /%s", method, apiMethod)
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read %s /%s server response", method, apiMethod)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("telegram api returned error %d: %s", resp.StatusCode, string(buf))
	}
	return buf, nil
}

func getTagValues(v interface{}) map[string]string {
	tagValues := make(map[string]string)
	fields := reflect.VisibleFields(reflect.TypeOf(v))
	value := reflect.ValueOf(v)
	for _, f := range fields {
		if key, exists := f.Tag.Lookup("json"); exists {
			if i := strings.Index(key, ","); i > 0 {
				key = key[0:i]
			}
			tagValue := value.FieldByName(f.Name).Interface()
			if f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice {
				bytes, _ := json.Marshal(tagValue)
				tagValues[key] = string(bytes)
			} else {
				valStr := fmt.Sprintf("%v", tagValue)
				if valStr == "" {
					continue
				}
				tagValues[key] = valStr
			}
		}
	}
	return tagValues
}
