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
#include <memory>
#include <filesystem>

namespace etcpp {

class EncryptionKey {
public:
    static constexpr size_t KEY_SIZE = 32; // AES-256
    static constexpr size_t SALT_SIZE = 16;

    EncryptionKey(const std::string& password, const std::vector<uint8_t>& salt = {});
    
    const std::vector<uint8_t>& getKey() const { return mKey; }
    const std::vector<uint8_t>& getSalt() const { return mSalt; }

private:
    std::vector<uint8_t> mKey;
    std::vector<uint8_t> mSalt;
    
    void deriveKey(const std::string& password, const std::vector<uint8_t>& salt);
};

class FileEncryptor {
public:
    explicit FileEncryptor(const EncryptionKey& key);
    ~FileEncryptor();

    bool encryptFile(const std::filesystem::path& inputPath, const std::filesystem::path& outputPath);
    bool decryptFile(const std::filesystem::path& inputPath, const std::filesystem::path& outputPath);
    
    std::vector<uint8_t> encrypt(const std::vector<uint8_t>& data);
    std::vector<uint8_t> decrypt(const std::vector<uint8_t>& data);

private:
    class Impl;
    std::unique_ptr<Impl> mImpl;
};

} // namespace etcpp