// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#include "etincremental.hpp"
#include <fstream>
#include <nlohmann/json.hpp>
#include <chrono>

namespace etcpp {

IncrementalState::IncrementalState(const std::filesystem::path& backupPath) 
    : mStatePath(backupPath / ".incremental_state.json") {
}

void IncrementalState::loadState() {
    if (!std::filesystem::exists(mStatePath)) {
        // No previous state, start fresh
        return;
    }
    
    try {
        std::ifstream file(mStatePath);
        if (!file) {
            return; // Can't read, start fresh
        }
        
        nlohmann::json state;
        file >> state;
        
        if (state.contains("messages")) {
            for (const auto& messageEntry : state["messages"]) {
                if (messageEntry.contains("messageId") && messageEntry.contains("checksum")) {
                    std::string key = messageEntry["messageId"].get<std::string>() + ":" + 
                                     messageEntry["checksum"].get<std::string>();
                    mMessages.insert(key);
                }
            }
        }
        
        if (state.contains("lastBackupTime")) {
            uint64_t timestamp = state["lastBackupTime"];
            mLastBackupTime = std::chrono::system_clock::from_time_t(timestamp);
        }
        
    } catch (const std::exception&) {
        // If we can't parse the state file, start fresh
        mMessages.clear();
        mLastBackupTime.reset();
    }
}

void IncrementalState::saveState() {
    try {
        // Ensure directory exists
        std::filesystem::create_directories(mStatePath.parent_path());
        
        nlohmann::json state;
        
        // Save messages
        nlohmann::json messages = nlohmann::json::array();
        for (const auto& messageKey : mMessages) {
            size_t colonPos = messageKey.find(':');
            if (colonPos != std::string::npos) {
                nlohmann::json messageEntry;
                messageEntry["messageId"] = messageKey.substr(0, colonPos);
                messageEntry["checksum"] = messageKey.substr(colonPos + 1);
                messages.push_back(messageEntry);
            }
        }
        state["messages"] = messages;
        
        // Save last backup time
        if (mLastBackupTime) {
            auto timestamp = std::chrono::system_clock::to_time_t(*mLastBackupTime);
            state["lastBackupTime"] = static_cast<uint64_t>(timestamp);
        }
        
        std::ofstream file(mStatePath);
        if (file) {
            file << state.dump(4);
        }
        
    } catch (const std::exception&) {
        // If we can't save state, continue anyway
        // The next backup will just be less efficient
    }
}

bool IncrementalState::shouldBackupMessage(const MessageMetadata& metadata) const {
    std::string key = metadata.messageId + ":" + metadata.checksum;
    return mMessages.find(key) == mMessages.end();
}

void IncrementalState::addMessage(const MessageMetadata& metadata) {
    std::string key = metadata.messageId + ":" + metadata.checksum;
    mMessages.insert(key);
}

std::optional<std::chrono::system_clock::time_point> IncrementalState::getLastBackupTime() const {
    return mLastBackupTime;
}

void IncrementalState::setLastBackupTime(std::chrono::system_clock::time_point time) {
    mLastBackupTime = time;
}

std::string IncrementalState::generateStateFilePath() const {
    return mStatePath.string();
}

IncrementalBackupFilter::IncrementalBackupFilter(IncrementalState& state) 
    : mState(state) {
}

bool IncrementalBackupFilter::shouldIncludeMessage(const MessageMetadata& metadata) const {
    return mState.shouldBackupMessage(metadata);
}

void IncrementalBackupFilter::markMessageProcessed(const MessageMetadata& metadata) {
    mState.addMessage(metadata);
}

} // namespace etcpp