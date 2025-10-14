// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#pragma once

#include <cxxopts.hpp>
#include <etfilters.hpp>
#include <etencryption.hpp>
#include <etplugins.hpp>
#include <optional>

namespace etcli {

struct EnhancedOptions {
    // Basic options
    std::string operation;
    std::filesystem::path directory;
    std::string username;
    std::string password;
    std::string totpCode;
    
    // Encryption options
    bool encrypt = false;
    std::string encryptionPassword;
    
    // Incremental backup
    bool incremental = false;
    
    // Export format
    std::string exportFormat = "eml";
    std::string exporterPlugin;
    
    // Filtering options
    std::optional<std::string> dateStart;
    std::optional<std::string> dateEnd;
    std::vector<std::string> senderFilters;
    std::vector<std::string> recipientFilters;
    std::vector<std::string> subjectFilters;
    std::optional<bool> hasAttachments;
    std::vector<std::string> attachmentFilters;
    std::optional<uint64_t> minSize;
    std::optional<uint64_t> maxSize;
    std::vector<std::string> folderFilters;
    
    // Progress options
    bool useEnhancedUI = false;
    std::string progressStyle = "default";
};

class EnhancedCLIParser {
public:
    static EnhancedOptions parseArguments(int argc, const char** argv);
    static void printHelp();
    static void printAvailableFormats();
    static void printFilterExamples();
    
private:
    static cxxopts::Options createOptions();
    static etcpp::FilterCriteria buildFilterCriteria(const EnhancedOptions& options);
};

class ConfigFileManager {
public:
    static void saveConfig(const EnhancedOptions& options, const std::filesystem::path& configPath);
    static EnhancedOptions loadConfig(const std::filesystem::path& configPath);
    static std::vector<EnhancedOptions> loadPresets(const std::filesystem::path& presetsPath);
    
private:
    static std::string serializeToJSON(const EnhancedOptions& options);
    static EnhancedOptions deserializeFromJSON(const std::string& json);
};

// Usage examples that will be shown in help
const char* FILTER_EXAMPLES = R"(
Filter Examples:
  
  Export emails from specific date range:
    --date-start "2023-01-01" --date-end "2023-12-31"
  
  Export emails from specific senders:
    --sender "john@example.com" --sender "*.company.com"
  
  Export only emails with attachments:
    --has-attachments
  
  Export emails larger than 1MB:
    --min-size 1048576
  
  Export from specific folders:
    --folder "INBOX" --folder "Sent"
  
  Complex filter (emails with attachments from 2023):
    --date-start "2023-01-01" --has-attachments --sender "*.business.com"
  
  Export with encryption:
    --encrypt --encryption-password "your-strong-password"
  
  Incremental backup:
    --incremental
  
  Export to PDF format:
    --format pdf --exporter pdf-plugin
)";

} // namespace etcli
