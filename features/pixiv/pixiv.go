package pixiv

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/config"
	"qq/features"
	"sync"
	"time"

	"github.com/NateScarlet/pixiv/pkg/artwork"
	"github.com/NateScarlet/pixiv/pkg/client"
)

var (
	session = config.PixivSession()
	mu      sync.RWMutex

	httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
)

func newClientCtx() (context.Context, error) {
	var s string
	func() {
		mu.RLock()
		defer mu.RUnlock()
		s = session
	}()
	if s == "" {
		return nil, errors.New("请先设置session: pixiv-session +<session>")
	}
	// 使用 PHPSESSID Cookie 登录 (推荐)。
	c := &client.Client{
		Client: *httpClient,
	}
	c.SetDefaultHeader("User-Agent", client.DefaultUserAgent)
	c.SetPHPSESSID(s)

	// 所有查询从 context 获取客户端设置, 如未设置将使用默认客户端。
	var ctx = context.TODO()
	ctx = client.With(ctx, c)
	return ctx, nil
}

func init() {
	log.Println(config.PixivSession())
	rand.Seed(time.Now().UnixNano())
	features.AddKeyword("pixiv-session", "设置 pixiv session", func(bot bot.Bot, content string) error {
		mu.Lock()
		defer mu.Unlock()
		session = content
		bot.Send("已设置 pixiv session")
		return nil
	})
	features.AddKeyword("pixiv", "+<ai/r18/r18_ai> pixiv top", func(bot bot.Bot, content string) error {
		//daily_r18_ai
		//daily_r18
		//daily
		//daily_ai
		var mode = "daily"
		switch content {
		case "ai":
			mode = "daily_ai"
		case "r18":
			mode = "daily_r18"
		case "r18_ai":
			mode = "daily_r18_ai"
		}
		ctx, err := newClientCtx()
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
		rank := &artwork.Rank{Mode: mode, Page: rand.Intn(5)}
		if err = rank.Fetch(ctx); err != nil {
			bot.Send(err.Error())
			return nil
		}
		request, _ := http.NewRequest("GET", rank.Items[rand.Intn(len(rank.Items))].Image.Regular, nil)
		request.Header.Add("Referer", "https://www.pixiv.net/")
		get, err := httpClient.Do(request)
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
		defer get.Body.Close()
		url := rank.Items[rand.Intn(len(rank.Items))].Image.Regular
		base := Md5(filepath.Base(url)) + filepath.Ext(filepath.Base(url))
		os.MkdirAll("/data/images", 0755)
		fpath := filepath.Join("/data", "images", base)
		all, _ := io.ReadAll(get.Body)
		os.WriteFile(fpath, all, 0644)
		bot.Send(fmt.Sprintf("[CQ:image,type=image,file=%s]", base))
		bot.Send(fmt.Sprintf("[CQ:image,type=image,file=%s]", filepath.Join("images", base)))
		return nil
	})
}

func Md5(data string) string {
	hash := md5.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}
