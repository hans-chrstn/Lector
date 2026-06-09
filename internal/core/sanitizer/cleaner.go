package sanitizer

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

func CleanHTML(html string, chapterTitle string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	doc.Find("script, style, iframe, ads, ins, .ads, #ads, [id*='google_ads'], [class*='ads-'], [class*='sponsored']").Each(func(i int, sel *goquery.Selection) {
		sel.Remove()
	})

	if chapterTitle != "" {
		doc.Find("h1, h2, h3, h4, h5, div, p").Each(func(i int, sel *goquery.Selection) {
			text := strings.TrimSpace(sel.Text())
			if strings.EqualFold(text, chapterTitle) ||
				(strings.HasPrefix(strings.ToLower(text), "chapter") && strings.Contains(strings.ToLower(chapterTitle), strings.ToLower(text))) {
				if i < 5 {
					sel.Remove()
				}
			}
		})
	}

	doc.Find("a, div, p, span").Each(func(i int, sel *goquery.Selection) {
		text := strings.ToLower(sel.Text())
		href, _ := sel.Attr("href")

		if href != "" && (strings.Contains(href, "freegames.click") ||
			strings.Contains(href, "sponsored") ||
			strings.Contains(href, "/ads/")) {
			sel.Remove()
			return
		}

		if len(text) < 300 && (strings.Contains(text, "report chapter") ||
			strings.Contains(text, "standard content") ||
			strings.Contains(text, "find any errors") ||
			strings.Contains(text, "broken links") ||
			strings.Contains(text, "let us know") ||
			strings.Contains(text, "please let us know")) {
			sel.Remove()
		}
	})

	styleRe := regexp.MustCompile(`(width|height):\s*(300|250)px`)
	doc.Find("div").Each(func(i int, sel *goquery.Selection) {
		style, _ := sel.Attr("style")
		if styleRe.MatchString(style) {
			sel.Remove()
			return
		}
		if strings.TrimSpace(sel.Text()) == "" && sel.Children().Length() == 0 {
			class, _ := sel.Attr("class")
			if strings.Contains(class, "ads") {
				sel.Remove()
			}
		}
	})

	h, _ := doc.Find("body").Html()
	if h == "" {
		h, _ = doc.Html()
	}

	p := bluemonday.UGCPolicy()
	return p.Sanitize(h)
}
