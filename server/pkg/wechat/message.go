package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/neinei960/cat/server/config"
)

var (
	accessToken     string
	tokenExpireAt   time.Time
	tokenMu         sync.Mutex
)

func getAccessToken() (string, error) {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	if accessToken != "" && time.Now().Before(tokenExpireAt) {
		return accessToken, nil
	}

	cfg := config.AppConfig.WeChat
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		cfg.AppID, cfg.AppSecret,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	json.Unmarshal(body, &result)
	if result.ErrCode != 0 {
		return "", fmt.Errorf("get access token failed: %d %s", result.ErrCode, result.ErrMsg)
	}

	accessToken = result.AccessToken
	tokenExpireAt = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	return accessToken, nil
}

type TemplateMessage struct {
	ToUser     string                       `json:"touser"`
	TemplateID string                       `json:"template_id"`
	Page       string                       `json:"page,omitempty"`
	Data       map[string]map[string]string `json:"data"`
}

func SendTemplateMessage(msg *TemplateMessage) error {
	token, err := getAccessToken()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", token)
	body, _ := json.Marshal(msg)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	json.Unmarshal(respBody, &result)
	if result.ErrCode != 0 {
		return fmt.Errorf("send message failed: %d %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
