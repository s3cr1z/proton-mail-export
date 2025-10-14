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
	"strings"
	"time"

	"github.com/ProtonMail/go-proton-api"
)

// ExportFilter represents the filtering configuration for email export
type ExportFilter struct {
	// Date range filtering
	DateRange *DateRangeFilter `json:"dateRange,omitempty"`
	
	// Label/folder filtering
	Labels *LabelFilter `json:"labels,omitempty"`
	
	// Sender/recipient filtering
	Contacts *ContactFilter `json:"contacts,omitempty"`
	
	// Domain filtering
	Domains *DomainFilter `json:"domains,omitempty"`
}

// DateRangeFilter represents date-based filtering options
type DateRangeFilter struct {
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

// LabelFilter represents label/folder-based filtering options
type LabelFilter struct {
	IncludeLabels []string `json:"includeLabels,omitempty"` // Label IDs to include
	ExcludeLabels []string `json:"excludeLabels,omitempty"` // Label IDs to exclude
	IncludeFolders []string `json:"includeFolders,omitempty"` // Folder IDs to include
	ExcludeFolders []string `json:"excludeFolders,omitempty"` // Folder IDs to exclude
}

// ContactFilter represents sender/recipient-based filtering options
type ContactFilter struct {
	IncludeSenders    []string `json:"includeSenders,omitempty"`    // Email addresses to include as senders
	ExcludeSenders    []string `json:"excludeSenders,omitempty"`    // Email addresses to exclude as senders
	IncludeRecipients []string `json:"includeRecipients,omitempty"` // Email addresses to include as recipients
	ExcludeRecipients []string `json:"excludeRecipients,omitempty"` // Email addresses to exclude as recipients
}

// DomainFilter represents domain-based filtering options
type DomainFilter struct {
	IncludeDomains []string `json:"includeDomains,omitempty"` // Domains to include
	ExcludeDomains []string `json:"excludeDomains,omitempty"` // Domains to exclude
}

// IsEmpty returns true if the filter has no criteria set
func (f *ExportFilter) IsEmpty() bool {
	return f.DateRange == nil && f.Labels == nil && f.Contacts == nil && f.Domains == nil
}

// HasDateFilter returns true if date filtering is configured
func (f *ExportFilter) HasDateFilter() bool {
	return f.DateRange != nil && (f.DateRange.StartDate != nil || f.DateRange.EndDate != nil)
}

// HasLabelFilter returns true if label filtering is configured
func (f *ExportFilter) HasLabelFilter() bool {
	return f.Labels != nil && (len(f.Labels.IncludeLabels) > 0 || len(f.Labels.ExcludeLabels) > 0 ||
		len(f.Labels.IncludeFolders) > 0 || len(f.Labels.ExcludeFolders) > 0)
}

// HasContactFilter returns true if contact filtering is configured
func (f *ExportFilter) HasContactFilter() bool {
	return f.Contacts != nil && (len(f.Contacts.IncludeSenders) > 0 || len(f.Contacts.ExcludeSenders) > 0 ||
		len(f.Contacts.IncludeRecipients) > 0 || len(f.Contacts.ExcludeRecipients) > 0)
}

// HasDomainFilter returns true if domain filtering is configured
func (f *ExportFilter) HasDomainFilter() bool {
	return f.Domains != nil && (len(f.Domains.IncludeDomains) > 0 || len(f.Domains.ExcludeDomains) > 0)
}

// MatchesMessage checks if a message matches the filter criteria
// This is used for client-side filtering when server-side filtering is not available
func (f *ExportFilter) MatchesMessage(msg proton.MessageMetadata) bool {
	if f.IsEmpty() {
		return true
	}

	// Check date filter
	if f.HasDateFilter() && !f.matchesDateFilter(msg) {
		return false
	}

	// Check label filter
	if f.HasLabelFilter() && !f.matchesLabelFilter(msg) {
		return false
	}

	// Check contact filter
	if f.HasContactFilter() && !f.matchesContactFilter(msg) {
		return false
	}

	// Check domain filter
	if f.HasDomainFilter() && !f.matchesDomainFilter(msg) {
		return false
	}

	return true
}

// matchesDateFilter checks if the message matches the date filter criteria
func (f *ExportFilter) matchesDateFilter(msg proton.MessageMetadata) bool {
	if f.DateRange == nil {
		return true
	}

	msgTime := time.Unix(msg.Time, 0)

	if f.DateRange.StartDate != nil && msgTime.Before(*f.DateRange.StartDate) {
		return false
	}

	if f.DateRange.EndDate != nil && msgTime.After(*f.DateRange.EndDate) {
		return false
	}

	return true
}

// matchesLabelFilter checks if the message matches the label filter criteria
func (f *ExportFilter) matchesLabelFilter(msg proton.MessageMetadata) bool {
	if f.Labels == nil {
		return true
	}

	msgLabels := make(map[string]bool)
	for _, labelID := range msg.LabelIDs {
		msgLabels[labelID] = true
	}

	// Check include labels - if specified, message must have at least one
	if len(f.Labels.IncludeLabels) > 0 {
		hasIncludeLabel := false
		for _, labelID := range f.Labels.IncludeLabels {
			if msgLabels[labelID] {
				hasIncludeLabel = true
				break
			}
		}
		if !hasIncludeLabel {
			return false
		}
	}

	// Check include folders - if specified, message must have at least one
	if len(f.Labels.IncludeFolders) > 0 {
		hasIncludeFolder := false
		for _, folderID := range f.Labels.IncludeFolders {
			if msgLabels[folderID] {
				hasIncludeFolder = true
				break
			}
		}
		if !hasIncludeFolder {
			return false
		}
	}

	// Check exclude labels - if message has any, exclude it
	for _, labelID := range f.Labels.ExcludeLabels {
		if msgLabels[labelID] {
			return false
		}
	}

	// Check exclude folders - if message has any, exclude it
	for _, folderID := range f.Labels.ExcludeFolders {
		if msgLabels[folderID] {
			return false
		}
	}

	return true
}

// matchesContactFilter checks if the message matches the contact filter criteria
func (f *ExportFilter) matchesContactFilter(msg proton.MessageMetadata) bool {
	if f.Contacts == nil {
		return true
	}

	// Extract sender email
	senderEmail := extractEmailFromAddress(msg.Sender.Address)

	// Extract recipient emails
	var recipientEmails []string
	for _, recipient := range msg.ToList {
		recipientEmails = append(recipientEmails, extractEmailFromAddress(recipient.Address))
	}
	for _, recipient := range msg.CCList {
		recipientEmails = append(recipientEmails, extractEmailFromAddress(recipient.Address))
	}
	for _, recipient := range msg.BCCList {
		recipientEmails = append(recipientEmails, extractEmailFromAddress(recipient.Address))
	}

	// Check include senders
	if len(f.Contacts.IncludeSenders) > 0 {
		found := false
		for _, includeSender := range f.Contacts.IncludeSenders {
			if strings.EqualFold(senderEmail, includeSender) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check exclude senders
	for _, excludeSender := range f.Contacts.ExcludeSenders {
		if strings.EqualFold(senderEmail, excludeSender) {
			return false
		}
	}

	// Check include recipients
	if len(f.Contacts.IncludeRecipients) > 0 {
		found := false
		for _, includeRecipient := range f.Contacts.IncludeRecipients {
			for _, recipientEmail := range recipientEmails {
				if strings.EqualFold(recipientEmail, includeRecipient) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check exclude recipients
	for _, excludeRecipient := range f.Contacts.ExcludeRecipients {
		for _, recipientEmail := range recipientEmails {
			if strings.EqualFold(recipientEmail, excludeRecipient) {
				return false
			}
		}
	}

	return true
}

// matchesDomainFilter checks if the message matches the domain filter criteria
func (f *ExportFilter) matchesDomainFilter(msg proton.MessageMetadata) bool {
	if f.Domains == nil {
		return true
	}

	// Extract domains from sender and recipients
	var domains []string
	
	// Sender domain
	senderDomain := extractDomainFromAddress(msg.Sender.Address)
	if senderDomain != "" {
		domains = append(domains, senderDomain)
	}

	// Recipient domains
	for _, recipient := range msg.ToList {
		if domain := extractDomainFromAddress(recipient.Address); domain != "" {
			domains = append(domains, domain)
		}
	}
	for _, recipient := range msg.CCList {
		if domain := extractDomainFromAddress(recipient.Address); domain != "" {
			domains = append(domains, domain)
		}
	}
	for _, recipient := range msg.BCCList {
		if domain := extractDomainFromAddress(recipient.Address); domain != "" {
			domains = append(domains, domain)
		}
	}

	// Check include domains
	if len(f.Domains.IncludeDomains) > 0 {
		found := false
		for _, includeDomain := range f.Domains.IncludeDomains {
			for _, domain := range domains {
				if strings.EqualFold(domain, includeDomain) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check exclude domains
	for _, excludeDomain := range f.Domains.ExcludeDomains {
		for _, domain := range domains {
			if strings.EqualFold(domain, excludeDomain) {
				return false
			}
		}
	}

	return true
}

// extractEmailFromAddress extracts the email address from a potentially formatted address
func extractEmailFromAddress(address string) string {
	// Handle cases like "Name <email@domain.com>" or just "email@domain.com"
	if strings.Contains(address, "<") && strings.Contains(address, ">") {
		start := strings.Index(address, "<")
		end := strings.Index(address, ">")
		if start < end {
			return strings.TrimSpace(address[start+1 : end])
		}
	}
	return strings.TrimSpace(address)
}

// extractDomainFromAddress extracts the domain from an email address
func extractDomainFromAddress(address string) string {
	email := extractEmailFromAddress(address)
	if strings.Contains(email, "@") {
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			return strings.ToLower(strings.TrimSpace(parts[1]))
		}
	}
	return ""
}