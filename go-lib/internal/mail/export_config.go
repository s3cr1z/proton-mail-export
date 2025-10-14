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

// ExportFormat represents the available export formats
type ExportFormat int

const (
	ExportFormatEML ExportFormat = iota
	ExportFormatPDF
)

func (f ExportFormat) String() string {
	switch f {
	case ExportFormatEML:
		return "EML"
	case ExportFormatPDF:
		return "PDF"
	default:
		return "Unknown"
	}
}

// ExportConfig holds configuration for email export
type ExportConfig struct {
	Format ExportFormat
	// Future: Add PDF-specific options like page size, font, etc.
}

// NewDefaultExportConfig returns a default export configuration
func NewDefaultExportConfig() *ExportConfig {
	return &ExportConfig{
		Format: ExportFormatEML, // Default to EML for backward compatibility
	}
}

// NewPDFExportConfig returns a configuration for PDF export
func NewPDFExportConfig() *ExportConfig {
	return &ExportConfig{
		Format: ExportFormatPDF,
	}
}