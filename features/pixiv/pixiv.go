package pixiv

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/util/proxy"
	"qq/util/retry"
	"strings"
	"time"

	"github.com/NateScarlet/pixiv/pkg/artwork"
	log "github.com/sirupsen/logrus"

	"github.com/NateScarlet/pixiv/pkg/client"
)

var (
	httpClient = proxy.NewHttpProxyClient
)

func newClientCtx() (context.Context, error) {
	var s string = config.PixivSession()
	// 使用 PHPSESSID Cookie 登录 (推荐)。
	c := &client.Client{
		Client: *httpClient(),
	}
	c.SetDefaultHeader("User-Agent", client.DefaultUserAgent)
	c.SetPHPSESSID(s)

	// 所有查询从 context 获取客户端设置, 如未设置将使用默认客户端。
	var ctx = context.TODO()
	ctx = client.With(ctx, c)
	return ctx, nil
}

func init() {
	features.AddKeyword("pd", "pixiv_mode 设置成 daily", func(bot bot.Bot, content string) error {
		config.Set(map[string]string{"pixiv_mode": "daily"})
		bot.Send("pixiv_mode 已设置成 daily")
		return nil
	}, features.WithHidden(), features.WithGroup("pixiv"))
	features.AddKeyword("pw", "pixiv_mode 设置成 weekly", func(bot bot.Bot, content string) error {
		config.Set(map[string]string{"pixiv_mode": "weekly"})
		bot.Send("pixiv_mode 已设置成 weekly")
		return nil
	}, features.WithHidden())
	features.AddKeyword("pm", "pixiv_mode 设置成 monthly", func(bot bot.Bot, content string) error {
		config.Set(map[string]string{"pixiv_mode": "monthly"})
		bot.Send("pixiv_mode 已设置成 monthly")
		return nil
	}, features.WithHidden(), features.WithGroup("pixiv"))
	features.AddKeyword("p", "<+n/r/rai> pixiv 热榜图片", func(bot bot.Bot, content string) error {
		image, err := Image(content)
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
		var msgID string
		if bot.Message().WeSendImg != nil {
			open, _ := os.Open(image)
			defer open.Close()
			img, _ := bot.Message().WeSendImg(open)
			msgID = img.MsgId
		} else {
			msgID = bot.Send(fmt.Sprintf("[CQ:image,file=file://%s]", image))
		}
		os.Remove(image)
		if bot.IsGroupMessage() {
			tID := bot.Send("图片即将在 30s 之后撤回，要保存的赶紧了~")
			time.Sleep(30 * time.Second)
			bot.DeleteMsg(msgID)
			bot.DeleteMsg(tID)
		}
		return nil
	}, features.WithGroup("pixiv"))

}
func Image(content string) (string, error) {
	//daily_r18_ai
	//daily_r18
	//daily
	//daily_ai
	//monthly
	isDaily := func() bool {
		return strings.Contains(config.PixivMode(), "daily")
	}
	var mode = config.PixivMode()
	switch content {
	case "n":
	case "r":
		mode = mode + "_r18"
	case "rai":
		if isDaily() {
			mode = mode + "_r18_ai"
			break
		}
		mode = mode + "_r18"
	default:
		if isDaily() {
			mode = mode + "_ai"
		}
	}
	ctx, err := newClientCtx()
	if err != nil {
		log.Println(err)
		return "", err
	}
	rank := &artwork.Rank{Mode: mode}
	err = retry.Times(20, func() error {
		rank.Page = 1
		if !isDaily() {
			rank.Page = rand.Intn(5) + 1
		}
		return rank.Fetch(ctx)
	})
	if err != nil {
		log.Println(err)
		return "", err
	}
	image := rank.Items[rand.Intn(len(rank.Items))]
	a := artwork.Artwork{
		ID: image.ID,
	}
	err = retry.Times(20, func() error {
		return a.Fetch(ctx)
	})
	if err != nil {
		log.Println(err)
		return "", err
	}
	var get *http.Response
	c := httpClient()
	err = retry.Times(20, func() error {
		var err error
		request, _ := http.NewRequest("GET", a.Image.Original, nil)
		request.Header.Add("Referer", "https://www.pixiv.net/")
		get, err = c.Do(request)
		return err
	})
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer get.Body.Close()
	base := filepath.Base(a.Image.Original)
	fpath := filepath.Join("/data", "images", base)

	os.MkdirAll(filepath.Join("/data", "images"), 0755)
	err = func() error {
		file, err := os.OpenFile(fpath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Println(err)
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, get.Body)
		return err
	}()
	if err != nil {
		log.Println(err)
		return "", err
	}

	return fpath, err
}
