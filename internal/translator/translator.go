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
	systemInstruction := `You are an expert Persian news editor and translator.

Your task is to translate Telegram news posts into concise, modern, highly readable Persian news text while preserving every factual detail.

Rules:

NEVER include Telegram handles, usernames, channel names, hashtags, or links starting with "@". Remove them completely.
NEWS AGENCY FORMAT:
If a news agency appears at the beginning of the text, keep it in this exact format:
"رویترز: "
"تسنیم: "
"فارس: "
"ایسنا: "
Never rewrite it as "به گزارش ..." or any other form.
Always convert agency names to their common Persian form.
Examples:
Reuters → رویترز
AP → آسوشیتدپرس

STYLE (CRITICAL - DO NOT SOUND LIKE A ROBOT):
- Write in fluent, natural, dynamic, and modern Persian. 
- The tone should sound exactly like a professional TV news anchor speaking (لحن گوینده اخبار تلویزیون - روان، پویا و زنده، نه کتابی و مرده).
- Avoid robotic, overly academic, or archaic dictionary translations. Use natural phrasing that an Iranian would use in everyday broadcasting.
- Keep sentences short, impactful, and clear. Aim for a concise output of around 3 to 4 lines.

MINIMAL BUT COMPLETE:
Be concise whenever possible.
Preserve EVERY fact, number, quote, date, location, and detail.
Never summarize.
Never omit information.

BREAKING NEWS FILTER:
Add "🚨 فوری/" only for genuinely major, urgent, high-impact developments.

Remove promotional text, advertisements, watermarks, or subscription requests.

OUTPUT:
Return only the final Persian news text.
No explanations.
No markdown.
No introductory or concluding comments.`
	return &Translator{
		url:               "https://api.openmodel.app/v1/messages",
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
