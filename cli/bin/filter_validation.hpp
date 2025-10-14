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
#include <optional>
#include <chrono>
#include <regex>
#include <stdexcept>

namespace etcli {

class FilterValidationException : public std::runtime_error {
public:
    explicit FilterValidationException(const std::string& message) : std::runtime_error(message) {}
};

struct FilterCriteria {
    std::optional<std::chrono::system_clock::time_point> sinceDate;
    std::optional<std::chrono::system_clock::time_point> untilDate;
    std::vector<std::string> addresses;
    std::vector<std::string> domains;
    std::vector<std::string> folders;
    std::vector<std::string> labels;
    std::vector<std::string> subjects;
    std::optional<bool> hasAttachments;
    std::optional<uint64_t> minSize;
    std::optional<uint64_t> maxSize;
    
    // Legacy support
    bool hasLegacyOptions = false;
    std::string legacyDateStart;
    std::string legacyDateEnd;
    std::vector<std::string> legacySenders;
};

// Date validation functions
std::chrono::system_clock::time_point parseDateTime(const std::string& dateStr);
bool isValidDateFormat(const std::string& dateStr);

// Email validation functions
bool isValidEmailPattern(const std::string& pattern);
bool isValidDomainPattern(const std::string& domain);

// Size validation functions
uint64_t parseSize(const std::string& sizeStr);
bool isValidSizeFormat(const std::string& sizeStr);

// Pattern validation functions
bool isValidWildcardPattern(const std::string& pattern);

// Main validation function
FilterCriteria validateAndBuildFilterCriteria(const cxxopts::ParseResult& args);

// Error message helpers
std::string getDateFormatHelp();
std::string getSizeFormatHelp();
std::string getPatternFormatHelp();

// Legacy option handling
void handleLegacyOptions(FilterCriteria& criteria, const cxxopts::ParseResult& args);

// Usage examples
void printFilterExamples();

} // namespace etcli