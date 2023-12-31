package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sashabaranov/go-openai"
)

var c atomic.Value

const configFile = "/data/qq-bot.json"

func init() {
	c.Store(mappingKV)
	if Pod() != "" {
		file, err := os.ReadFile(configFile)
		if err == nil {
			v := KV{}
			json.Unmarshal(file, &v)
			Set(v)
		}
	}
}

func Configs() KV {
	return c.Load().(KV)
}

type KV map[string]string

func (k KV) String() string {
	var s string
	for key, value := range k {
		s += fmt.Sprintf("%s=%s\n", key, value)
	}
	return s
}

func PixivSession() string {
	return c.Load().(KV)["pixiv_session"]
}
func AiBrowserModel() string {
	return c.Load().(KV)["ai_browser_model"]
}

func AiProxyUrl() string {
	return c.Load().(KV)["ai_browser_proxy_url"]
}

func AiToken() string {
	return c.Load().(KV)["ai_token"]
}

func AiMode() string {
	return c.Load().(KV)["ai_mode"]
}

func ChatGPTApiModel() string {
	return c.Load().(KV)["chatgpt_model"]
}
func AzureToken() string {
	return c.Load().(KV)["azure_token"]
}
func AzureModel() string {
	return c.Load().(KV)["azure_model"]
}
func AzureUrl() string {
	return c.Load().(KV)["azure_url"]
}

func AiAccessToken() string {
	return c.Load().(KV)["ai_browser_access_token"]
}

func PixivMode() string {
	return c.Load().(KV)["pixiv_mode"]
}

func GroupID() string {
	return c.Load().(KV)["group_id"]
}

func UserID() string {
	return c.Load().(KV)["user_id"]
}

func BiliCookie() string {
	return c.Load().(KV)["bili_cookie"]
}

func Namespace() string {
	return c.Load().(KV)["namespace"]
}

func Pod() string {
	return c.Load().(KV)["pod_name"]
}

func WeatherKey() string {
	return c.Load().(KV)["weather_key"]
}

func GeoKey() string {
	return c.Load().(KV)["weather_key"]
}

func TianApiKey() string {
	return c.Load().(KV)["tian_api_key"]
}

func HttpProxy() string {
	return c.Load().(KV)["http_proxy"]
}

func BinanceKey() string {
	return c.Load().(KV)["binance_key"]
}

func BinanceSecret() string {
	return c.Load().(KV)["binance_secret"]
}

func BinanceDiff() string {
	return c.Load().(KV)["binance_diff"]
}

func BarkUrls() []string {
	return strings.Split(c.Load().(KV)["bark_url"], ",")
}

const (
	AIProxyOne = "https://gpt.pawan.krd/backend-api/conversation"
	AIProxyTwo = "https://chatgpt.duti.tech/api/conversation"
)

var mappingKV = KV{
	// https://api.day.app/xxxxxx/标题/内容
	"bark_url":                "",
	"bili_cookie":             "",
	"azure_token":             "",
	"azure_model":             "",
	"azure_url":               "",
	"chatgpt_model":           openai.GPT3Dot5Turbo,
	"ai_mode":                 "api",
	"ai_browser_access_token": "",
	"ai_browser_proxy_url":    AIProxyOne,
	"ai_browser_model":        "text-davinci-002-render-sha",
	"user_id":                 "",
	"pixiv_mode":              "daily",
	"pixiv_session":           os.Getenv("PIXIV_SESSION"),
	"ai_token":                os.Getenv("AI_TOKEN"),
	"group_id":                os.Getenv("GROUP_ID"),
	"namespace":               os.Getenv("APP_NAMESPACE"),
	"pod_name":                os.Getenv("POD_NAME"),
	"weather_key":             os.Getenv("WEATHER_KEY"),
	"tian_api_key":            os.Getenv("TIAN_API_KEY"),
	"http_proxy":              os.Getenv("HTTP_PROXY"),
	"binance_key":             "",
	"binance_secret":          "",
	"binance_diff":            "100",
	"maotai":                  "",
}

type MTInfos map[string]MaoTaiInfo

func AddMaoTaiInfo(info MaoTaiInfo) {
	var infos MTInfos
	if err := json.Unmarshal([]byte(c.Load().(KV)["maotai"]), &infos); err != nil {
		infos = MTInfos{}
	}
	infos[info.Phone] = info
	marshal, _ := json.Marshal(&infos)
	Set(map[string]string{"maotai": string(marshal)})
}

func DelMaoTaiInfo(phone string) {
	var infos MTInfos
	if err := json.Unmarshal([]byte(c.Load().(KV)["maotai"]), &infos); err != nil {
		infos = MTInfos{}
	}
	delete(infos, phone)
	marshal, _ := json.Marshal(&infos)
	Set(map[string]string{"maotai": string(marshal)})
}

type MaoTaiInfo struct {
	Phone    string    `json:"phone"`
	Uid      int       `json:"uid"`
	Token    string    `json:"token"`
	Cookie   string    `json:"cookie"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
	ExpireAt time.Time `json:"expire_at"`
}

func (m *MaoTaiInfo) Expired() bool {
	return m.ExpireAt.Before(time.Now())
}

func MaoTaiInfoMap() map[string]MaoTaiInfo {
	var infos MTInfos
	err := json.Unmarshal([]byte(c.Load().(KV)["maotai"]), &infos)
	if err != nil {
		infos = MTInfos{}
	}

	return infos
}

func Set(m map[string]string) {
	var newKv = KV{}
	for k, v := range c.Load().(KV) {
		newv := v
		if s, ok := m[k]; ok && !(k == "pod_name" || k == "namespace") {
			newv = s
		}
		newKv[k] = newv
	}
	c.Store(newKv)
	if Pod() != "" {
		marshal, _ := json.Marshal(newKv)
		os.WriteFile(configFile, marshal, 0644)
	}
}
