package binder

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/user/lector/internal/models"
)

func BindPDF(doc *models.Document, path string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(doc.Title, true)
	pdf.SetAuthor(doc.Author, true)

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 24)
	pdf.MultiCell(0, 20, doc.Title, "", "C", false)
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 16)
	pdf.MultiCell(0, 10, fmt.Sprintf("By %s", doc.Author), "", "C", false)
	pdf.Ln(20)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 7, stripHTML(doc.Synopsis), "", "L", false)

	for _, ch := range doc.Chapters {
		if ch.Content == "" {
			continue
		}
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 18)
		pdf.MultiCell(0, 10, ch.Title, "", "L", false)
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 11)

		content := stripHTML(ch.Content)
		pdf.MultiCell(0, 6, content, "", "L", false)
	}

	return pdf.OutputFileAndClose(path)
}

func stripHTML(html string) string {
	r := strings.NewReplacer("<p>", "\n", "</p>", "\n", "<br>", "\n", "<br/>", "\n", "&nbsp;", " ")
	text := r.Replace(html)

	re := regexp.MustCompile("<[^>]*>")
	text = re.ReplaceAllString(text, "")

	re2 := regexp.MustCompile("\n\n+")
	text = re2.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}
