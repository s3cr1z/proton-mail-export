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
	"regexp"
	"strings"
	"time"

	"github.com/ProtonMail/go-proton-api"
)

// ExportFilter contains all filtering criteria for selective email export
type ExportFilter struct {
	// Date range filtering
	DateStart *time.Time
	DateEnd   *time.Time

	// Address and domain filtering
	SenderAddresses []string
	SenderDomains   []string
	SenderPatterns  []*regexp.Regexp

	// Label and folder filtering
	LabelIDs    []string
	LabelNames  []string
	FolderIDs   []string
	FolderNames []string

	// Content filtering
	HasAttachments *bool

	// Advanced filtering
	SubjectPatterns []*regexp.Regexp
	BodyPatterns    []*regexp.Regexp
}

// NewExportFilter creates a new ExportFilter with default values
func NewExportFilter() *ExportFilter {
	return &ExportFilter{
		SenderAddresses: make([]string, 0),
		SenderDomains:   make([]string, 0),
		SenderPatterns:  make([]*regexp.Regexp, 0),
		LabelIDs:        make([]string, 0),
		LabelNames:      make([]string, 0),
		FolderIDs:       make([]string, 0),
		FolderNames:     make([]string, 0),
		SubjectPatterns: make([]*regexp.Regexp, 0),
		BodyPatterns:    make([]*regexp.Regexp, 0),
	}
}

// AddSenderAddress adds a sender address to filter by
func (f *ExportFilter) AddSenderAddress(address string) {
	f.SenderAddresses = append(f.SenderAddresses, strings.ToLower(address))
}

// AddSenderDomain adds a sender domain to filter by
func (f *ExportFilter) AddSenderDomain(domain string) {
	f.SenderDomains = append(f.SenderDomains, strings.ToLower(domain))
}

// AddSenderPattern adds a regex pattern for sender filtering
func (f *ExportFilter) AddSenderPattern(pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	f.SenderPatterns = append(f.SenderPatterns, regex)
	return nil
}

// AddLabelName adds a label name to filter by
func (f *ExportFilter) AddLabelName(name string) {
	f.LabelNames = append(f.LabelNames, name)
}

// AddFolderName adds a folder name to filter by
func (f *ExportFilter) AddFolderName(name string) {
	f.FolderNames = append(f.FolderNames, name)
}

// SetDateRange sets the date range for filtering
func (f *ExportFilter) SetDateRange(start, end *time.Time) {
	f.DateStart = start
	f.DateEnd = end
}

// SetHasAttachments sets the attachment filter
func (f *ExportFilter) SetHasAttachments(hasAttachments bool) {
	f.HasAttachments = &hasAttachments
}

// IsEmpty returns true if no filters are set
func (f *ExportFilter) IsEmpty() bool {
	return f.DateStart == nil &&
		f.DateEnd == nil &&
		len(f.SenderAddresses) == 0 &&
		len(f.SenderDomains) == 0 &&
		len(f.SenderPatterns) == 0 &&
		len(f.LabelIDs) == 0 &&
		len(f.LabelNames) == 0 &&
		len(f.FolderIDs) == 0 &&
		len(f.FolderNames) == 0 &&
		f.HasAttachments == nil &&
		len(f.SubjectPatterns) == 0 &&
		len(f.BodyPatterns) == 0
}

// MatchesMessage checks if a message matches the filter criteria
func (f *ExportFilter) MatchesMessage(msg proton.MessageMetadata) bool {
	// If no filters are set, match all messages
	if f.IsEmpty() {
		return true
	}

	// Date range filtering
	if f.DateStart != nil && msg.Time < f.DateStart.Unix() {
		return false
	}
	if f.DateEnd != nil && msg.Time > f.DateEnd.Unix() {
		return false
	}

	// Sender filtering
	if len(f.SenderAddresses) > 0 || len(f.SenderDomains) > 0 || len(f.SenderPatterns) > 0 {
		if !f.matchesSender(msg.Sender) {
			return false
		}
	}

	// Label filtering
	if len(f.LabelIDs) > 0 || len(f.LabelNames) > 0 {
		if !f.matchesLabels(msg.LabelIDs) {
			return false
		}
	}

	// Attachment filtering
	if f.HasAttachments != nil {
		hasAttachments := len(msg.Attachments) > 0
		if *f.HasAttachments != hasAttachments {
			return false
		}
	}

	return true
}

// matchesSender checks if the sender matches any of the sender filters
func (f *ExportFilter) matchesSender(sender *proton.MessageAddress) bool {
	if sender == nil {
		return len(f.SenderAddresses) == 0 && len(f.SenderDomains) == 0 && len(f.SenderPatterns) == 0
	}

	senderEmail := strings.ToLower(sender.Address)

	// Check exact address matches
	for _, addr := range f.SenderAddresses {
		if senderEmail == addr {
			return true
		}
	}

	// Check domain matches
	for _, domain := range f.SenderDomains {
		if strings.HasSuffix(senderEmail, "@"+domain) {
			return true
		}
	}

	// Check pattern matches
	for _, pattern := range f.SenderPatterns {
		if pattern.MatchString(senderEmail) {
			return true
		}
	}

	return false
}

// matchesLabels checks if the message has any of the specified labels
func (f *ExportFilter) matchesLabels(labelIDs []string) bool {
	// Check label ID matches
	for _, msgLabelID := range labelIDs {
		for _, filterLabelID := range f.LabelIDs {
			if msgLabelID == filterLabelID {
				return true
			}
		}
	}

	// Note: Label name matching requires resolving label IDs to names
	// This will be implemented in the metadata stage where we have access to label data
	return len(f.LabelIDs) == 0 && len(f.LabelNames) == 0
}

// ToProtonFilter converts the export filter to a proton.MessageFilter for API calls
// Note: Not all filter criteria can be applied at the API level
func (f *ExportFilter) ToProtonFilter() proton.MessageFilter {
	filter := proton.MessageFilter{
		Desc: true, // Always sort by descending date
	}

	// Apply date filtering if supported by API
	if f.DateStart != nil {
		filter.Begin = f.DateStart.Unix()
	}
	if f.DateEnd != nil {
		filter.End = f.DateEnd.Unix()
	}

	// Apply label filtering if supported by API
	if len(f.LabelIDs) > 0 {
		// Use the first label ID for API filtering
		// Multiple labels will be handled by client-side filtering
		filter.LabelID = f.LabelIDs[0]
	}

	return filter
}

// RequiresClientSideFiltering returns true if the filter contains criteria
// that cannot be applied at the API level and requires client-side filtering
func (f *ExportFilter) RequiresClientSideFiltering() bool {
	return len(f.SenderAddresses) > 0 ||
		len(f.SenderDomains) > 0 ||
		len(f.SenderPatterns) > 0 ||
		len(f.LabelIDs) > 1 ||
		len(f.LabelNames) > 0 ||
		len(f.FolderIDs) > 0 ||
		len(f.FolderNames) > 0 ||
		f.HasAttachments != nil ||
		len(f.SubjectPatterns) > 0 ||
		len(f.BodyPatterns) > 0
}