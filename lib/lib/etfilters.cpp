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
#include <algorithm>
#include <sstream>
#include <iomanip>
#include <filesystem>
#include <fstream>

namespace etcpp {

// FilterCriteria implementation
void FilterCriteria::setDateRange(std::optional<std::chrono::system_clock::time_point> start,
                                 std::optional<std::chrono::system_clock::time_point> end) {
    mStartDate = start;
    mEndDate = end;
}

void FilterCriteria::addSenderFilter(const std::string& senderPattern, bool isRegex) {
    if (isRegex) {
        try {
            mSenderRegexes.emplace_back(senderPattern, std::regex_constants::icase);
        } catch (const std::regex_error& e) {
            throw std::invalid_argument("Invalid sender regex pattern: " + senderPattern);
        }
    } else {
        mSenderPatterns.push_back(senderPattern);
    }
}

void FilterCriteria::addRecipientFilter(const std::string& recipientPattern, bool isRegex) {
    if (isRegex) {
        try {
            mRecipientRegexes.emplace_back(recipientPattern, std::regex_constants::icase);
        } catch (const std::regex_error& e) {
            throw std::invalid_argument("Invalid recipient regex pattern: " + recipientPattern);
        }
    } else {
        mRecipientPatterns.push_back(recipientPattern);
    }
}

void FilterCriteria::addSubjectFilter(const std::string& subjectPattern, bool isRegex) {
    if (isRegex) {
        try {
            mSubjectRegexes.emplace_back(subjectPattern, std::regex_constants::icase);
        } catch (const std::regex_error& e) {
            throw std::invalid_argument("Invalid subject regex pattern: " + subjectPattern);
        }
    } else {
        mSubjectPatterns.push_back(subjectPattern);
    }
}

void FilterCriteria::setHasAttachments(bool hasAttachments) {
    mHasAttachments = hasAttachments;
}

void FilterCriteria::addAttachmentNameFilter(const std::string& namePattern, bool isRegex) {
    if (isRegex) {
        try {
            mAttachmentNameRegexes.emplace_back(namePattern, std::regex_constants::icase);
        } catch (const std::regex_error& e) {
            throw std::invalid_argument("Invalid attachment name regex pattern: " + namePattern);
        }
    } else {
        mAttachmentNamePatterns.push_back(namePattern);
    }
}

void FilterCriteria::setSizeRange(std::optional<uint64_t> minSize, std::optional<uint64_t> maxSize) {
    mMinSize = minSize;
    mMaxSize = maxSize;
}

void FilterCriteria::addFolderFilter(const std::string& folderName) {
    mFolders.push_back(folderName);
}

void FilterCriteria::addCustomFilter(std::function<bool(const EmailMessage&)> filter) {
    mCustomFilters.push_back(std::move(filter));
}