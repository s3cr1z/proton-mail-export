// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#include "filter_builder.hpp"
#include <iostream>
#include <climits>

std::optional<etcpp::FilterCriteria> buildFilterCriteria(const cxxopts::ParseResult& argParseResult) {
    etcpp::FilterBuilder builder;
    bool hasFilters = false;

    // Date range filtering
    if (argParseResult.count("date-start") || argParseResult.count("date-end")) {
        std::string startDate = argParseResult.count("date-start") ? argParseResult["date-start"].as<std::string>() : "";
        std::string endDate = argParseResult.count("date-end") ? argParseResult["date-end"].as<std::string>() : "";
        
        try {
            builder.dateRange(startDate, endDate);
            hasFilters = true;
        } catch (const std::exception& e) {
            std::cerr << "Invalid date range: " << e.what() << std::endl;
            return std::nullopt;
        }
    }

    // Sender filtering
    if (argParseResult.count("sender")) {
        auto senders = argParseResult["sender"].as<std::vector<std::string>>();
        for (const auto& sender : senders) {
            builder.sender(sender, false); // Not regex by default
            hasFilters = true;
        }
    }

    // Recipient filtering
    if (argParseResult.count("recipient")) {
        auto recipients = argParseResult["recipient"].as<std::vector<std::string>>();
        for (const auto& recipient : recipients) {
            builder.recipient(recipient, false); // Not regex by default
            hasFilters = true;
        }
    }

    // Domain filtering (convert to sender/recipient patterns)
    if (argParseResult.count("domain")) {
        auto domains = argParseResult["domain"].as<std::vector<std::string>>();
        for (const auto& domain : domains) {
            std::string domainPattern = "*@" + domain;
            builder.sender(domainPattern, false);
            builder.recipient(domainPattern, false);
            hasFilters = true;
        }
    }

    // Folder filtering
    if (argParseResult.count("folder")) {
        auto folders = argParseResult["folder"].as<std::vector<std::string>>();
        for (const auto& folder : folders) {
            builder.folder(folder);
            hasFilters = true;
        }
    }

    // Subject filtering
    if (argParseResult.count("subject")) {
        auto subjects = argParseResult["subject"].as<std::vector<std::string>>();
        for (const auto& subject : subjects) {
            builder.subject(subject, false); // Not regex by default
            hasFilters = true;
        }
    }

    // Attachment filtering
    if (argParseResult.count("has-attachments")) {
        bool hasAttachments = argParseResult["has-attachments"].as<bool>();
        builder.hasAttachments(hasAttachments);
        hasFilters = true;
    }

    // Size filtering
    if (argParseResult.count("min-size") || argParseResult.count("max-size")) {
        uint64_t minSize = argParseResult.count("min-size") ? argParseResult["min-size"].as<uint64_t>() : 0;
        uint64_t maxSize = argParseResult.count("max-size") ? argParseResult["max-size"].as<uint64_t>() : UINT64_MAX;
        
        if (minSize > maxSize) {
            std::cerr << "Minimum size cannot be greater than maximum size" << std::endl;
            return std::nullopt;
        }
        
        builder.sizeRange(minSize, maxSize);
        hasFilters = true;
    }

    // Validate exclude-folder is not used with folder (conflicting options)
    if (argParseResult.count("exclude-folder") && argParseResult.count("folder")) {
        std::cerr << "Cannot use both --folder and --exclude-folder options simultaneously" << std::endl;
        return std::nullopt;
    }

    if (argParseResult.count("exclude-folder")) {
        std::cerr << "Warning: --exclude-folder is not yet implemented. Use --folder to specify folders to include." << std::endl;
    }

    if (!hasFilters) {
        return std::nullopt; // No filters specified
    }

    try {
        return builder.build();
    } catch (const std::exception& e) {
        std::cerr << "Failed to build filter criteria: " << e.what() << std::endl;
        return std::nullopt;
    }
}