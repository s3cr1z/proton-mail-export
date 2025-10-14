// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#pragma once

#include <string>
#include <filesystem>
#include <unordered_set>
#include <chrono>
#include <optional>

namespace etcpp {

struct MessageMetadata {
    std::string messageId;
    uint64_t timestamp;
    std::string checksum;
    uint64_t size;
    
    bool operator==(const MessageMetadata& other) const {
        return messageId == other.messageId && checksum == other.checksum;
    }
};

class IncrementalState {
public:
    IncrementalState(const std::filesystem::path& backupPath);
    
    void loadState();
    void saveState();
    
    bool shouldBackupMessage(const MessageMetadata& metadata) const;
    void addMessage(const MessageMetadata& metadata);
    
    std::optional<std::chrono::system_clock::time_point> getLastBackupTime() const;
    void setLastBackupTime(std::chrono::system_clock::time_point time);
    
    size_t getMessageCount() const { return mMessages.size(); }
    
private:
    std::filesystem::path mStatePath;
    std::unordered_set<std::string> mMessages; // messageId -> checksum
    std::optional<std::chrono::system_clock::time_point> mLastBackupTime;
    
    std::string generateStateFilePath() const;
};

class IncrementalBackupFilter {
public:
    IncrementalBackupFilter(IncrementalState& state);
    
    bool shouldIncludeMessage(const MessageMetadata& metadata) const;
    void markMessageProcessed(const MessageMetadata& metadata);
    
private:
    IncrementalState& mState;
};

} // namespace etcpp
