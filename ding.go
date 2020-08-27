package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RobotMsg struct {
	Msgtype string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

func base64Encode(msg []byte) string {
	return base64.StdEncoding.EncodeToString(msg)
}

func hmacSha256(key string, msg string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	_, _ = h.Write([]byte(msg))
	return h.Sum(nil)
}

func robot(cfg ChopperCfg) error {
	if cfg.Robot.Name == "" || cfg.Robot.Content == "" {
		return nil
	}

	var secret string
	var token string
	if cfg.Robot.Name == "游戏开发" {
		secret = "SECcfbe68b5435ea54b2435613b1789573f3b56e6fb57e360cc440474a31d1473c6"
		token = "5fe90d3605f4a9b5532ab7c6011f5c06132c22ed528de31779c4d440371114af"
	}

	if secret == "" || token == "" {
		return nil
	}

	timeStamp := time.Now().Unix() * 1000
	signStr := fmt.Sprintf("%d\n%s", timeStamp, secret)
	signSha256 := hmacSha256(secret, signStr)
	signBase64 := base64Encode(signSha256)
	signEncode := url.QueryEscape(signBase64)

	url := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s&timestamp=%d&sign=%s", token, timeStamp, signEncode)
	robotMsg := &RobotMsg{}
	robotMsg.Msgtype = "text"
	robotMsg.Text.Content = cfg.Robot.Content
	content, err := json.Marshal(robotMsg)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", strings.NewReader(string(content)))
	return err
}
