package reader

import (
	"fmt"
	"strings"
	"time"

	ansi "clx/utils/strip-ansi"

	"clx/constants/unicode"
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
			// Process summary (visible by default)
			summaryMarkdown := fmt.Sprintf("# AI Summary\n\n%s\n\n", summary)
			summaryBlocks := parser.ConvertToMarkdownBlocks(summaryMarkdown)
			summaryFormatted := terminal.ConvertToTerminalFormat(summaryBlocks, width, indentationSymbol)

			// Add toggle button
			toggleButton := createToggleButton("Show Full Article", width)

			// Process full article (hidden by default)
			articleMarkdown := fmt.Sprintf("# Full Article\n\n%s", articleInMarkdown)
			articleBlocks := parser.ConvertToMarkdownBlocks(articleMarkdown)
			articleFormatted := terminal.ConvertToTerminalFormat(articleBlocks, width, indentationSymbol)

			// Add collapse markers to hide the full article by default
			articleWithCollapseMarkers := addCollapseMarkers(articleFormatted)

			articleInTerminalFormal = postprocessor.Process(header+summaryFormatted+toggleButton+articleWithCollapseMarkers, url)
		}
	} else {
		// Normal article without summary
		markdownBlocks := parser.ConvertToMarkdownBlocks(articleInMarkdown)
		articleInTerminalFormal = terminal.ConvertToTerminalFormat(markdownBlocks, width, indentationSymbol)
		articleInTerminalFormal = postprocessor.Process(header+articleInTerminalFormal, url)
	}

	return articleInTerminalFormal, nil
}

// createToggleButton creates a button to toggle between collapsed and expanded views
func createToggleButton(label string, width int) string {
	buttonCollapsed := fmt.Sprintf("▶ %s", label)
	buttonExpanded := fmt.Sprintf("▼ %s", label)

	// Center the buttons
	collapsedLine := fmt.Sprintf("%*s", (width+len(buttonCollapsed))/2, buttonCollapsed)
	expandedLine := fmt.Sprintf("%*s", (width+len(buttonExpanded))/2, buttonExpanded)

	// Button for collapsed state (shown by default) + invisible marker for collapsed content
	// Button for expanded state (hidden by default) + invisible marker for expanded content
	return fmt.Sprintf("\n%s%s\n%s%s\n",
		collapsedLine, unicode.InvisibleCharacterForCollapse,
		expandedLine, unicode.InvisibleCharacterForExpansion)
}

// addCollapseMarkers adds invisible Unicode characters to each line to mark it as collapsible
func addCollapseMarkers(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i == len(lines)-1 {
			result.WriteString(line)
		} else {
			result.WriteString(line)
			result.WriteString(unicode.InvisibleCharacterForExpansion)
			result.WriteString("\n")
		}
	}

	return result.String()
}
