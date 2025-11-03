package summarizer

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// SummarizeArticle uses Google's Gemini AI to summarize an article
func SummarizeArticle(articleContent string, articleURL string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	prompt := fmt.Sprintf(`Please provide a concise summary of the following article from %s.
Focus on the main points and key takeaways. Format the summary in a clear, readable manner.

Article content:
%s`, articleURL, articleContent)

	contents := genai.Text(prompt)

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-lite", contents, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	summary := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		summary += part.Text
	}

	return summary, nil
}
