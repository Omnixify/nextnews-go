package translator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

type Translator struct {
	browserCtx context.Context
	cancelFunc context.CancelFunc
	url        string
}

func New() (*Translator, error) {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", false),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	baseCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithDebugf(log.Printf))

	err := chromedp.Run(baseCtx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start headless chrome: %w", err)
	}

	return &Translator{
		browserCtx: baseCtx,
		cancelFunc: cancel,
		url:        "https://duckduckgo.com/?q=duck+duck+go+translate&ia=web",
	}, nil
}

func (t *Translator) Translate(ctx context.Context, text string) (string, error) {
	reqCtx, cancel := context.WithTimeout(t.browserCtx, 15*time.Second)
	defer cancel()

	// var translateResult string

	err := chromedp.Run(reqCtx,
		chromedp.Navigate(t.url),
		chromedp.WaitVisible(".module--translations-section .module--translations-original .js-module--translations-original", chromedp.ByQuery),
		chromedp.Click(".module--translations-section .module--translations-original .js-module--translations-original", chromedp.ByQuery),
	)

	if err != nil {
		return "", err
	}
	return "", nil
}
