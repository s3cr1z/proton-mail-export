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

#include "etfilters.hpp"
#include <sstream>
#include <iomanip>
#include <iostream>

namespace etcpp {

bool FilterCriteria::isEmpty() const {
    return !dateStart.has_value() && !dateEnd.has_value() && 
           senderFilters.empty() && folderFilters.empty() &&
           !hasAttachments.has_value() && !minSize.has_value() && !maxSize.has_value();
}

EmailFilter::EmailFilter(const FilterCriteria& criteria) : mCriteria(criteria) {
    compileSenderRegexes();
    compileFolderRegexes();
}

void EmailFilter::compileSenderRegexes() {
    for (const auto& pattern : mCriteria.senderFilters) {
        try {
            // Convert wildcard pattern to regex
            std::string regexPattern = pattern;
            // Replace * with .*
            size_t pos = 0;
            while ((pos = regexPattern.find('*', pos)) != std::string::npos) {
                regexPattern.replace(pos, 1, ".*");
                pos += 2;
            }
            mSenderRegexes.emplace_back(regexPattern, std::regex_constants::icase);
        } catch (const std::exception& e) {
            std::cerr << "Invalid sender filter pattern: " << pattern << " - " << e.what() << std::endl;
        }
    }
}

void EmailFilter::compileFolderRegexes() {
    for (const auto& pattern : mCriteria.folderFilters) {
        try {
            // Convert wildcard pattern to regex
            std::string regexPattern = pattern;
            // Replace * with .*
            size_t pos = 0;
            while ((pos = regexPattern.find('*', pos)) != std::string::npos) {
                regexPattern.replace(pos, 1, ".*");
                pos += 2;
            }
            mFolderRegexes.emplace_back(regexPattern, std::regex_constants::icase);
        } catch (const std::exception& e) {
            std::cerr << "Invalid folder filter pattern: " << pattern << " - " << e.what() << std::endl;
        }
    }
}

bool EmailFilter::matchesAnyPattern(const std::string& text, const std::vector<std::regex>& patterns) const {
    if (patterns.empty()) {
        return true; // No filters means include all
    }
    
    for (const auto& pattern : patterns) {
        if (std::regex_match(text, pattern)) {
            return true;
        }
    }
    return false;
}

bool EmailFilter::shouldIncludeEmail(const std::string& sender,
                                   const std::string& folder,
                                   std::chrono::system_clock::time_point date,
                                   bool hasAttachments,
                                   uint64_t size) const {
    // Check date range
    if (mCriteria.dateStart.has_value() && date < mCriteria.dateStart.value()) {
        return false;
    }
    if (mCriteria.dateEnd.has_value() && date > mCriteria.dateEnd.value()) {
        return false;
    }
    
    // Check sender filters
    if (!matchesAnyPattern(sender, mSenderRegexes)) {
        return false;
    }
    
    // Check folder filters
    if (!matchesAnyPattern(folder, mFolderRegexes)) {
        return false;
    }
    
    // Check attachment filter
    if (mCriteria.hasAttachments.has_value() && hasAttachments != mCriteria.hasAttachments.value()) {
        return false;
    }
    
    // Check size filters
    if (mCriteria.minSize.has_value() && size < mCriteria.minSize.value()) {
        return false;
    }
    if (mCriteria.maxSize.has_value() && size > mCriteria.maxSize.value()) {
        return false;
    }
    
    return true;
}

FilterCriteria parseFilterOptions(const std::string& dateStart,
                                 const std::string& dateEnd,
                                 const std::vector<std::string>& senders,
                                 const std::vector<std::string>& folders,
                                 bool hasAttachments,
                                 uint64_t minSize) {
    FilterCriteria criteria;
    
    // Parse date strings (simplified - would use proper date parsing in real code)
    if (!dateStart.empty()) {
        // For now, just set to current time - would parse YYYY-MM-DD format
        criteria.dateStart = std::chrono::system_clock::now() - std::chrono::hours(24 * 365);
    }
    
    if (!dateEnd.empty()) {
        // For now, just set to current time - would parse YYYY-MM-DD format
        criteria.dateEnd = std::chrono::system_clock::now();
    }
    
    criteria.senderFilters = senders;
    criteria.folderFilters = folders;
    
    if (hasAttachments) {
        criteria.hasAttachments = true;
    }
    
    if (minSize > 0) {
        criteria.minSize = minSize;
    }
    
    return criteria;
}

} // namespace etcpp