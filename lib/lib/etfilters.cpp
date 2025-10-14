// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#include "etfilters.hpp"
#include <algorithm>
#include <sstream>
#include <iomanip>
#include <cctype>

namespace etcpp {

// FilterCriteria Implementation

void FilterCriteria::setDateRange(std::optional<std::chrono::system_clock::time_point> start,
                                 std::optional<std::chrono::system_clock::time_point> end) {
    mStartDate = start;
    mEndDate = end;
}

void FilterCriteria::addSenderFilter(const std::string& senderPattern, bool isRegex) {
    if (isRegex) {
        try {
            mSenderRegexes.emplace_back(senderPattern, std::regex_constants::icase);
        } catch (const std::regex_error&) {
            // If regex compilation fails, treat as literal pattern
            mSenderPatterns.push_back(senderPattern);
        }
    } else {
        mSenderPatterns.push_back(senderPattern);
    }
}

void FilterCriteria::addRecipientFilter(const std::string& recipientPattern, bool isRegex) {
    if (isRegex) {
        try {
            mRecipientRegexes.emplace_back(recipientPattern, std::regex_constants::icase);
        } catch (const std::regex_error&) {
            // If regex compilation fails, treat as literal pattern
            mRecipientPatterns.push_back(recipientPattern);
        }
    } else {
        mRecipientPatterns.push_back(recipientPattern);
    }
}

void FilterCriteria::addSubjectFilter(const std::string& subjectPattern, bool isRegex) {
    if (isRegex) {
        try {
            mSubjectRegexes.emplace_back(subjectPattern, std::regex_constants::icase);
        } catch (const std::regex_error&) {
            // If regex compilation fails, treat as literal pattern
            mSubjectPatterns.push_back(subjectPattern);
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
        } catch (const std::regex_error&) {
            // If regex compilation fails, treat as literal pattern
            mAttachmentNamePatterns.push_back(namePattern);
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

    // Sender check
    if (!mSenderPatterns.empty() || !mSenderRegexes.empty()) {
        if (!matchesPattern(message.sender, mSenderPatterns, mSenderRegexes)) {
            return false;
        }
    }

    // Recipient check
    if (!mRecipientPatterns.empty() || !mRecipientRegexes.empty()) {
        bool recipientMatch = false;
        for (const auto& recipient : message.recipients) {
            if (matchesPattern(recipient, mRecipientPatterns, mRecipientRegexes)) {
                recipientMatch = true;
                break;
            }
        }
        if (!recipientMatch) {
            return false;
        }
    }

    // Subject check
    if (!mSubjectPatterns.empty() || !mSubjectRegexes.empty()) {
        if (!matchesPattern(message.subject, mSubjectPatterns, mSubjectRegexes)) {
            return false;
        }
    }

    // Attachment presence check
    if (mHasAttachments && message.hasAttachments != *mHasAttachments) {
        return false;
    }

    // Attachment name check
    if (!mAttachmentNamePatterns.empty() || !mAttachmentNameRegexes.empty()) {
        bool attachmentMatch = false;
        for (const auto& attachmentName : message.attachmentNames) {
            if (matchesPattern(attachmentName, mAttachmentNamePatterns, mAttachmentNameRegexes)) {
                attachmentMatch = true;
                break;
            }
        }
        if (!attachmentMatch) {
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

    // Folder check
    if (!mFolders.empty()) {
        bool folderMatch = false;
        for (const auto& folder : mFolders) {
            if (message.folderName == folder) {
                folderMatch = true;
                break;
            }
        }
        if (!folderMatch) {
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
    // Check literal patterns (case-insensitive substring match)
    for (const auto& pattern : patterns) {
        std::string lowerText = text;
        std::string lowerPattern = pattern;
        std::transform(lowerText.begin(), lowerText.end(), lowerText.begin(), ::tolower);
        std::transform(lowerPattern.begin(), lowerPattern.end(), lowerPattern.begin(), ::tolower);
        
        if (lowerText.find(lowerPattern) != std::string::npos) {
            return true;
        }
    }

    // Check regex patterns
    for (const auto& regex : regexes) {
        if (std::regex_search(text, regex)) {
            return true;
        }
    }

    return patterns.empty() && regexes.empty();
}

// FilterBuilder Implementation

FilterBuilder& FilterBuilder::dateRange(const std::string& start, const std::string& end) {
    std::optional<std::chrono::system_clock::time_point> startTime;
    std::optional<std::chrono::system_clock::time_point> endTime;

    if (!start.empty()) {
        startTime = parseDate(start);
    }
    if (!end.empty()) {
        endTime = parseDate(end);
    }

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
    
    // Try YYYY-MM-DD format first
    ss >> std::get_time(&tm, "%Y-%m-%d");
    if (!ss.fail()) {
        return std::chrono::system_clock::from_time_t(std::mktime(&tm));
    }

    // Try YYYY/MM/DD format
    ss.clear();
    ss.str(dateStr);
    ss >> std::get_time(&tm, "%Y/%m/%d");
    if (!ss.fail()) {
        return std::chrono::system_clock::from_time_t(std::mktime(&tm));
    }

    // Try DD-MM-YYYY format
    ss.clear();
    ss.str(dateStr);
    ss >> std::get_time(&tm, "%d-%m-%Y");
    if (!ss.fail()) {
        return std::chrono::system_clock::from_time_t(std::mktime(&tm));
    }

    // If all parsing fails, throw an exception
    throw std::invalid_argument("Invalid date format: " + dateStr + ". Expected formats: YYYY-MM-DD, YYYY/MM/DD, or DD-MM-YYYY");
}

FilterCriteria FilterBuilder::fromCLIArgs(const std::vector<std::string>& args) {
    FilterBuilder builder;
    
    for (size_t i = 0; i < args.size(); ++i) {
        const auto& arg = args[i];
        
        if (arg == "--date-start" && i + 1 < args.size()) {
            builder.dateRange(args[++i], "");
        } else if (arg == "--date-end" && i + 1 < args.size()) {
            builder.dateRange("", args[++i]);
        } else if (arg == "--sender" && i + 1 < args.size()) {
            builder.sender(args[++i]);
        } else if (arg == "--recipient" && i + 1 < args.size()) {
            builder.recipient(args[++i]);
        } else if (arg == "--subject" && i + 1 < args.size()) {
            builder.subject(args[++i]);
        } else if (arg == "--has-attachments") {
            builder.hasAttachments(true);
        } else if (arg == "--no-attachments") {
            builder.hasAttachments(false);
        } else if (arg == "--folder" && i + 1 < args.size()) {
            builder.folder(args[++i]);
        }
    }
    
    return builder.build();
}

FilterCriteria FilterBuilder::fromConfigFile(const std::filesystem::path& configPath) {
    // TODO: Implement config file parsing
    // For now, return empty criteria
    return FilterBuilder().build();
}

} // namespace etcpp