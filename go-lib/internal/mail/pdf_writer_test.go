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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProtonMail/export-tool/internal/utils"
	"github.com/ProtonMail/go-proton-api"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPDFMessageWriter_WriteMessage(t *testing.T) {
	// Create a test message
	msg := proton.FullMessage{
		Message: proton.Message{
			ID:      "test-message-id",
			Subject: "Test Email Subject",
			Sender: proton.MessageAddress{
				Address: "sender@example.com",
			},
			ToList: []proton.MessageAddress{
				{Address: "recipient@example.com"},
			},
			Time: time.Now(),
		},
	}

	// Create EML content
	emlContent := `From: sender@example.com
To: recipient@example.com
Subject: Test Email Subject
Date: ` + time.Now().Format(time.RFC1123Z) + `

This is a test email body.
It has multiple lines.
And should be converted to PDF properly.`

	var emlBuffer bytes.Buffer
	emlBuffer.WriteString(emlContent)

	// Create PDF writer
	pdfWriter := NewPDFMessageWriter(msg, emlBuffer)

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pdf-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outputDir := filepath.Join(tempDir, "output")
	err = os.MkdirAll(outputDir, 0o755)
	require.NoError(t, err)

	// Create logger
	logger := logrus.NewEntry(logrus.New())

	// Create integrity checker
	integrityChecker := &utils.Sha256IntegrityChecker{}

	// Write the PDF
	err = pdfWriter.WriteMessage(outputDir, tempDir, logger, integrityChecker)
	require.NoError(t, err)

	// Check that PDF file was created
	pdfPath := filepath.Join(outputDir, msg.ID+pdfExtension)
	assert.FileExists(t, pdfPath)

	// Check that the file is not empty
	stat, err := os.Stat(pdfPath)
	require.NoError(t, err)
	assert.Greater(t, stat.Size(), int64(0))

	// Verify metadata
	metadata := pdfWriter.GetMetadata()
	assert.Equal(t, MessageWriterTypePDF, metadata.WriterType)
	assert.Equal(t, msg.ID, metadata.ID)
}

func TestPDFMessageWriter_ExtractTextFromHTML(t *testing.T) {
	msg := proton.FullMessage{
		Message: proton.Message{
			ID: "test-html-message",
		},
	}

	htmlContent := `<html>
<body>
<h1>Test Header</h1>
<p>This is a paragraph with <strong>bold text</strong>.</p>
<br>
<div>This is a div element.</div>
</body>
</html>`

	var emlBuffer bytes.Buffer
	emlBuffer.WriteString("From: test@example.com\nSubject: HTML Test\n\n" + htmlContent)

	pdfWriter := NewPDFMessageWriter(msg, emlBuffer)
	extractedText := pdfWriter.extractTextFromHTML(htmlContent)

	assert.Contains(t, extractedText, "Test Header")
	assert.Contains(t, extractedText, "This is a paragraph")
	assert.Contains(t, extractedText, "bold text")
	assert.Contains(t, extractedText, "This is a div element")
	assert.NotContains(t, extractedText, "<html>")
	assert.NotContains(t, extractedText, "<body>")
}

func TestPDFMessageWriter_StripHTMLTags(t *testing.T) {
	msg := proton.FullMessage{}
	var emlBuffer bytes.Buffer
	pdfWriter := NewPDFMessageWriter(msg, emlBuffer)

	htmlContent := `<p>Hello <strong>world</strong>!</p><br><div>Test</div>`
	strippedText := pdfWriter.stripHTMLTags(htmlContent)

	assert.Equal(t, "Hello world!Test", strippedText)
	assert.NotContains(t, strippedText, "<")
	assert.NotContains(t, strippedText, ">")
}

func TestExportConfig(t *testing.T) {
	// Test default config
	defaultConfig := NewDefaultExportConfig()
	assert.Equal(t, ExportFormatEML, defaultConfig.Format)

	// Test PDF config
	pdfConfig := NewPDFExportConfig()
	assert.Equal(t, ExportFormatPDF, pdfConfig.Format)

	// Test format string representation
	assert.Equal(t, "EML", ExportFormatEML.String())
	assert.Equal(t, "PDF", ExportFormatPDF.String())
}