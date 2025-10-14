// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#pragma once

#include <string>
#include <vector>
#include <chrono>
#include <regex>
#include <functional>

namespace etcpp {

struct EmailMessage {
    std::string messageId;
    std::string subject;
    std::string sender;
    std::vector<std::string> recipients;
    std::chrono::system_clock::time_point timestamp;
    bool hasAttachments;
    std::vector<std::string> attachmentNames;
    uint64_t size;
    std::string folderName;
};

class FilterCriteria {
public:
    // Date filtering
    void setDateRange(std::optional<std::chrono::system_clock::time_point> start,
                     std::optional<std::chrono::system_clock::time_point> end);
    
    // Sender filtering
    void addSenderFilter(const std::string& senderPattern, bool isRegex = false);
    void addRecipientFilter(const std::string& recipientPattern, bool isRegex = false);
    
    // Subject filtering
    void addSubjectFilter(const std::string& subjectPattern, bool isRegex = false);
    
    // Attachment filtering
    void setHasAttachments(bool hasAttachments);
    void addAttachmentNameFilter(const std::string& namePattern, bool isRegex = false);
    
    // Size filtering
    void setSizeRange(std::optional<uint64_t> minSize, std::optional<uint64_t> maxSize);
    
    // Folder filtering
    void addFolderFilter(const std::string& folderName);
    
    // Custom filter functions
    void addCustomFilter(std::function<bool(const EmailMessage&)> filter);
    
    bool matches(const EmailMessage& message) const;
    
private:
    std::optional<std::chrono::system_clock::time_point> mStartDate;
    std::optional<std::chrono::system_clock::time_point> mEndDate;
    
    std::vector<std::regex> mSenderRegexes;
    std::vector<std::string> mSenderPatterns;
    
    std::vector<std::regex> mRecipientRegexes;
    std::vector<std::string> mRecipientPatterns;
    
    std::vector<std::regex> mSubjectRegexes;
    std::vector<std::string> mSubjectPatterns;
    
    std::optional<bool> mHasAttachments;
    std::vector<std::regex> mAttachmentNameRegexes;
    std::vector<std::string> mAttachmentNamePatterns;
    
    std::optional<uint64_t> mMinSize;
    std::optional<uint64_t> mMaxSize;
    
    std::vector<std::string> mFolders;
    
    std::vector<std::function<bool(const EmailMessage&)>> mCustomFilters;
    
    bool matchesPattern(const std::string& text, 
                       const std::vector<std::string>& patterns,
                       const std::vector<std::regex>& regexes) const;
};

class FilterBuilder {
public:
    static FilterCriteria fromCLIArgs(const std::vector<std::string>& args);
    static FilterCriteria fromConfigFile(const std::filesystem::path& configPath);
    
    FilterBuilder& dateRange(const std::string& start, const std::string& end);
    FilterBuilder& sender(const std::string& pattern, bool isRegex = false);
    FilterBuilder& recipient(const std::string& pattern, bool isRegex = false);
    FilterBuilder& subject(const std::string& pattern, bool isRegex = false);
    FilterBuilder& hasAttachments(bool value);
    FilterBuilder& attachmentName(const std::string& pattern, bool isRegex = false);
    FilterBuilder& sizeRange(uint64_t minSize, uint64_t maxSize);
    FilterBuilder& folder(const std::string& folderName);
    
    FilterCriteria build() const;
    
private:
    FilterCriteria mCriteria;
    
    std::chrono::system_clock::time_point parseDate(const std::string& dateStr) const;
};

} // namespace etcpp
