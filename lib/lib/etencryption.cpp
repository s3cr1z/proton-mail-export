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
        mSalt.resize(SALT_SIZE);
        if (RAND_bytes(mSalt.data(), SALT_SIZE) != 1) {
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

class FileEncryptor::Impl {
public:
    explicit Impl(const EncryptionKey& key) : mKey(key) {}
    
    bool encryptFile(const std::filesystem::path& inputPath, const std::filesystem::path& outputPath) {
        // Implementation would go here - for now just copy file
        try {
            std::filesystem::copy_file(inputPath, outputPath, std::filesystem::copy_options::overwrite_existing);
            return true;
        } catch (const std::exception&) {
            return false;
        }
    }
    
    bool decryptFile(const std::filesystem::path& inputPath, const std::filesystem::path& outputPath) {
        // Implementation would go here - for now just copy file
        try {
            std::filesystem::copy_file(inputPath, outputPath, std::filesystem::copy_options::overwrite_existing);
            return true;
        } catch (const std::exception&) {
            return false;
        }
    }
    
    std::vector<uint8_t> encrypt(const std::vector<uint8_t>& data) {
        // Simplified implementation - in real code would use AES-256-GCM
        return data;
    }
    
    std::vector<uint8_t> decrypt(const std::vector<uint8_t>& data) {
        // Simplified implementation - in real code would use AES-256-GCM
        return data;
    }

private:
    const EncryptionKey& mKey;
};

FileEncryptor::FileEncryptor(const EncryptionKey& key) 
    : mImpl(std::make_unique<Impl>(key)) {}

FileEncryptor::~FileEncryptor() = default;

bool FileEncryptor::encryptFile(const std::filesystem::path& inputPath, const std::filesystem::path& outputPath) {
    return mImpl->encryptFile(inputPath, outputPath);
}

bool FileEncryptor::decryptFile(const std::filesystem::path& inputPath, const std::filesystem::path& outputPath) {
    return mImpl->decryptFile(inputPath, outputPath);
}

std::vector<uint8_t> FileEncryptor::encrypt(const std::vector<uint8_t>& data) {
    return mImpl->encrypt(data);
}

std::vector<uint8_t> FileEncryptor::decrypt(const std::vector<uint8_t>& data) {
    return mImpl->decrypt(data);
}

} // namespace etcpp