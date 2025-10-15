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

#pragma once

#include <string>
#include <unordered_set>
#include <filesystem>
#include <chrono>

namespace etcpp {

class IncrementalState {
public:
    explicit IncrementalState(const std::filesystem::path& backupPath);
    
    bool loadState();
    bool saveState();
    
    bool isEmailProcessed(const std::string& emailId) const;
    void markEmailProcessed(const std::string& emailId);
    
    void setLastBackupTime(std::chrono::system_clock::time_point time);
    std::chrono::system_clock::time_point getLastBackupTime() const;
    
    size_t getProcessedEmailCount() const { return mProcessedEmails.size(); }
    
private:
    std::filesystem::path mStatePath;
    std::unordered_set<std::string> mProcessedEmails;
    std::chrono::system_clock::time_point mLastBackupTime;
    
    std::filesystem::path getStateFilePath() const;
};

} // namespace etcpp