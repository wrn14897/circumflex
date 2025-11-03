package reader

import (
	"fmt"
	"time"

	ansi "clx/utils/strip-ansi"

	"clx/reader/markdown/postprocessor"
	"clx/reader/markdown/terminal"
	"clx/reader/summarizer"

	"clx/reader/markdown/html"
	"clx/reader/markdown/parser"

	"github.com/go-shiori/go-readability"
)

func GetArticle(url string, title string, width int, indentationSymbol string) (string, error) {
	return getArticleInternal(url, title, width, indentationSymbol, false)
}

func GetArticleWithSummary(url string, title string, width int, indentationSymbol string) (string, error) {
	return getArticleInternal(url, title, width, indentationSymbol, true)
}

func getArticleInternal(url string, title string, width int, indentationSymbol string, summarize bool) (string, error) {
	articleInRawHtml, httpErr := readability.FromURL(url, 6*time.Second)
	if httpErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	articleContentInRawHtmlAndSanitized := ansi.Strip(articleInRawHtml.Content)

	articleInMarkdown, mdErr := html.ConvertToMarkdown(articleContentInRawHtmlAndSanitized)
	if mdErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	var articleInTerminalFormal string
	header := terminal.CreateHeader(title, url, width)

	// If summarization is requested, process summary separately
	if summarize {
		summary, summErr := summarizer.SummarizeArticle(articleInMarkdown, url)

		if summErr != nil {
			// If summarization fails, add error message but still show original article
			errorMarkdown := fmt.Sprintf("**Error generating summary:** %s\n\n---\n\n", summErr.Error())
			errorBlocks := parser.ConvertToMarkdownBlocks(errorMarkdown)
			errorFormatted := terminal.ConvertToTerminalFormat(errorBlocks, width, indentationSymbol)

			markdownBlocks := parser.ConvertToMarkdownBlocks(articleInMarkdown)
			articleFormatted := terminal.ConvertToTerminalFormat(markdownBlocks, width, indentationSymbol)

			articleInTerminalFormal = postprocessor.Process(header+errorFormatted+articleFormatted, url)
		} else {
			// Combine summary and article as markdown, then process together
			// The summary from Gemini is already in markdown format
			combinedMarkdown := fmt.Sprintf("# AI Summary\n\n%s\n\n---\n\n# Full Article\n\n%s", summary, articleInMarkdown)

			markdownBlocks := parser.ConvertToMarkdownBlocks(combinedMarkdown)
			articleInTerminalFormal = terminal.ConvertToTerminalFormat(markdownBlocks, width, indentationSymbol)
			articleInTerminalFormal = postprocessor.Process(header+articleInTerminalFormal, url)
		}
	} else {
		// Normal article without summary
		markdownBlocks := parser.ConvertToMarkdownBlocks(articleInMarkdown)
		articleInTerminalFormal = terminal.ConvertToTerminalFormat(markdownBlocks, width, indentationSymbol)
		articleInTerminalFormal = postprocessor.Process(header+articleInTerminalFormal, url)
	}

	return articleInTerminalFormal, nil
}
