// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#include "filter_validation.hpp"
#include <cxxopts.hpp>
#include <iostream>
#include <sstream>
#include <iomanip>
#include <algorithm>
#include <cctype>

namespace etcli {

std::chrono::system_clock::time_point parseDateTime(const std::string& dateStr) {
    std::tm tm = {};
    std::istringstream ss(dateStr);
    
    // Try YYYY-MM-DD HH:MM:SS format first
    ss >> std::get_time(&tm, "%Y-%m-%d %H:%M:%S");
    if (!ss.fail()) {
        return std::chrono::system_clock::from_time_t(std::mktime(&tm));
    }
    
    // Try YYYY-MM-DD format
    ss.clear();
    ss.str(dateStr);
    ss >> std::get_time(&tm, "%Y-%m-%d");
    if (!ss.fail()) {
        return std::chrono::system_clock::from_time_t(std::mktime(&tm));
    }
    
    throw FilterValidationException("Invalid date format: " + dateStr + ". " + getDateFormatHelp());
}

bool isValidDateFormat(const std::string& dateStr) {
    try {
        parseDateTime(dateStr);
        return true;
    } catch (const FilterValidationException&) {
        return false;
    }
}

bool isValidEmailPattern(const std::string& pattern) {
    if (pattern.empty()) return false;
    
    // Allow wildcards
    if (pattern.find('*') != std::string::npos) {
        return isValidWildcardPattern(pattern);
    }
    
    // Basic email validation
    std::regex emailRegex(R"([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})");
    return std::regex_match(pattern, emailRegex);
}

bool isValidDomainPattern(const std::string& domain) {
    if (domain.empty()) return false;
    
    // Allow wildcards
    if (domain.find('*') != std::string::npos) {
        return isValidWildcardPattern(domain);
    }
    
    // Basic domain validation
    std::regex domainRegex(R"([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})");
    return std::regex_match(domain, domainRegex);
}

uint64_t parseSize(const std::string& sizeStr) {
    if (sizeStr.empty()) {
        throw FilterValidationException("Size cannot be empty. " + getSizeFormatHelp());
    }
    
    std::string numPart;
    std::string unitPart;
    
    // Split number and unit
    size_t i = 0;
    while (i < sizeStr.length() && (std::isdigit(sizeStr[i]) || sizeStr[i] == '.')) {
        numPart += sizeStr[i];
        i++;
    }
    
    while (i < sizeStr.length()) {
        unitPart += std::tolower(sizeStr[i]);
        i++;
    }
    
    if (numPart.empty()) {
        throw FilterValidationException("Invalid size format: " + sizeStr + ". " + getSizeFormatHelp());
    }
    
    double value;
    try {
        value = std::stod(numPart);
    } catch (const std::exception&) {
        throw FilterValidationException("Invalid size number: " + numPart + ". " + getSizeFormatHelp());
    }
    
    if (value < 0) {
        throw FilterValidationException("Size cannot be negative: " + sizeStr);
    }
    
    uint64_t multiplier = 1;
    if (unitPart.empty() || unitPart == "b") {
        multiplier = 1;
    } else if (unitPart == "kb") {
        multiplier = 1024;
    } else if (unitPart == "mb") {
        multiplier = 1024 * 1024;
    } else if (unitPart == "gb") {
        multiplier = 1024 * 1024 * 1024;
    } else {
        throw FilterValidationException("Unknown size unit: " + unitPart + ". " + getSizeFormatHelp());
    }
    
    return static_cast<uint64_t>(value * multiplier);
}

bool isValidSizeFormat(const std::string& sizeStr) {
    try {
        parseSize(sizeStr);
        return true;
    } catch (const FilterValidationException&) {
        return false;
    }
}

bool isValidWildcardPattern(const std::string& pattern) {
    // Basic wildcard pattern validation
    // Allow letters, numbers, dots, hyphens, underscores, @ and *
    for (char c : pattern) {
        if (!std::isalnum(c) && c != '.' && c != '-' && c != '_' && c != '@' && c != '*') {
            return false;
        }
    }
    return true;
}

std::string getDateFormatHelp() {
    return "Use format: YYYY-MM-DD or YYYY-MM-DD HH:MM:SS (e.g., 2024-01-01 or 2024-01-01 09:30:00)";
}

std::string getSizeFormatHelp() {
    return "Use format: number + unit (B, KB, MB, GB). Examples: 1024, 5MB, 1.5GB";
}

std::string getPatternFormatHelp() {
    return "Use * for wildcards. Examples: *@domain.com, user@*, *keyword*";
}

FilterCriteria validateAndBuildFilterCriteria(const cxxopts::ParseResult& args) {
    FilterCriteria criteria;
    
    // Handle legacy options first
    handleLegacyOptions(criteria, args);
    
    // Validate and parse date filters
    if (args.count("since")) {
        const auto sinceStr = args["since"].as<std::string>();
        try {
            criteria.sinceDate = parseDateTime(sinceStr);
        } catch (const FilterValidationException& e) {
            throw FilterValidationException("Invalid --since date: " + std::string(e.what()));
        }
    }
    
    if (args.count("until")) {
        const auto untilStr = args["until"].as<std::string>();
        try {
            criteria.untilDate = parseDateTime(untilStr);
        } catch (const FilterValidationException& e) {
            throw FilterValidationException("Invalid --until date: " + std::string(e.what()));
        }
    }
    
    // Validate date range
    if (criteria.sinceDate && criteria.untilDate && 
        criteria.sinceDate.value() >= criteria.untilDate.value()) {
        throw FilterValidationException("--since date must be before --until date");
    }
    
    // Validate and parse address filters
    if (args.count("address")) {
        criteria.addresses = args["address"].as<std::vector<std::string>>();
        for (const auto& addr : criteria.addresses) {
            if (!isValidEmailPattern(addr)) {
                throw FilterValidationException("Invalid email address pattern: " + addr + ". " + getPatternFormatHelp());
            }
        }
    }
    
    // Validate and parse domain filters
    if (args.count("domain")) {
        criteria.domains = args["domain"].as<std::vector<std::string>>();
        for (const auto& domain : criteria.domains) {
            if (!isValidDomainPattern(domain)) {
                throw FilterValidationException("Invalid domain pattern: " + domain + ". " + getPatternFormatHelp());
            }
        }
    }
    
    // Parse folder and label filters (no validation needed, case-sensitive)
    if (args.count("folder")) {
        criteria.folders = args["folder"].as<std::vector<std::string>>();
    }
    
    if (args.count("label")) {
        criteria.labels = args["label"].as<std::vector<std::string>>();
    }
    
    // Parse subject filters
    if (args.count("subject")) {
        criteria.subjects = args["subject"].as<std::vector<std::string>>();
        for (const auto& subject : criteria.subjects) {
            if (!isValidWildcardPattern(subject)) {
                throw FilterValidationException("Invalid subject pattern: " + subject + ". " + getPatternFormatHelp());
            }
        }
    }
    
    // Parse attachment filters
    if (args.count("has-attachments") && args.count("no-attachments")) {
        throw FilterValidationException("Cannot specify both --has-attachments and --no-attachments");
    }
    
    if (args.count("has-attachments")) {
        criteria.hasAttachments = args["has-attachments"].as<bool>();
    } else if (args.count("no-attachments")) {
        criteria.hasAttachments = !args["no-attachments"].as<bool>();
    }
    
    // Validate and parse size filters
    if (args.count("min-size")) {
        const auto minSizeStr = args["min-size"].as<std::string>();
        try {
            criteria.minSize = parseSize(minSizeStr);
        } catch (const FilterValidationException& e) {
            throw FilterValidationException("Invalid --min-size: " + std::string(e.what()));
        }
    }
    
    if (args.count("max-size")) {
        const auto maxSizeStr = args["max-size"].as<std::string>();
        try {
            criteria.maxSize = parseSize(maxSizeStr);
        } catch (const FilterValidationException& e) {
            throw FilterValidationException("Invalid --max-size: " + std::string(e.what()));
        }
    }
    
    // Validate size range
    if (criteria.minSize && criteria.maxSize && 
        criteria.minSize.value() >= criteria.maxSize.value()) {
        throw FilterValidationException("--min-size must be less than --max-size");
    }
    
    return criteria;
}

void handleLegacyOptions(FilterCriteria& criteria, const cxxopts::ParseResult& args) {
    bool hasLegacy = false;
    
    if (args.count("date-start")) {
        criteria.legacyDateStart = args["date-start"].as<std::string>();
        hasLegacy = true;
        std::cerr << "Warning: --date-start is deprecated, use --since instead" << std::endl;
    }
    
    if (args.count("date-end")) {
        criteria.legacyDateEnd = args["date-end"].as<std::string>();
        hasLegacy = true;
        std::cerr << "Warning: --date-end is deprecated, use --until instead" << std::endl;
    }
    
    if (args.count("sender")) {
        criteria.legacySenders = args["sender"].as<std::vector<std::string>>();
        hasLegacy = true;
        std::cerr << "Warning: --sender is deprecated, use --address instead" << std::endl;
    }
    
    criteria.hasLegacyOptions = hasLegacy;
    
    // Convert legacy options to new format
    if (!criteria.legacyDateStart.empty()) {
        try {
            criteria.sinceDate = parseDateTime(criteria.legacyDateStart);
        } catch (const FilterValidationException& e) {
            throw FilterValidationException("Invalid --date-start: " + std::string(e.what()));
        }
    }
    
    if (!criteria.legacyDateEnd.empty()) {
        try {
            criteria.untilDate = parseDateTime(criteria.legacyDateEnd);
        } catch (const FilterValidationException& e) {
            throw FilterValidationException("Invalid --date-end: " + std::string(e.what()));
        }
    }
    
    if (!criteria.legacySenders.empty()) {
        criteria.addresses = criteria.legacySenders;
        for (const auto& sender : criteria.legacySenders) {
            if (!isValidEmailPattern(sender)) {
                throw FilterValidationException("Invalid sender pattern: " + sender + ". " + getPatternFormatHelp());
            }
        }
    }
}

void printFilterExamples() {
    std::cout << "\n" << std::string(80, '=') << std::endl;
    std::cout << "FILTER USAGE EXAMPLES" << std::endl;
    std::cout << std::string(80, '=') << std::endl;
    
    std::cout << "\n1. Basic backup:" << std::endl;
    std::cout << "   proton-mail-export-cli --operation backup --user john@proton.me --dir ~/backup" << std::endl;
    
    std::cout << "\n2. Export emails from last 30 days:" << std::endl;
    std::cout << "   proton-mail-export-cli --operation backup --user john@proton.me --since 2024-01-01 --dir ~/recent" << std::endl;
    
    std::cout << "\n3. Export work emails with attachments:" << std::endl;
    std::cout << "   proton-mail-export-cli --operation backup --user john@proton.me \\" << std::endl;
    std::cout << "     --domain company.com --has-attachments --folder INBOX --dir ~/work" << std::endl;
    
    std::cout << "\n4. Export specific email addresses:" << std::endl;
    std::cout << "   proton-mail-export-cli --operation backup --user john@proton.me \\" << std::endl;
    std::cout << "     --address boss@company.com --address *@important.org --dir ~/important" << std::endl;
    
    std::cout << "\n5. Export large emails only:" << std::endl;
    std::cout << "   proton-mail-export-cli --operation backup --user john@proton.me \\" << std::endl;
    std::cout << "     --min-size 5MB --max-size 50MB --dir ~/large-emails" << std::endl;
    
    std::cout << "\n6. Date range with subject filtering:" << std::endl;
    std::cout << "   proton-mail-export-cli --operation backup --user john@proton.me \\" << std::endl;
    std::cout << "     --since \"2024-01-01 09:00:00\" --until \"2024-12-31 17:00:00\" \\" << std::endl;
    std::cout << "     --subject \"*invoice*\" --subject \"*receipt*\" --dir ~/financial" << std::endl;
    
    std::cout << "\n7. Encrypted incremental backup:" << std::endl;
    std::cout << "   proton-mail-export-cli --operation backup --user john@proton.me \\" << std::endl;
    std::cout << "     --incremental --encrypt --encryption-password \"MySecurePass123!\" --dir ~/secure" << std::endl;
    
    std::cout << "\n8. Export to PDF format:" << std::endl;
    std::cout << "   proton-mail-export-cli --operation backup --user john@proton.me \\" << std::endl;
    std::cout << "     --format pdf --folder \"Important\" --dir ~/pdf-export" << std::endl;
    
    std::cout << "\nFILTER NOTES:" << std::endl;
    std::cout << "• Date formats: YYYY-MM-DD or YYYY-MM-DD HH:MM:SS" << std::endl;
    std::cout << "• Wildcards: Use * for pattern matching (e.g., *@domain.com, *keyword*)" << std::endl;
    std::cout << "• Size formats: Support B, KB, MB, GB (e.g., 1MB, 500KB, 1024)" << std::endl;
    std::cout << "• Multiple values: Repeat options for OR logic (--address addr1 --address addr2)" << std::endl;
    std::cout << "• Case sensitivity: Folder and label names are case-sensitive" << std::endl;
    
    std::cout << "\nPLATFORM-SPECIFIC PATHS:" << std::endl;
#if defined(_WIN32)
    std::cout << "• Windows: --dir \"C:\\Users\\Username\\Documents\\ProtonBackup\"" << std::endl;
    std::cout << "• Windows: --dir \"%USERPROFILE%\\Documents\\ProtonBackup\"" << std::endl;
#elif defined(__APPLE__)
    std::cout << "• macOS: --dir \"~/Documents/ProtonBackup\"" << std::endl;
    std::cout << "• macOS: --dir \"/Users/username/Documents/ProtonBackup\"" << std::endl;
#else
    std::cout << "• Linux: --dir \"~/Documents/ProtonBackup\"" << std::endl;
    std::cout << "• Linux: --dir \"/home/username/Documents/ProtonBackup\"" << std::endl;
#endif
    
    std::cout << "\nFor more information, visit: https://proton.me/support/proton-mail-export-tool" << std::endl;
}

} // namespace etcli