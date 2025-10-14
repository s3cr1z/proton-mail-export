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
        "context"

        "github.com/ProtonMail/export-tool/internal/apiclient"
        "github.com/ProtonMail/go-proton-api"
        "github.com/bradenaw/juniper/xslices"
        "github.com/sirupsen/logrus"
)

type MetadataFileChecker interface {
        HasMessage(msgID string) (bool, error)
}

type MetadataStage struct {
        client    apiclient.Client
        log       *logrus.Entry
        outputCh  chan []proton.MessageMetadata
        pageSize  int
        splitSize int
        filter    *ExportFilter
}

func NewMetadataStage(
        client apiclient.Client,
        entry *logrus.Entry,
        pageSize int,
        splitSize int,
        filter *ExportFilter,
) *MetadataStage {
        return &MetadataStage{
                client:    client,
                log:       entry.WithField("stage", "metadata"),
                outputCh:  make(chan []proton.MessageMetadata),
                pageSize:  pageSize,
                splitSize: splitSize,
                filter:    filter,
        }
}

func (m *MetadataStage) Run(
        ctx context.Context,
        errReporter StageErrorReporter,
        mfc MetadataFileChecker,
        reporter Reporter,
) {
        m.log.Debug("Starting")
        defer m.log.Debug("Exiting")
        defer close(m.outputCh)

        client := m.client

        var lastMessageID string

        for {
                if ctx.Err() != nil {
                        return
                }

                var metadata []proton.MessageMetadata

                // Build the message filter with server-side filtering where possible
                messageFilter := m.buildMessageFilter(lastMessageID)

                if lastMessageID != "" {
                        meta, err := client.GetMessageMetadataPage(ctx, 0, m.pageSize, messageFilter)

                        if err != nil {
                                errReporter.ReportStageError(err)
                                return
                        }

                        // * There is only one message returned and it matches the EndID query.
                        if len(meta) != 0 && meta[0].ID == lastMessageID {
                                meta = meta[1:]
                        }

                        metadata = meta
                } else {
                        meta, err := client.GetMessageMetadataPage(ctx, 0, m.pageSize, messageFilter)
                        if err != nil {
                                errReporter.ReportStageError(err)
                                return
                        }
                        metadata = meta
                }

                // Nothing left to do
                if len(metadata) == 0 {
                        return
                }

                lastMessageID = metadata[len(metadata)-1].ID

                initialLen := len(metadata)
                
                // Apply client-side filtering for criteria not supported server-side
                if m.filter != nil && !m.filter.IsEmpty() {
                        metadata = m.applyClientSideFiltering(metadata)
                }
                
                // Filter out messages that already exist
                metadata = xslices.Filter(metadata, func(t proton.MessageMetadata) bool {
                        isPresent, err := mfc.HasMessage(t.ID)
                        if err != nil {
                                errReporter.ReportStageError(err)
                                return false
                        }

                        return !isPresent
                })

                if len(metadata) != initialLen {
                        reporter.OnProgress(initialLen - len(metadata))
                }

                if len(metadata) == 0 {
                        continue
                }

                for _, chunk := range xslices.Chunk(metadata, m.splitSize) {
                        select {
                        case <-ctx.Done():
                                return
                        case m.outputCh <- chunk:
                        }
                }
        }
}

type alwaysMissingMetadataFileChecker struct{}

func (a alwaysMissingMetadataFileChecker) HasMessage(string) (bool, error) {
        return false, nil
}

// buildMessageFilter creates a proton.MessageFilter with server-side filtering where supported
func (m *MetadataStage) buildMessageFilter(lastMessageID string) proton.MessageFilter {
        filter := proton.MessageFilter{
                Desc: true,
        }
        
        if lastMessageID != "" {
                filter.EndID = lastMessageID
        }
        
        // TODO: Add server-side filtering capabilities here when supported by proton.MessageFilter
        // For now, we'll rely on client-side filtering for all advanced criteria
        
        return filter
}

// applyClientSideFiltering applies filtering criteria that cannot be handled server-side
func (m *MetadataStage) applyClientSideFiltering(metadata []proton.MessageMetadata) []proton.MessageMetadata {
        if m.filter == nil || m.filter.IsEmpty() {
                return metadata
        }
        
        return xslices.Filter(metadata, func(msg proton.MessageMetadata) bool {
                return m.filter.MatchesMessage(msg)
        })
}
