package exporters

import (
	"fmt"
	"html"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// PDFExporter handles exporting emails to PDF format
type PDFExporter struct {
	outputPath string
	options    PDFOptions
}

// PDFOptions configures PDF export behavior
type PDFOptions struct {
	PageSize        string  // "A4", "Letter", etc.
	Orientation     string  // "P" (Portrait) or "L" (Landscape)
	FontFamily      string  // "Arial", "Times", etc.
	FontSize        float64
	IncludeHeaders  bool
	IncludeAttachments bool
	OneEmailPerPage bool
	WatermarkText   string
	Margins         PDFMargins
}

type PDFMargins struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// Email represents an email for PDF export
type Email struct {
	Subject     string
	From        string
	To          []string
	CC          []string
	BCC         []string
	Date        time.Time
	Body        string
	BodyHTML    string
	Attachments []Attachment
	MessageID   string
	InReplyTo   string
	References  []string
}

type Attachment struct {
	Name     string
	Size     int64
	MimeType string
	Content  []byte
}

// NewPDFExporter creates a new PDF exporter
func NewPDFExporter(outputPath string, options PDFOptions) *PDFExporter {
	// Set defaults
	if options.PageSize == "" {
		options.PageSize = "A4"
	}
	if options.Orientation == "" {
		options.Orientation = "P"
	}
	if options.FontFamily == "" {
		options.FontFamily = "Arial"
	}
	if options.FontSize == 0 {
		options.FontSize = 10
	}
	if options.Margins.Top == 0 {
		options.Margins = PDFMargins{Top: 20, Right: 20, Bottom: 20, Left: 20}
	}
	
	return &PDFExporter{
		outputPath: outputPath,
		options:    options,
	}
}

// ExportEmails exports a slice of emails to PDF
func (e *PDFExporter) ExportEmails(emails []Email, filename string) error {
	pdf := fpdf.New(e.options.Orientation, "mm", e.options.PageSize, "")
	
	// Set margins
	pdf.SetMargins(e.options.Margins.Left, e.options.Margins.Top, e.options.Margins.Right)
	pdf.SetAutoPageBreak(true, e.options.Margins.Bottom)
	
	// Add font
	pdf.SetFont(e.options.FontFamily, "", e.options.FontSize)
	
	for i, email := range emails {
		if e.options.OneEmailPerPage && i > 0 {
			pdf.AddPage()
		} else if i > 0 {
			// Add separator between emails
			pdf.Ln(10)
			e.addSeparator(pdf)
			pdf.Ln(5)
		}
		
		if err := e.addEmailToPDF(pdf, email); err != nil {
			return fmt.Errorf("failed to add email to PDF: %w", err)
		}
	}
	
	// Add watermark if specified
	if e.options.WatermarkText != "" {
		e.addWatermark(pdf)
	}
	
	// Save PDF
	outputFile := filepath.Join(e.outputPath, filename)
	return pdf.OutputFileAndClose(outputFile)
}

// ExportSingleEmail exports a single email to PDF
func (e *PDFExporter) ExportSingleEmail(email Email, filename string) error {
	return e.ExportEmails([]Email{email}, filename)
}

func (e *PDFExporter) addEmailToPDF(pdf *fpdf.Fpdf, email Email) error {
	// Add page if this is the first email or one email per page is enabled
	if pdf.PageCount() == 0 {
		pdf.AddPage()
	}
	
	// Email header section
	if e.options.IncludeHeaders {
		e.addEmailHeaders(pdf, email)
		pdf.Ln(5)
	}
	
	// Email body
	e.addEmailBody(pdf, email)
	
	// Attachments section
	if e.options.IncludeAttachments && len(email.Attachments) > 0 {
		pdf.Ln(5)
		e.addAttachmentsSection(pdf, email.Attachments)
	}
	
	return nil
}

func (e *PDFExporter) addEmailHeaders(pdf *fpdf.Fpdf, email Email) {
	// Subject (bold, larger font)
	pdf.SetFont(e.options.FontFamily, "B", e.options.FontSize+2)
	pdf.Cell(0, 8, "Subject: "+email.Subject)
	pdf.Ln(8)
	
	// Reset font
	pdf.SetFont(e.options.FontFamily, "", e.options.FontSize)
	
	// From
	pdf.Cell(0, 6, "From: "+email.From)
	pdf.Ln(6)
	
	// To
	if len(email.To) > 0 {
		pdf.Cell(0, 6, "To: "+strings.Join(email.To, ", "))
		pdf.Ln(6)
	}
	
	// CC
	if len(email.CC) > 0 {
		pdf.Cell(0, 6, "CC: "+strings.Join(email.CC, ", "))
		pdf.Ln(6)
	}
	
	// Date
	pdf.Cell(0, 6, "Date: "+email.Date.Format("2006-01-02 15:04:05 MST"))
	pdf.Ln(6)
	
	// Message ID (if available)
	if email.MessageID != "" {
		pdf.SetFont(e.options.FontFamily, "", e.options.FontSize-1)
		pdf.Cell(0, 5, "Message-ID: "+email.MessageID)
		pdf.Ln(5)
		pdf.SetFont(e.options.FontFamily, "", e.options.FontSize)
	}
}

func (e *PDFExporter) addEmailBody(pdf *fpdf.Fpdf, email Email) {
	// Add separator line
	e.addSeparator(pdf)
	pdf.Ln(3)
	
	// Choose body content (prefer HTML if available, but convert to text)
	var bodyText string
	if email.BodyHTML != "" {
		bodyText = e.htmlToText(email.BodyHTML)
	} else {
		bodyText = email.Body
	}
	
	// Split text into lines and add to PDF
	lines := strings.Split(bodyText, "\n")
	for _, line := range lines {
		// Handle long lines by wrapping
		if len(line) > 100 {
			wrappedLines := e.wrapText(line, 100)
			for _, wrappedLine := range wrappedLines {
				pdf.Cell(0, 5, wrappedLine)
				pdf.Ln(5)
			}
		} else {
			pdf.Cell(0, 5, line)
			pdf.Ln(5)
		}
	}
}

func (e *PDFExporter) addAttachmentsSection(pdf *fpdf.Fpdf, attachments []Attachment) {
	// Attachments header
	pdf.SetFont(e.options.FontFamily, "B", e.options.FontSize)
	pdf.Cell(0, 6, "Attachments:")
	pdf.Ln(6)
	
	// Reset font
	pdf.SetFont(e.options.FontFamily, "", e.options.FontSize)
	
	// List attachments
	for _, attachment := range attachments {
		sizeStr := e.formatFileSize(attachment.Size)
		attachmentInfo := fmt.Sprintf("• %s (%s, %s)", attachment.Name, sizeStr, attachment.MimeType)
		pdf.Cell(0, 5, attachmentInfo)
		pdf.Ln(5)
	}
}

func (e *PDFExporter) addSeparator(pdf *fpdf.Fpdf) {
	// Draw a horizontal line
	x, y := pdf.GetXY()
	pdf.Line(x, y, x+180, y) // 180mm line width for A4
}

func (e *PDFExporter) addWatermark(pdf *fpdf.Fpdf) {
	// Add watermark to all pages
	pageCount := pdf.PageCount()
	for i := 1; i <= pageCount; i++ {
		pdf.SetPage(i)
		
		// Set watermark properties
		pdf.SetFont(e.options.FontFamily, "", 50)
		pdf.SetTextColor(200, 200, 200) // Light gray
		
		// Calculate position for center of page
		pageWidth, pageHeight := pdf.GetPageSize()
		textWidth := pdf.GetStringWidth(e.options.WatermarkText)
		x := (pageWidth - textWidth) / 2
		y := pageHeight / 2
		
		// Add watermark text
		pdf.SetXY(x, y)
		pdf.Cell(textWidth, 20, e.options.WatermarkText)
		
		// Reset text color
		pdf.SetTextColor(0, 0, 0)
	}
}

func (e *PDFExporter) htmlToText(htmlContent string) string {
	// Simple HTML to text conversion
	// In a real implementation, you might want to use a proper HTML parser
	text := htmlContent
	
	// Remove HTML tags
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")
	text = strings.ReplaceAll(text, "<p>", "\n")
	text = strings.ReplaceAll(text, "</p>", "\n")
	text = strings.ReplaceAll(text, "<div>", "\n")
	text = strings.ReplaceAll(text, "</div>", "\n")
	
	// Remove remaining HTML tags (simple approach)
	for strings.Contains(text, "<") && strings.Contains(text, ">") {
		start := strings.Index(text, "<")
		end := strings.Index(text[start:], ">")
		if end != -1 {
			text = text[:start] + text[start+end+1:]
		} else {
			break
		}
	}
	
	// Decode HTML entities
	text = html.UnescapeString(text)
	
	// Clean up extra whitespace
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

func (e *PDFExporter) wrapText(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}
	
	var lines []string
	words := strings.Fields(text)
	currentLine := ""
	
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= maxLength {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}
	
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	return lines
}

func (e *PDFExporter) formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PDFExportWriter implements io.Writer for streaming PDF export
type PDFExportWriter struct {
	exporter *PDFExporter
	emails   []Email
	filename string
}

func (e *PDFExporter) NewWriter(filename string) *PDFExportWriter {
	return &PDFExportWriter{
		exporter: e,
		filename: filename,
		emails:   make([]Email, 0),
	}
}

func (w *PDFExportWriter) Write(p []byte) (n int, err error) {
	// This would parse the input bytes as email data
	// For now, this is a placeholder implementation
	return len(p), nil
}

func (w *PDFExportWriter) AddEmail(email Email) {
	w.emails = append(w.emails, email)
}

func (w *PDFExportWriter) Close() error {
	return w.exporter.ExportEmails(w.emails, w.filename)
}