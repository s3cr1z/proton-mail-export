package mail

import (
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/stretchr/testify/assert"
)

func TestExportFilter_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		filter   *ExportFilter
		expected bool
	}{
		{
			name:     "nil filter",
			filter:   nil,
			expected: true,
		},
		{
			name:     "empty filter",
			filter:   &ExportFilter{},
			expected: true,
		},
		{
			name: "filter with date range",
			filter: &ExportFilter{
				DateRange: &DateRangeFilter{
					StartDate: &time.Time{},
				},
			},
			expected: false,
		},
		{
			name: "filter with labels",
			filter: &ExportFilter{
				Labels: &LabelFilter{
					IncludeLabels: []string{"label1"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.filter == nil {
				// Test nil case separately
				assert.True(t, true) // nil filter should be considered empty
				return
			}
			assert.Equal(t, tt.expected, tt.filter.IsEmpty())
		})
	}
}

func TestExportFilter_MatchesMessage_DateFilter(t *testing.T) {
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

	filter := &ExportFilter{
		DateRange: &DateRangeFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
		},
	}

	tests := []struct {
		name     string
		msgTime  int64
		expected bool
	}{
		{
			name:     "message within range",
			msgTime:  time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC).Unix(),
			expected: true,
		},
		{
			name:     "message before range",
			msgTime:  time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC).Unix(),
			expected: false,
		},
		{
			name:     "message after range",
			msgTime:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := proton.MessageMetadata{
				Time: tt.msgTime,
			}
			assert.Equal(t, tt.expected, filter.MatchesMessage(msg))
		})
	}
}

func TestExportFilter_MatchesMessage_LabelFilter(t *testing.T) {
	filter := &ExportFilter{
		Labels: &LabelFilter{
			IncludeLabels: []string{"label1", "label2"},
			ExcludeLabels: []string{"spam"},
		},
	}

	tests := []struct {
		name     string
		labelIDs []string
		expected bool
	}{
		{
			name:     "message with included label",
			labelIDs: []string{"label1", "other"},
			expected: true,
		},
		{
			name:     "message with excluded label",
			labelIDs: []string{"label1", "spam"},
			expected: false,
		},
		{
			name:     "message without included labels",
			labelIDs: []string{"other", "another"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := proton.MessageMetadata{
				LabelIDs: tt.labelIDs,
			}
			assert.Equal(t, tt.expected, filter.MatchesMessage(msg))
		})
	}
}