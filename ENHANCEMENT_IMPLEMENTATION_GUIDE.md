# Proton Mail Export Tool - Enhancement Implementation Guide

This guide provides step-by-step instructions for implementing the requested enhancements to the Proton Mail Export Tool.

## 📋 Overview of Enhancements

1. **Encrypted Backup Options** - AES-256-GCM encryption for backups
2. **Incremental Backups** - Only backup new/changed emails
3. **Advanced Filtering** - Filter by date, sender, attachments, etc.
4. **Plugin System** - Custom exporters (PDF format)
5. **Enhanced Progress UI** - Beautiful progress bars using Bubbletea

## 🚀 Implementation Steps

### Step 1: Update Dependencies

First, update your `vcpkg.json` to include required dependencies:

```json
{
  "name": "proton-mail-export",
  "version": "1.0.5",
  "dependencies": [
    "catch2",
    "cxxopts",
    "fmt",
    "openssl",
    "nlohmann-json",
    "wkhtmltopdf"
  ]
}
```

Update Go dependencies for Bubbletea:

```bash
cd go-lib
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
```

### Step 2: Update CMakeLists.txt

Add the new libraries and source files:

```cmake
# Add to existing CMakeLists.txt
find_package(OpenSSL REQUIRED)
find_package(nlohmann_json CONFIG REQUIRED)

# Add new source files
set(ETCPP_SOURCES
    ${ETCPP_SOURCES}
    lib/etencryption.cpp
    lib/etincremental.cpp
    lib/etfilters.cpp
    lib/etplugins.cpp
)

# Link additional libraries
target_link_libraries(etcpp PRIVATE OpenSSL::SSL OpenSSL::Crypto nlohmann_json::nlohmann_json)
```

### Step 3: Implement Core Features

#### 3.1 Encryption Implementation

Create `lib/lib/etencryption.cpp`:

```cpp
#include "etencryption.hpp"
#include <openssl/evp.h>
#include <openssl/rand.h>
#include <openssl/kdf.h>
#include <fstream>
#include <stdexcept>

namespace etcpp {

EncryptionKey::EncryptionKey(const std::string& password, const std::vector<uint8_t>& salt) 
    : mSalt(salt) {
    if (mSalt.empty()) {
        mSalt.resize(16);
        if (RAND_bytes(mSalt.data(), 16) != 1) {
            throw std::runtime_error("Failed to generate salt");
        }
    }
    deriveKey(password, mSalt);
}

void EncryptionKey::deriveKey(const std::string& password, const std::vector<uint8_t>& salt) {
    mKey.resize(KEY_SIZE);
    
    if (PKCS5_PBKDF2_HMAC(password.c_str(), password.length(),
                          salt.data(), salt.size(),
                          100000, // iterations
                          EVP_sha256(),
                          KEY_SIZE,
                          mKey.data()) != 1) {
        throw std::runtime_error("Failed to derive encryption key");
    }
}

// Implementation continues...
}
```

#### 3.2 Enhanced CLI Options

Update `cli/bin/main.cpp` to include new options:

```cpp
// Add to existing options
options.add_options()
    // Encryption options
    ("encrypt", "Enable backup encryption", cxxopts::value<bool>()->default_value("false"))
    ("encryption-password", "Password for backup encryption", cxxopts::value<std::string>())
    
    // Incremental backup
    ("incremental", "Enable incremental backup", cxxopts::value<bool>()->default_value("false"))
    
    // Filtering options
    ("date-start", "Start date for filtering (YYYY-MM-DD)", cxxopts::value<std::string>())
    ("date-end", "End date for filtering (YYYY-MM-DD)", cxxopts::value<std::string>())
    ("sender", "Filter by sender email/pattern", cxxopts::value<std::vector<std::string>>())
    ("has-attachments", "Filter emails with attachments", cxxopts::value<bool>())
    ("min-size", "Minimum email size in bytes", cxxopts::value<uint64_t>())
    ("folder", "Filter by folder name", cxxopts::value<std::vector<std::string>>())
    
    // Export format
    ("format", "Export format (eml, pdf)", cxxopts::value<std::string>()->default_value("eml"))
    ("exporter", "Custom exporter plugin", cxxopts::value<std::string>())
    
    // UI options
    ("progress-style", "Progress display style (simple, bubbletea)", cxxopts::value<std::string>()->default_value("simple"));
```

### Step 4: Integration Points

#### 4.1 Modify BackupTask

Update `cli/bin/tasks/backup_task.cpp`:

```cpp
void BackupTask::run() {
    // Initialize encryption if enabled
    std::unique_ptr<etcpp::FileEncryptor> encryptor;
    if (mEncryptionEnabled) {
        auto key = std::make_unique<etcpp::EncryptionKey>(mEncryptionPassword, std::vector<uint8_t>());
        encryptor = std::make_unique<etcpp::FileEncryptor>(*key);
    }
    
    // Initialize incremental state if enabled
    std::unique_ptr<etcpp::IncrementalState> incrementalState;
    if (mIncrementalEnabled) {
        incrementalState = std::make_unique<etcpp::IncrementalState>(mBackup.getExportPath());
        incrementalState->loadState();
    }
    
    // Initialize filter criteria
    etcpp::FilterCriteria filterCriteria = buildFilterCriteria();
    
    // Run backup with new features
    mBackup.startWithOptions(filterCriteria, encryptor.get(), incrementalState.get(), *this);
    
    // Save incremental state
    if (incrementalState) {
        incrementalState->saveState();
    }
}
```

#### 4.2 Progress Integration

Update progress handling in `cli/bin/tui_util.cpp`:

```cpp
#include "enhanced_progress.hpp"

void runTaskWithProgress(const TaskAppState& appState, TaskWithProgress<void>& task) {
    auto progressType = etcli::ProgressDisplayFactory::Type::Simple;
    
    // Check if enhanced UI is available and requested
    if (shouldUseEnhancedUI()) {
        progressType = etcli::ProgressDisplayFactory::Type::BubbleTea;
    }
    
    auto display = etcli::ProgressDisplayFactory::create(progressType);
    etcli::ProgressManager progressManager(std::move(display));
    
    progressManager.startOperation(task.description(), task.getTotalItems());
    
    // Run task with progress updates
    task.run();
    
    progressManager.finishOperation();
}
```

### Step 5: Build and Test

#### 5.1 Build Commands

```bash
# Clean and rebuild
rm -rf build
mkdir build && cd build

# Configure with new dependencies
cmake .. -DCMAKE_BUILD_TYPE=Release

# Build
cmake --build . -- -j$(nproc)
```

#### 5.2 Test New Features

```bash
# Test encrypted backup
./proton-mail-export-cli --operation backup --encrypt --encryption-password "test123" --dir /tmp/encrypted_backup --user test@proton.me

# Test incremental backup
./proton-mail-export-cli --operation backup --incremental --dir /tmp/incremental_backup --user test@proton.me

# Test filtering
./proton-mail-export-cli --operation backup --date-start "2023-01-01" --date-end "2023-12-31" --sender "*.company.com" --has-attachments --dir /tmp/filtered_backup --user test@proton.me

# Test PDF export
./proton-mail-export-cli --operation backup --format pdf --dir /tmp/pdf_backup --user test@proton.me

# Test enhanced UI
./proton-mail-export-cli --operation backup --progress-style bubbletea --dir /tmp/enhanced_backup --user test@proton.me
```

## 🔧 Configuration Files

### Example filter configuration (`~/.proton-export/filters.json`):

```json
{
  "presets": {
    "work_emails": {
      "dateStart": "2023-01-01",
      "senderFilters": ["*.company.com", "boss@work.com"],
      "hasAttachments": true,
      "folders": ["INBOX", "Work"]
    },
    "personal_backup": {
      "dateStart": "2020-01-01",
      "excludeSenders": ["noreply@*", "marketing@*"],
      "minSize": 1024
    }
  }
}
```

### PDF Exporter configuration (`~/.proton-export/pdf-config.json`):

```json
{
  "includeHeaders": true,
  "includeAttachments": false,
  "fontFamily": "Arial",
  "fontSize": 12,
  "enableBookmarks": true,
  "pageSize": "A4",
  "orientation": "portrait"
}
```

## 🎯 Usage Examples

### 1. Complete Encrypted Incremental Backup
```bash
./proton-mail-export-cli \
  --operation backup \
  --user john@proton.me \
  --dir ~/ProtonBackup \
  --encrypt \
  --encryption-password "MySecurePassword123!" \
  --incremental \
  --progress-style bubbletea
```

### 2. Filtered Export (Work Emails Only)
```bash
./proton-mail-export-cli \
  --operation backup \
  --user john@proton.me \
  --dir ~/WorkEmails \
  --date-start "2023-01-01" \
  --sender "*.company.com" \
  --sender "*.work.com" \
  --folder "INBOX" \
  --folder "Work" \
  --has-attachments \
  --format pdf
```

### 3. Large Email Archive (Size-based filtering)
```bash
./proton-mail-export-cli \
  --operation backup \
  --user john@proton.me \
  --dir ~/LargeEmails \
  --min-size 5242880 \
  --date-start "2020-01-01" \
  --progress-style bubbletea
```

## 🔍 Troubleshooting

### Common Issues:

1. **Encryption not working**
   - Ensure OpenSSL is properly installed
   - Check that encryption password is provided
   - Verify write permissions to backup directory

2. **Bubbletea UI not displaying**
   - Ensure terminal supports ANSI colors
   - Check if Go components are built correctly
   - Fall back to simple progress bar

3. **Filtering not working as expected**
   - Verify date format (YYYY-MM-DD)
   - Check regex patterns for sender filters
   - Ensure folder names match exactly

4. **PDF export failing**
   - Install wkhtmltopdf system dependency
   - Check available disk space
   - Verify PDF plugin is loaded correctly

## 🚀 Performance Optimizations

1. **Concurrent Processing**: Implement parallel email processing for large exports
2. **Memory Management**: Stream large emails instead of loading entirely in memory  
3. **Compression**: Add optional compression for encrypted backups
4. **Caching**: Cache email metadata for faster incremental backups

## 📚 Next Steps

1. **Testing**: Create comprehensive test suite for all new features
2. **Documentation**: Update user manual and API documentation
3. **CI/CD**: Update build pipelines to include new dependencies
4. **Packaging**: Update installation packages with new requirements

This implementation provides a solid foundation for all requested features while maintaining backward compatibility with the existing tool.
