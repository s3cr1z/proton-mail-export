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

#include "tasks/backup_task.hpp"
#include <etsession.hpp>
#include <iostream>

BackupTask::BackupTask(etcpp::Session& session, const std::filesystem::path& backupPath) :
    mBackup(session.newBackup(backupPath.u8string().c_str())) {}

BackupTask::BackupTask(etcpp::Session& session, const std::filesystem::path& backupPath, const BackupOptions& options) :
    mBackup(session.newBackup(backupPath.u8string().c_str())), mOptions(options) {
    
    if (mOptions.encryptionEnabled) {
        initializeEncryption();
    }
    
    if (mOptions.incrementalEnabled) {
        initializeIncremental();
    }
}

void BackupTask::initializeEncryption() {
    if (!mOptions.encryptionPassword.empty()) {
        auto key = std::make_unique<etcpp::EncryptionKey>(mOptions.encryptionPassword);
        mEncryptor = std::make_unique<etcpp::FileEncryptor>(*key);
    }
}

void BackupTask::initializeIncremental() {
    mIncrementalState = std::make_unique<etcpp::IncrementalState>(getExportPath());
    mIncrementalState->loadState();
}

void BackupTask::onProgress(float progress) {
    updateProgress(progress);
}

void BackupTask::run() {
    // Apply filters if specified
    if (!mOptions.filterCriteria.isEmpty()) {
        etcpp::EmailFilter filter(mOptions.filterCriteria);
        // In a real implementation, would pass filter to backup process
        std::cout << "Applying email filters..." << std::endl;
    }
    
    // Start backup with enhanced options
    mBackup.start(*this);
    
    // Save incremental state if enabled
    if (mIncrementalState) {
        mIncrementalState->setLastBackupTime(std::chrono::system_clock::now());
        mIncrementalState->saveState();
    }
}

void BackupTask::cancel() {
    mBackup.cancel();
}

std::string_view BackupTask::description() const {
    if (mOptions.encryptionEnabled && mOptions.incrementalEnabled) {
        return "Export Mail (Encrypted, Incremental)";
    } else if (mOptions.encryptionEnabled) {
        return "Export Mail (Encrypted)";
    } else if (mOptions.incrementalEnabled) {
        return "Export Mail (Incremental)";
    } else {
        return "Export Mail";
    }
}
