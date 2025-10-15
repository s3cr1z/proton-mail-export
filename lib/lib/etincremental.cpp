// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#include "etincremental.hpp"
#include <fstream>
#include <nlohmann/json.hpp>
#include <stdexcept>

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
            throw std::runtime_error("Failed to open state file: " + mStatePath.string());
        }
        
        nlohmann::json j;
        file >> j;
        
        if (j.contains("messages")) {
            for (const auto& messageEntry : j["messages"]) {
                if (messageEntry.contains("messageId") && messageEntry.contains("checksum")) {
                    std::string key = messageEntry["messageId"].get<std::string>() + ":" + 
                                     messageEntry["checksum"].get<std::string>();
                    mMessages.insert(key);
                }
            }
        }
        
        if (j.contains("lastBackupTime")) {
            uint64_t timestamp = j["lastBackupTime"];
            mLastBackupTime = std::chrono::system_clock::from_time_t(timestamp);
        }
        
    } catch (const std::exception& e) {
        throw std::runtime_error("Failed to load incremental state: " + std::string(e.what()));
    }
}

void IncrementalState::saveState() {
    try {
        // Ensure the directory exists
        std::filesystem::create_directories(mStatePath.parent_path());
        
        nlohmann::json j;
        
        // Save messages
        j["messages"] = nlohmann::json::array();
        for (const auto& messageKey : mMessages) {
            size_t colonPos = messageKey.find(':');
            if (colonPos != std::string::npos) {
                nlohmann::json messageEntry;
                messageEntry["messageId"] = messageKey.substr(0, colonPos);
                messageEntry["checksum"] = messageKey.substr(colonPos + 1);
                j["messages"].push_back(messageEntry);
            }
        }
        
        // Save last backup time
        if (mLastBackupTime) {
            j["lastBackupTime"] = std::chrono::system_clock::to_time_t(*mLastBackupTime);
        }
        
        // Save metadata
        j["version"] = "1.0";
        j["timestamp"] = std::chrono::system_clock::to_time_t(std::chrono::system_clock::now());
        
        std::ofstream file(mStatePath);
        if (!file) {
            throw std::runtime_error("Failed to open state file for writing: " + mStatePath.string());
        }
        
        file << j.dump(4);
        
    } catch (const std::exception& e) {
        throw std::runtime_error("Failed to save incremental state: " + std::string(e.what()));
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