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
	"encoding/json"
	"time"
)

// ExportConfig contains configuration for email export including filtering options
type ExportConfig struct {
	// Date range filtering (Unix timestamps)
	DateStart *int64 `json:"date_start,omitempty"`
	DateEnd   *int64 `json:"date_end,omitempty"`

	// Address and domain filtering
	SenderAddresses []string `json:"sender_addresses,omitempty"`
	SenderDomains   []string `json:"sender_domains,omitempty"`
	SenderPatterns  []string `json:"sender_patterns,omitempty"`

	// Label and folder filtering
	LabelNames  []string `json:"label_names,omitempty"`
	FolderNames []string `json:"folder_names,omitempty"`

	// Content filtering
	HasAttachments *bool `json:"has_attachments,omitempty"`

	// Export format options
	Format string `json:"format,omitempty"` // "eml", "pdf", etc.

	// Encryption options
	Encrypt           bool   `json:"encrypt,omitempty"`
	EncryptionPassword string `json:"encryption_password,omitempty"`

	// Incremental backup
	Incremental bool `json:"incremental,omitempty"`
}

// NewExportConfig creates a new ExportConfig with default values
func NewExportConfig() *ExportConfig {
	return &ExportConfig{
		SenderAddresses: make([]string, 0),
		SenderDomains:   make([]string, 0),
		SenderPatterns:  make([]string, 0),
		LabelNames:      make([]string, 0),
		FolderNames:     make([]string, 0),
		Format:          "eml",
	}
}

// ToExportFilter converts the config to an ExportFilter
func (c *ExportConfig) ToExportFilter() (*ExportFilter, error) {
	filter := NewExportFilter()

	// Convert date range
	if c.DateStart != nil {
		t := time.Unix(*c.DateStart, 0)
		filter.DateStart = &t
	}
	if c.DateEnd != nil {
		t := time.Unix(*c.DateEnd, 0)
		filter.DateEnd = &t
	}

	// Copy address and domain filters
	filter.SenderAddresses = append(filter.SenderAddresses, c.SenderAddresses...)
	filter.SenderDomains = append(filter.SenderDomains, c.SenderDomains...)

	// Compile regex patterns
	for _, pattern := range c.SenderPatterns {
		if err := filter.AddSenderPattern(pattern); err != nil {
			return nil, err
		}
	}

	// Copy label and folder filters
	filter.LabelNames = append(filter.LabelNames, c.LabelNames...)
	filter.FolderNames = append(filter.FolderNames, c.FolderNames...)

	// Copy attachment filter
	if c.HasAttachments != nil {
		filter.SetHasAttachments(*c.HasAttachments)
	}

	return filter, nil
}

// FromJSON creates an ExportConfig from JSON string
func (c *ExportConfig) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), c)
}

// ToJSON converts the config to JSON string
func (c *ExportConfig) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsEmpty returns true if no filtering options are set
func (c *ExportConfig) IsEmpty() bool {
	return c.DateStart == nil &&
		c.DateEnd == nil &&
		len(c.SenderAddresses) == 0 &&
		len(c.SenderDomains) == 0 &&
		len(c.SenderPatterns) == 0 &&
		len(c.LabelNames) == 0 &&
		len(c.FolderNames) == 0 &&
		c.HasAttachments == nil
}