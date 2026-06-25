package scraper

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Scraper struct {
	client    *http.Client
	urlRegex  *regexp.Regexp
	textRegex *regexp.Regexp
}

func New() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},

		urlRegex:  regexp.MustCompile(`url\(['"](.+?)['"]\)`),
		textRegex: regexp.MustCompile(`\b[!-ÿ]+\b`),
	}
}

func (sc *Scraper) ScrapeChannel(ctx context.Context, url string) ([]MessageGroup, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	messageGroup := doc.Find(".tgme_widget_message_wrap")
	var posts = make([]MessageGroup, 0, messageGroup.Length())

	messageGroup.Each(func(i int, s *goquery.Selection) {
		//messageId
		var messageID string
		messageNode := s.Find(".js-widget_message")
		if dataPost, exist := messageNode.Attr("data-post"); exist {
			messageID = dataPost
		}
		//images
		photoNode := s.Find(".tgme_widget_message_photo_wrap")
		imageLinks := make([]string, 0, photoNode.Length())
		photoNode.Each(func(i int, s *goquery.Selection) {

			if style, exist := s.Attr("style"); exist {
				matches := sc.urlRegex.FindAllStringSubmatch(style, -1)

				for _, match := range matches {
					if len(match) > 0 {
						imageLinks = append(imageLinks, match[1])

					}
				}
			}

		})
		videoNode := s.Find(".tgme_widget_message_video.js-message_video")
		videoLinks := make([]string, 0, videoNode.Length())
		videoNode.Each(func(i int, s *goquery.Selection) {
			if src, exist := s.Attr("src"); exist {
				videoLinks = append(videoLinks, src)
			}
		})

		//text
		messageElement := s.Find(".tgme_widget_message_text.js-message_text")

		posts = append(posts, MessageGroup{
			ID:        messageID,
			Message:   messageElement.Text(),
			ImageUris: imageLinks,
			VideoUris: videoLinks,
		})
	})
	return posts, nil
}
