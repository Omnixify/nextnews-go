package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Translator struct {
	apiKey            string
	model             string
	url               string
	systemInstruction string
}

func New(apiKey string) *Translator {
	systemInstruction := `You are an expert news editor and translator fluent in both English and Persian.
		Your task is to translate and format Telegram news posts into a modern, highly readable, and semi-formal Persian journalistic style.

		Rules:
		1. NEVER include any Telegram handles, usernames, channel names, or links starting with '@'. Strip them completely.
		2. EXACT AGENCY FORMAT: If a news agency (like Tasnim, Reuters, etc.) is mentioned at the beginning, keep it EXACTLY in this format: "Agency Name: " (e.g., "تسنیم: "). DO NOT change it to phrases like "به گزارش تسنیم،".
		3. TONE & READABILITY (CRITICAL): Use simple, fluid, and natural Persian phrasing. Avoid overly complex, archaic, or heavy academic words (ادبیات سخت و سنگین). It should feel smooth and accessible (روان و خوش‌خوان), but still maintain a professional news standard (not slang/شکسته).
		4. NO DATA LOSS: You must translate and include EVERY SINGLE piece of information, detail, and fact from the source text. Do not summarize, skip, or omit any details while making it readable.
		5. INTELLIGENT BREAKING NEWS FILTER: Evaluate the importance of the news yourself.
		- DO NOT add "فوری" or "🚨" to regular political updates, meeting details, or standard reports.
		- ONLY use breaking news headers (e.g., "🚨 **فوری/** ...") if the event is a major, high-stakes, sudden, or critical development. Be very selective.
		6. Remove promotional footers, watermarks, or "Join our channel" calls-to-action.
		7. Fix any raw scraping errors, triple dots (...), or typos from the source text seamlessly.
		8. Output ONLY the clean translated Persian news text. Do not add intro or outro comments.`
	return &Translator{
		url:               "https://wispy-cell-b30e.hazem-omnixify.workers.dev/v1/messages",
		apiKey:            apiKey,
		model:             "deepseek-v4-flash",
		systemInstruction: systemInstruction,
	}
}

func (t *Translator) Translate(ctx context.Context, text string) (string, error) {

	payload := OpenModelRequest{
		Model: t.model,
		Message: []Message{
			{Role: "system", Content: t.systemInstruction},
			{Role: "user", Content: text},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errorResp OpenModelErrorResponse
		if err := json.Unmarshal(respByte, &errorResp); err != nil {
			fmt.Println("oiii")
			return "", err
		}
		return errorResp.Error.Message, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		read, _ := io.ReadAll(resp.Body)
		fmt.Println(read)
		return "", err
	}

	var apiResp OpenModelResponse
	if err := json.Unmarshal(respByte, &apiResp); err != nil {
		return "", err
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("received empty content response from API")
	}
	fmt.Printf("response : %s\n\n", apiResp.Content[1].Text)

	return apiResp.Content[1].Text, nil
}
