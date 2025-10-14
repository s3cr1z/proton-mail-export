// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#pragma once

#include <string>
#include <filesystem>
#include <vector>
#include <memory>

namespace etcpp {

class EncryptionKey {
public:
    static constexpr size_t KEY_SIZE = 32; // 256 bits
    static constexpr size_t IV_SIZE = 12;  // 96 bits for GCM
    
    EncryptionKey(const std::string& password, const std::vector<uint8_t>& salt);
    
    const std::vector<uint8_t>& getKey() const { return mKey; }
    const std::vector<uint8_t>& getSalt() const { return mSalt; }
    
private:
    std::vector<uint8_t> mKey;
    std::vector<uint8_t> mSalt;
    
    void deriveKey(const std::string& password, const std::vector<uint8_t>& salt);
};

class FileEncryptor {
public:
    FileEncryptor(const EncryptionKey& key);
    
    void encryptFile(const std::filesystem::path& inputPath, 
                    const std::filesystem::path& outputPath);
    void decryptFile(const std::filesystem::path& inputPath, 
                    const std::filesystem::path& outputPath);
    
private:
    const EncryptionKey& mKey;
    
    std::vector<uint8_t> generateIV();
    void encryptData(const std::vector<uint8_t>& plaintext,
                    const std::vector<uint8_t>& iv,
                    std::vector<uint8_t>& ciphertext,
                    std::vector<uint8_t>& tag);
    void decryptData(const std::vector<uint8_t>& ciphertext,
                    const std::vector<uint8_t>& iv,
                    const std::vector<uint8_t>& tag,
                    std::vector<uint8_t>& plaintext);
};

struct BackupManifest {
    bool encrypted = false;
    std::string encryptionMethod;
    std::vector<uint8_t> salt;
    uint64_t timestamp;
    std::string version;
    
    void writeToFile(const std::filesystem::path& path) const;
    static BackupManifest readFromFile(const std::filesystem::path& path);
};

} // namespace etcpp
