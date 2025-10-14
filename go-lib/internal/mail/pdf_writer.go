// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Export Tool.  If not, see <https://www.gnu.org/licenses/>.

package mail

import (
	"bytes"
	"fmt"
	"html"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ProtonMail/export-tool/internal/utils"
	"github.com/ProtonMail/go-proton-api"
	"github.com/jung-kurt/gofpdf"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// PDFMessageWriter implements MessageWriter interface for PDF export
type PDFMessageWriter struct {
	msg proton.FullMessage
	eml bytes.Buffer
}

func NewPDFMessageWriter(msg proton.FullMessage, eml bytes.Buffer) *PDFMessageWriter {
	return &PDFMessageWriter{
		msg: msg,
		eml: eml,
	}
}

func (p *PDFMessageWriter) WriteMessage(dir string, tempDir string, log *logrus.Entry, integrityChecker utils.IntegrityChecker) error {
	filePath := filepath.Join(dir, p.msg.ID+pdfExtension)

	pdfBytes, err := p.generatePDF()
	if err != nil {
		log.WithField("msg-id", p.msg.ID).WithError(err).Error("Failed to generate PDF")
		return fmt.Errorf("failed to generate PDF for message '%v': %w", p.msg.ID, err)
	}

	if err := utils.WriteFileSafe(tempDir, filePath, pdfBytes, integrityChecker); err != nil {
		log.WithField("msg-id", p.msg.ID).WithError(err).Errorf("Failed to write file %v", filePath)
		return fmt.Errorf("failed to write PDF '%v': %w", filePath, err)
	}

	return nil
}

func (p *PDFMessageWriter) GetMetadata() MessageMetadata {
	return NewMessageMetadata(MessageWriterTypePDF, &p.msg.Message)
}

func (p *PDFMessageWriter) generatePDF() ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Add title
	pdf.Cell(40, 10, "Email Export")
	pdf.Ln(12)

	// Add email headers
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, fmt.Sprintf("From: %s", p.msg.Sender.Address))
	pdf.Ln(8)

	if len(p.msg.ToList) > 0 {
		toAddresses := make([]string, len(p.msg.ToList))
		for i, addr := range p.msg.ToList {
			toAddresses[i] = addr.Address
		}
		pdf.Cell(40, 10, fmt.Sprintf("To: %s", strings.Join(toAddresses, ", ")))
		pdf.Ln(8)
	}

	if len(p.msg.CCList) > 0 {
		ccAddresses := make([]string, len(p.msg.CCList))
		for i, addr := range p.msg.CCList {
			ccAddresses[i] = addr.Address
		}
		pdf.Cell(40, 10, fmt.Sprintf("CC: %s", strings.Join(ccAddresses, ", ")))
		pdf.Ln(8)
	}

	pdf.Cell(40, 10, fmt.Sprintf("Subject: %s", p.msg.Subject))
	pdf.Ln(8)

	pdf.Cell(40, 10, fmt.Sprintf("Date: %s", p.msg.Time.Format("2006-01-02 15:04:05")))
	pdf.Ln(12)

	// Add separator
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(8)

	// Add body content
	pdf.SetFont("Arial", "", 10)
	bodyText := p.extractTextFromEML()
	p.addTextToPDF(pdf, bodyText)

	// Add attachments info
	if len(p.msg.Attachments) > 0 {
		pdf.Ln(8)
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(40, 10, "Attachments:")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)

		for _, attachment := range p.msg.Attachments {
			pdf.Cell(40, 10, fmt.Sprintf("- %s (%d bytes)", attachment.Name, attachment.Size))
			pdf.Ln(6)
		}
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF output: %w", err)
	}

	return buf.Bytes(), nil
}

func (p *PDFMessageWriter) extractTextFromEML() string {
	emlContent := p.eml.String()
	
	// Extract body content after headers
	lines := strings.Split(emlContent, "\n")
	bodyStarted := false
	var bodyLines []string

	for _, line := range lines {
		if bodyStarted {
			bodyLines = append(bodyLines, line)
		} else if strings.TrimSpace(line) == "" {
			bodyStarted = true
		}
	}

	bodyText := strings.Join(bodyLines, "\n")
	
	// Check if content is HTML and extract text
	if strings.Contains(strings.ToLower(bodyText), "<html") || strings.Contains(strings.ToLower(bodyText), "<body") {
		return p.extractTextFromHTML(bodyText)
	}

	return bodyText
}

func (p *PDFMessageWriter) extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return p.stripHTMLTags(htmlContent)
	}

	var textContent strings.Builder
	p.extractTextFromNode(doc, &textContent)
	return textContent.String()
}

func (p *PDFMessageWriter) extractTextFromNode(n *html.Node, textContent *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			textContent.WriteString(text)
			textContent.WriteString(" ")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "br", "p", "div":
				textContent.WriteString("\n")
			}
		}
		p.extractTextFromNode(c, textContent)
	}
}

func (p *PDFMessageWriter) stripHTMLTags(htmlContent string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, "")
	return html.UnescapeString(text)
}

func (p *PDFMessageWriter) addTextToPDF(pdf *gofpdf.Fpdf, text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if len(line) > 80 {
			words := strings.Fields(line)
			currentLine := ""
			for _, word := range words {
				if len(currentLine)+len(word)+1 > 80 {
					if currentLine != "" {
						pdf.Cell(0, 6, currentLine)
						pdf.Ln(6)
						currentLine = word
					} else {
						pdf.Cell(0, 6, word)
						pdf.Ln(6)
					}
				} else {
					if currentLine != "" {
						currentLine += " " + word
					} else {
						currentLine = word
					}
				}
			}
			if currentLine != "" {
				pdf.Cell(0, 6, currentLine)
				pdf.Ln(6)
			}
		} else {
			pdf.Cell(0, 6, line)
			pdf.Ln(6)
		}
	}
}