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

#include <etbackup.hpp>
#include <etencryption.hpp>
#include <etincremental.hpp>
#include <etfilters.hpp>
#include <filesystem>
#include <memory>

#include "tasks/task.hpp"
#include "tui_util.hpp"

struct BackupOptions {
    bool encryptionEnabled = false;
    std::string encryptionPassword;
    bool incrementalEnabled = false;
    etcpp::FilterCriteria filterCriteria;
    std::string exportFormat = "eml";
    std::string progressStyle = "simple";
};

class BackupTask final : public TaskWithProgress<void>, etcpp::BackupCallback {
private:
    etcpp::Backup mBackup;
    CLIProgressBar mProgressBar;
    BackupOptions mOptions;
    std::unique_ptr<etcpp::FileEncryptor> mEncryptor;
    std::unique_ptr<etcpp::IncrementalState> mIncrementalState;

public:
    BackupTask(etcpp::Session& session, const std::filesystem::path& backupPath);
    BackupTask(etcpp::Session& session, const std::filesystem::path& backupPath, const BackupOptions& options);
    ~BackupTask() override = default;
    BackupTask(const BackupTask&) = delete;
    BackupTask(BackupTask&&) = delete;
    BackupTask& operator=(const BackupTask&) = delete;
    BackupTask& operator=(BackupTask&&) = delete;

    void run() override;

    void cancel() override;

    std::string_view description() const override;

    inline std::filesystem::path getExportPath() const { return mBackup.getExportPath(); }

    inline uint64_t getExpectedDiskUsage() const { return mBackup.getExpectedDiskUsage(); }

private:
    void onProgress(float progress) override;
    void initializeEncryption();
    void initializeIncremental();
};
