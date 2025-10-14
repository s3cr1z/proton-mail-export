package filters

import (
	"regexp"
	"strings"
	"time"
)

// EmailFilter represents a complete email filtering configuration
type EmailFilter struct {
	DateStart      *time.Time
	DateEnd        *time.Time
	Senders        []string
	Recipients     []string
	Domains        []string
	Labels         []string
	Folders        []string
	HasAttachments *bool
	MinSize        *int64
	MaxSize        *int64
	SearchQuery    string
	SearchRegex    *regexp.Regexp
}

// EmailData represents email metadata for filtering
type EmailData struct {
	Subject     string
	From        string
	To          []string
	CC          []string
	BCC         []string
	Date        time.Time
	Body        string
	Labels      []string
	Folder      string
	Size        int64
	Attachments []string
	MessageID   string
}

// NewEmailFilter creates a new email filter
func NewEmailFilter() *EmailFilter {
	return &EmailFilter{}
}

// Matches checks if an email matches all filter criteria
func (f *EmailFilter) Matches(email EmailData) bool {
	// Date range filter
	if f.DateStart != nil && email.Date.Before(*f.DateStart) {
		return false
	}
	if f.DateEnd != nil && email.Date.After(*f.DateEnd) {
		return false
	}
	
	// Sender filter
	if len(f.Senders) > 0 && !f.matchesAnyPattern(email.From, f.Senders) {
		return false
	}
	
	// Recipient filter
	if len(f.Recipients) > 0 {
		allRecipients := append(email.To, email.CC...)
		allRecipients = append(allRecipients, email.BCC...)
		if !f.matchesAnyRecipient(allRecipients, f.Recipients) {
			return false
		}
	}
	
	// Domain filter
	if len(f.Domains) > 0 && !f.matchesDomain(email, f.Domains) {
		return false
	}
	
	// Label filter
	if len(f.Labels) > 0 && !f.matchesAnyString(email.Labels, f.Labels) {
		return false
	}
	
	// Folder filter
	if len(f.Folders) > 0 && !f.matchesAnyPattern(email.Folder, f.Folders) {
		return false
	}
	
	// Attachment filter
	if f.HasAttachments != nil {
		hasAttachments := len(email.Attachments) > 0
		if *f.HasAttachments != hasAttachments {
			return false
		}
	}
	
	// Size filter
	if f.MinSize != nil && email.Size < *f.MinSize {
		return false
	}
	if f.MaxSize != nil && email.Size > *f.MaxSize {
		return false
	}
	
	// Search query filter
	if f.SearchQuery != "" && !f.matchesSearchQuery(email) {
		return false
	}
	
	return true
}

func (f *EmailFilter) matchesAnyPattern(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if f.matchesPattern(text, pattern) {
			return true
		}
	}
	return false
}

func (f *EmailFilter) matchesPattern(text, pattern string) bool {
	// Support wildcard patterns
	if strings.Contains(pattern, "*") {
		// Convert wildcard to regex
		regexPattern := strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*")
		regex, err := regexp.Compile("(?i)^" + regexPattern + "$")
		if err != nil {
			// Fallback to simple contains check
			return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
		}
		return regex.MatchString(text)
	}
	
	// Simple case-insensitive contains check
	return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
}

func (f *EmailFilter) matchesAnyRecipient(recipients, patterns []string) bool {
	for _, recipient := range recipients {
		if f.matchesAnyPattern(recipient, patterns) {
			return true
		}
	}
	return false
}

func (f *EmailFilter) matchesDomain(email EmailData, domains []string) bool {
	// Check sender domain
	if f.matchesDomainInEmail(email.From, domains) {
		return true
	}
	
	// Check recipient domains
	allRecipients := append(email.To, email.CC...)
	allRecipients = append(allRecipients, email.BCC...)
	
	for _, recipient := range allRecipients {
		if f.matchesDomainInEmail(recipient, domains) {
			return true
		}
	}
	
	return false
}

func (f *EmailFilter) matchesDomainInEmail(emailAddr string, domains []string) bool {
	atIndex := strings.LastIndex(emailAddr, "@")
	if atIndex == -1 {
		return false
	}
	
	emailDomain := emailAddr[atIndex+1:]
	for _, domain := range domains {
		if strings.EqualFold(emailDomain, domain) {
			return true
		}
	}
	
	return false
}

func (f *EmailFilter) matchesAnyString(haystack, needles []string) bool {
	for _, hay := range haystack {
		for _, needle := range needles {
			if strings.Contains(strings.ToLower(hay), strings.ToLower(needle)) {
				return true
			}
		}
	}
	return false
}

func (f *EmailFilter) matchesSearchQuery(email EmailData) bool {
	searchText := strings.ToLower(f.SearchQuery)
	
	// Search in subject
	if strings.Contains(strings.ToLower(email.Subject), searchText) {
		return true
	}
	
	// Search in body
	if strings.Contains(strings.ToLower(email.Body), searchText) {
		return true
	}
	
	// Search in sender
	if strings.Contains(strings.ToLower(email.From), searchText) {
		return true
	}
	
	// Use regex if configured
	if f.SearchRegex != nil {
		if f.SearchRegex.MatchString(email.Subject) ||
			f.SearchRegex.MatchString(email.Body) ||
			f.SearchRegex.MatchString(email.From) {
			return true
		}
	}
	
	return false
}

// SetSearchRegex sets a compiled regex for search queries
func (f *EmailFilter) SetSearchRegex(pattern string) error {
	if pattern == "" {
		f.SearchRegex = nil
		return nil
	}
	
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return err
	}
	
	f.SearchRegex = regex
	return nil
}

// FilterStats represents statistics about filtering results
type FilterStats struct {
	TotalEmails    int
	MatchedEmails  int
	FilteredEmails int
	FilterRatio    float64
}

// CalculateStats calculates filtering statistics
func (f *EmailFilter) CalculateStats(emails []EmailData) FilterStats {
	total := len(emails)
	matched := 0
	
	for _, email := range emails {
		if f.Matches(email) {
			matched++
		}
	}
	
	filtered := total - matched
	ratio := 0.0
	if total > 0 {
		ratio = float64(matched) / float64(total)
	}
	
	return FilterStats{
		TotalEmails:    total,
		MatchedEmails:  matched,
		FilteredEmails: filtered,
		FilterRatio:    ratio,
	}
}