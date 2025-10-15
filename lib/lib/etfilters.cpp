// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#include "etfilters.hpp"
#include <algorithm>
#include <sstream>
#include <iomanip>
#include <filesystem>
#include <fstream>
#include <nlohmann/json.hpp>

namespace etcpp {

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
            throw std::runtime_error("Invalid sender regex pattern: " + senderPattern);
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
            throw std::runtime_error("Invalid recipient regex pattern: " + recipientPattern);
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
            throw std::runtime_error("Invalid subject regex pattern: " + subjectPattern);
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
            throw std::runtime_error("Invalid attachment name regex pattern: " + namePattern);
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

bool FilterCriteria::matches(const EmailMessage& message) const {
    // Date range check
    if (mStartDate && message.timestamp < *mStartDate) {
        return false;
    }
    if (mEndDate && message.timestamp > *mEndDate) {
        return false;
    }
    
    // Sender filter check
    if (!mSenderPatterns.empty() || !mSenderRegexes.empty()) {
        if (!matchesPattern(message.sender, mSenderPatterns, mSenderRegexes)) {
            return false;
        }
    }
    
    // Recipient filter check
    if (!mRecipientPatterns.empty() || !mRecipientRegexes.empty()) {
        bool matchesAnyRecipient = false;
        for (const auto& recipient : message.recipients) {
            if (matchesPattern(recipient, mRecipientPatterns, mRecipientRegexes)) {
                matchesAnyRecipient = true;
                break;
            }
        }
        if (!matchesAnyRecipient) {
            return false;
        }
    }
    
    // Subject filter check
    if (!mSubjectPatterns.empty() || !mSubjectRegexes.empty()) {
        if (!matchesPattern(message.subject, mSubjectPatterns, mSubjectRegexes)) {
            return false;
        }
    }
    
    // Attachment presence check
    if (mHasAttachments && message.hasAttachments != *mHasAttachments) {
        return false;
    }
    
    // Attachment name filter check
    if (!mAttachmentNamePatterns.empty() || !mAttachmentNameRegexes.empty()) {
        bool matchesAnyAttachment = false;
        for (const auto& attachmentName : message.attachmentNames) {
            if (matchesPattern(attachmentName, mAttachmentNamePatterns, mAttachmentNameRegexes)) {
                matchesAnyAttachment = true;
                break;
            }
        }
        if (!matchesAnyAttachment) {
            return false;
        }
    }
    
    // Size range check
    if (mMinSize && message.size < *mMinSize) {
        return false;
    }
    if (mMaxSize && message.size > *mMaxSize) {
        return false;
    }
    
    // Folder filter check
    if (!mFolders.empty()) {
        auto it = std::find(mFolders.begin(), mFolders.end(), message.folderName);
        if (it == mFolders.end()) {
            return false;
        }
    }
    
    // Custom filters check
    for (const auto& customFilter : mCustomFilters) {
        if (!customFilter(message)) {
            return false;
        }
    }
    
    return true;
}

bool FilterCriteria::matchesPattern(const std::string& text, 
                                   const std::vector<std::string>& patterns,
                                   const std::vector<std::regex>& regexes) const {
    // Check simple patterns (with wildcard support)
    for (const auto& pattern : patterns) {
        if (pattern.find('*') != std::string::npos) {
            // Convert wildcard pattern to regex
            std::string regexPattern = pattern;
            std::replace(regexPattern.begin(), regexPattern.end(), '*', '.');
            regexPattern = ".*" + regexPattern + ".*";
            try {
                std::regex wildcardRegex(regexPattern, std::regex_constants::icase);
                if (std::regex_match(text, wildcardRegex)) {
                    return true;
                }
            } catch (const std::regex_error&) {
                // Fall back to simple string comparison
                if (text.find(pattern) != std::string::npos) {
                    return true;
                }
            }
        } else {
            // Simple case-insensitive substring match
            std::string lowerText = text;
            std::string lowerPattern = pattern;
            std::transform(lowerText.begin(), lowerText.end(), lowerText.begin(), ::tolower);
            std::transform(lowerPattern.begin(), lowerPattern.end(), lowerPattern.begin(), ::tolower);
            if (lowerText.find(lowerPattern) != std::string::npos) {
                return true;
            }
        }
    }
    
    // Check regex patterns
    for (const auto& regex : regexes) {
        if (std::regex_search(text, regex)) {
            return true;
        }
    }
    
    return patterns.empty() && regexes.empty(); // If no patterns, match everything
}

FilterCriteria FilterBuilder::fromCLIArgs(const std::vector<std::string>& args) {
    FilterBuilder builder;
    
    // This is a simplified implementation - in practice, you'd parse the CLI args properly
    // For now, we'll assume the args are in a specific format
    for (size_t i = 0; i < args.size(); i += 2) {
        if (i + 1 >= args.size()) break;
        
        const std::string& key = args[i];
        const std::string& value = args[i + 1];
        
        if (key == "--date-start" && i + 3 < args.size() && args[i + 2] == "--date-end") {
            builder.dateRange(value, args[i + 3]);
            i += 2; // Skip the next pair
        } else if (key == "--sender") {
            builder.sender(value);
        } else if (key == "--recipient") {
            builder.recipient(value);
        } else if (key == "--subject") {
            builder.subject(value);
        } else if (key == "--has-attachments") {
            builder.hasAttachments(value == "true" || value == "1");
        } else if (key == "--folder") {
            builder.folder(value);
        }
    }
    
    return builder.build();
}

FilterCriteria FilterBuilder::fromConfigFile(const std::filesystem::path& configPath) {
    if (!std::filesystem::exists(configPath)) {
        throw std::runtime_error("Config file does not exist: " + configPath.string());
    }
    
    std::ifstream file(configPath);
    if (!file) {
        throw std::runtime_error("Failed to open config file: " + configPath.string());
    }
    
    nlohmann::json j;
    file >> j;
    
    FilterBuilder builder;
    
    if (j.contains("dateStart") && j.contains("dateEnd")) {
        builder.dateRange(j["dateStart"], j["dateEnd"]);
    }
    
    if (j.contains("senderFilters")) {
        for (const auto& sender : j["senderFilters"]) {
            builder.sender(sender);
        }
    }
    
    if (j.contains("recipientFilters")) {
        for (const auto& recipient : j["recipientFilters"]) {
            builder.recipient(recipient);
        }
    }
    
    if (j.contains("subjectFilters")) {
        for (const auto& subject : j["subjectFilters"]) {
            builder.subject(subject);
        }
    }
    
    if (j.contains("hasAttachments")) {
        builder.hasAttachments(j["hasAttachments"]);
    }
    
    if (j.contains("minSize") && j.contains("maxSize")) {
        builder.sizeRange(j["minSize"], j["maxSize"]);
    }
    
    if (j.contains("folders")) {
        for (const auto& folder : j["folders"]) {
            builder.folder(folder);
        }
    }
    
    return builder.build();
}

FilterBuilder& FilterBuilder::dateRange(const std::string& start, const std::string& end) {
    auto startTime = parseDate(start);
    auto endTime = parseDate(end);
    mCriteria.setDateRange(startTime, endTime);
    return *this;
}

FilterBuilder& FilterBuilder::sender(const std::string& pattern, bool isRegex) {
    mCriteria.addSenderFilter(pattern, isRegex);
    return *this;
}

FilterBuilder& FilterBuilder::recipient(const std::string& pattern, bool isRegex) {
    mCriteria.addRecipientFilter(pattern, isRegex);
    return *this;
}

FilterBuilder& FilterBuilder::subject(const std::string& pattern, bool isRegex) {
    mCriteria.addSubjectFilter(pattern, isRegex);
    return *this;
}

FilterBuilder& FilterBuilder::hasAttachments(bool value) {
    mCriteria.setHasAttachments(value);
    return *this;
}

FilterBuilder& FilterBuilder::attachmentName(const std::string& pattern, bool isRegex) {
    mCriteria.addAttachmentNameFilter(pattern, isRegex);
    return *this;
}

FilterBuilder& FilterBuilder::sizeRange(uint64_t minSize, uint64_t maxSize) {
    mCriteria.setSizeRange(minSize, maxSize);
    return *this;
}

FilterBuilder& FilterBuilder::folder(const std::string& folderName) {
    mCriteria.addFolderFilter(folderName);
    return *this;
}

FilterCriteria FilterBuilder::build() const {
    return mCriteria;
}

std::chrono::system_clock::time_point FilterBuilder::parseDate(const std::string& dateStr) const {
    std::tm tm = {};
    std::istringstream ss(dateStr);
    ss >> std::get_time(&tm, "%Y-%m-%d");
    
    if (ss.fail()) {
        throw std::runtime_error("Invalid date format: " + dateStr + " (expected YYYY-MM-DD)");
    }
    
    return std::chrono::system_clock::from_time_t(std::mktime(&tm));
}

} // namespace etcpp