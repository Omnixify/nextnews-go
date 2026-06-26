package telegram

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/PooryaAlirezazadeh/TeleBot/internal/cache"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/scraper"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/translator"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Engine struct {
	scraper    *scraper.Scraper
	translator *translator.Translator
	cache      *cache.Client
	token      string
}

func NewEngine(s *scraper.Scraper, t *translator.Translator, token string, c *cache.Client) *Engine {
	return &Engine{scraper: s, translator: t, token: token, cache: c}
}

func (e *Engine) RunOnce(ctx context.Context) {
	b, err := bot.New(e.token)
	if err != nil {
		log.Fatalf("Bot init failed: %v", err)
	}

	posts, err := e.scraper.ScrapeChannel(ctx, "https://t.me/s/wfwitness")
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}

	latestID, _ := e.cache.GetLatestID(ctx)
	currentWatermark := extractIDnumber(latestID)

	for _, post := range posts {
		if extractIDnumber(post.ID) <= currentWatermark {
			continue
		}

		translatedText, _ := e.translator.Translate(ctx, post.Message)
		message := fmt.Sprintf("%s \n\n🌎 @xNextNews", translatedText)

		mediaGroup := []models.InputMedia{}

		for _, imgUrl := range post.ImageUris {
			mediaGroup = append(mediaGroup, &models.InputMediaPhoto{
				Media: fmt.Sprintf("https://wsrv.nl/?url=%s", url.QueryEscape(imgUrl)),
			})
		}

		for _, vidUrl := range post.VideoUris {
			resp, err := http.Get(vidUrl)
			if err == nil {
				data, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				mediaGroup = append(mediaGroup, &models.InputMediaVideo{
					Media:           "attach://video.mp4",
					MediaAttachment: bytes.NewReader(data),
				})
			}
		}

		if len(mediaGroup) > 0 {
			switch m := mediaGroup[0].(type) {
			case *models.InputMediaPhoto:
				m.Caption = message
				m.ParseMode = models.ParseModeHTML
			case *models.InputMediaVideo:
				m.Caption = message
				m.ParseMode = models.ParseModeHTML
			}

			b.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
				ChatID: "@xNextNews",
				Media:  mediaGroup,
			})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    "@xNextNews",
				Text:      message,
				ParseMode: models.ParseModeHTML,
			})
		}

		e.cache.SetLatestID(ctx, post.ID)
		time.Sleep(2 * time.Second)
	}
}
