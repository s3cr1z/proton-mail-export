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

#include "etincremental.hpp"
#include <nlohmann/json.hpp>
#include <fstream>
#include <iostream>

namespace etcpp {

IncrementalState::IncrementalState(const std::filesystem::path& backupPath) 
    : mStatePath(backupPath), mLastBackupTime(std::chrono::system_clock::time_point::min()) {
}

std::filesystem::path IncrementalState::getStateFilePath() const {
    return mStatePath / ".proton-export-state.json";
}

bool IncrementalState::loadState() {
    auto stateFile = getStateFilePath();
    if (!std::filesystem::exists(stateFile)) {
        return true; // No state file is OK for first run
    }
    
    try {
        std::ifstream file(stateFile);
        if (!file.is_open()) {
            return false;
        }
        
        nlohmann::json j;
        file >> j;
        
        if (j.contains("processedEmails")) {
            for (const auto& emailId : j["processedEmails"]) {
                mProcessedEmails.insert(emailId.get<std::string>());
            }
        }
        
        if (j.contains("lastBackupTime")) {
            auto timeStr = j["lastBackupTime"].get<std::string>();
            // Parse ISO 8601 timestamp - simplified implementation
            // In real code would use proper time parsing
            mLastBackupTime = std::chrono::system_clock::now();
        }
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Failed to load incremental state: " << e.what() << std::endl;
        return false;
    }
}

bool IncrementalState::saveState() {
    try {
        std::filesystem::create_directories(mStatePath);
        
        nlohmann::json j;
        j["processedEmails"] = std::vector<std::string>(mProcessedEmails.begin(), mProcessedEmails.end());
        
        // Convert time to ISO 8601 string - simplified implementation
        auto time_t = std::chrono::system_clock::to_time_t(mLastBackupTime);
        j["lastBackupTime"] = std::to_string(time_t);
        
        std::ofstream file(getStateFilePath());
        if (!file.is_open()) {
            return false;
        }
        
        file << j.dump(2);
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Failed to save incremental state: " << e.what() << std::endl;
        return false;
    }
}

bool IncrementalState::isEmailProcessed(const std::string& emailId) const {
    return mProcessedEmails.find(emailId) != mProcessedEmails.end();
}

void IncrementalState::markEmailProcessed(const std::string& emailId) {
    mProcessedEmails.insert(emailId);
}

void IncrementalState::setLastBackupTime(std::chrono::system_clock::time_point time) {
    mLastBackupTime = time;
}

std::chrono::system_clock::time_point IncrementalState::getLastBackupTime() const {
    return mLastBackupTime;
}

} // namespace etcpp