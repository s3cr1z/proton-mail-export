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
#include <vector>
#include <chrono>
#include <optional>
#include <regex>

namespace etcpp {

struct FilterCriteria {
    std::optional<std::chrono::system_clock::time_point> dateStart;
    std::optional<std::chrono::system_clock::time_point> dateEnd;
    std::vector<std::string> senderFilters;
    std::vector<std::string> folderFilters;
    std::optional<bool> hasAttachments;
    std::optional<uint64_t> minSize;
    std::optional<uint64_t> maxSize;
    
    bool isEmpty() const;
};

class EmailFilter {
public:
    explicit EmailFilter(const FilterCriteria& criteria);
    
    bool shouldIncludeEmail(const std::string& sender,
                           const std::string& folder,
                           std::chrono::system_clock::time_point date,
                           bool hasAttachments,
                           uint64_t size) const;
    
private:
    FilterCriteria mCriteria;
    std::vector<std::regex> mSenderRegexes;
    std::vector<std::regex> mFolderRegexes;
    
    void compileSenderRegexes();
    void compileFolderRegexes();
    bool matchesAnyPattern(const std::string& text, const std::vector<std::regex>& patterns) const;
};

FilterCriteria parseFilterOptions(const std::string& dateStart,
                                 const std::string& dateEnd,
                                 const std::vector<std::string>& senders,
                                 const std::vector<std::string>& folders,
                                 bool hasAttachments,
                                 uint64_t minSize);

} // namespace etcpp