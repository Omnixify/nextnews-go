package translator

import (
	"context"

	"google.golang.org/genai"
)

type Translator struct {
	genClient *genai.Client
	model     string
}

func New(apiKey string) (*Translator, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	return &Translator{
		genClient: client,
		model:     "gemini-3.1-flash-lite",
	}, nil
}

func (t *Translator) Translate(ctx context.Context, text string) (string, error) {
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
	config := genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction, genai.RoleUser),
	}

	resp, err := t.genClient.Models.GenerateContent(
		ctx,
		t.model,
		genai.Text(text),
		&config,
	)

	if err != nil {
		return "", err
	}
	return resp.Text(), nil
}
