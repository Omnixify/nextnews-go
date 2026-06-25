package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
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
	bot        *bot.Bot
}

func NewEngine(
	scraper *scraper.Scraper,
	translator *translator.Translator,
	token string,
	cache *cache.Client,
) *Engine {
	return &Engine{
		scraper:    scraper,
		translator: translator,
		token:      token,
		cache:      cache,
	}
}

func (e *Engine) Start(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   45 * time.Second,
		Transport: transport,
	}
	opts := []bot.Option{
		bot.WithDefaultHandler(e.handler),
		bot.WithHTTPClient(30*time.Second, httpClient),
		bot.WithServerURL("https://wispy-cell-b30e.hazem-omnixify.workers.dev"),
	}

	botClient, err := bot.New(e.token, opts...)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}
	e.bot = botClient

	go e.startPeriodicScraper(ctx)
	// e.translator.Translate(ctx, "hello")
	fmt.Println("Telegram bot is running...")
	botClient.Start(ctx)
}

func (e *Engine) startPeriodicScraper(ctx context.Context) {
	sleepTime := rand.Intn(50) + 320
	ticker := time.NewTicker(time.Duration(sleepTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping background periodic scraper...")
			return
		case <-ticker.C:
			e.executeScraperAndTranslate(ctx)
		}
	}
}

func (e *Engine) executeScraperAndTranslate(ctx context.Context) {
	scrapeCtx, scrapeCancel := context.WithTimeout(ctx, 10*time.Second)
	defer scrapeCancel()

	posts, err := e.scraper.ScrapeChannel(scrapeCtx, "https://wispy-cell-b30e.hazem-omnixify.workers.dev/s/wfwitness")
	if err != nil {
		log.Printf("error to scrape page: %v\n", err)
		return
	}
	fmt.Printf("post : %s \n\n", posts)
	postJson, err := json.MarshalIndent(posts, "", " ")

	fmt.Printf("\n%s\n\n", postJson)

	latestID, err := e.cache.GetLatestID(ctx)
	if err != nil {
		log.Println("cache error fetching watermark:", err)
		return
	}
	currentWatermark := extractIDnumber(latestID)

	newPosts := make([]scraper.MessageGroup, 0, len(posts))
	for _, post := range posts {
		if extractIDnumber(post.ID) > currentWatermark {
			newPosts = append(newPosts, post)
		}
	}

	if len(newPosts) == 0 {
		fmt.Println("no new post")
		return
	}

	ChannelID := "@xNextNews"

	sleepTime := rand.Intn(10) + 5
	ticker := time.NewTicker(time.Duration(sleepTime) * time.Second)
	defer ticker.Stop()

	for _, post := range newPosts {
		<-ticker.C
		translateCtx, translateCancel := context.WithTimeout(ctx, 60*time.Second)
		translatedText, err := e.translator.Translate(translateCtx, post.Message)
		translateCancel()

		message := fmt.Sprintf("%s \n\n🌎%s", translatedText, ChannelID)
		if err != nil {
			log.Printf("failed to translate message %s (API issue): %v\n", post.ID, err)
			return
		}

		totalMedia := len(post.ImageUris) + len(post.VideoUris)

		if totalMedia == 0 {
			_, err = e.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    ChannelID,
				Text:      message,
				ParseMode: models.ParseModeHTML,
			})
		} else {
			var mediaGroup []models.InputMedia
			for i, imgUrl := range post.ImageUris {
				mediaPhoto := &models.InputMediaPhoto{Media: fmt.Sprintf("https://wsrv.nl/?url=%s", url.QueryEscape(imgUrl))}
				if i == 0 {
					mediaPhoto.Caption = message
					mediaPhoto.ParseMode = models.ParseModeHTML
				}
				mediaGroup = append(mediaGroup, mediaPhoto)
			}

			for _, vidUrl := range post.VideoUris {
				pipeRead, pipeWrite := io.Pipe()

				go func(url string, w *io.PipeWriter) {
					req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
					if err != nil {
						w.CloseWithError(err)
						return
					}
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
					req.Header.Set("Referer", "https://t.me/")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						w.CloseWithError(err)
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						w.CloseWithError(fmt.Errorf("cdn server returned status: %s", resp.Status))
						return
					}

					_, err = io.Copy(w, resp.Body)
					w.CloseWithError(err)
				}(vidUrl, pipeWrite)

				mediaVideo := &models.InputMediaVideo{
					Media:           "attach://video.mp4",
					MediaAttachment: pipeRead,
				}

				if len(mediaGroup) == 0 {
					mediaVideo.Caption = message
					mediaVideo.ParseMode = models.ParseModeHTML
				}

				mediaGroup = append(mediaGroup, mediaVideo)

			}

			_, err = e.bot.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
				ChatID: ChannelID,
				Media:  mediaGroup,
			})
		}

		if err != nil {
			log.Printf("failed to send post %s to telegram channel: %v\n", post.ID, err)
			return
		}

		if err := e.cache.SetLatestID(ctx, post.ID); err != nil {
			log.Printf("failed to advance checkpoint watermark ID to %s: %v\n", post.ID, err)
			return
		}
		log.Printf("Successfully broadcasted and saved watermark for post: %s\n", post.ID)
	}
}

func (e *Engine) handler(ctx context.Context, b *bot.Bot, model *models.Update) {}
