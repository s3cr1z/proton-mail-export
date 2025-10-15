// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#include "etencryption.hpp"
#include <openssl/evp.h>
#include <openssl/rand.h>
#include <openssl/kdf.h>
#include <fstream>
#include <stdexcept>
#include <nlohmann/json.hpp>

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

FileEncryptor::FileEncryptor(const EncryptionKey& key) : mKey(key) {}

void FileEncryptor::encryptFile(const std::filesystem::path& inputPath, 
                               const std::filesystem::path& outputPath) {
    std::ifstream input(inputPath, std::ios::binary);
    if (!input) {
        throw std::runtime_error("Failed to open input file for encryption");
    }
    
    // Read entire file into memory
    std::vector<uint8_t> plaintext((std::istreambuf_iterator<char>(input)),
                                   std::istreambuf_iterator<char>());
    input.close();
    
    auto iv = generateIV();
    std::vector<uint8_t> ciphertext, tag;
    
    encryptData(plaintext, iv, ciphertext, tag);
    
    std::ofstream output(outputPath, std::ios::binary);
    if (!output) {
        throw std::runtime_error("Failed to open output file for encryption");
    }
    
    // Write IV, tag, then ciphertext
    output.write(reinterpret_cast<const char*>(iv.data()), iv.size());
    output.write(reinterpret_cast<const char*>(tag.data()), tag.size());
    output.write(reinterpret_cast<const char*>(ciphertext.data()), ciphertext.size());
}

void FileEncryptor::decryptFile(const std::filesystem::path& inputPath, 
                               const std::filesystem::path& outputPath) {
    std::ifstream input(inputPath, std::ios::binary);
    if (!input) {
        throw std::runtime_error("Failed to open input file for decryption");
    }
    
    // Read IV
    std::vector<uint8_t> iv(EncryptionKey::IV_SIZE);
    input.read(reinterpret_cast<char*>(iv.data()), iv.size());
    
    // Read tag
    std::vector<uint8_t> tag(16); // GCM tag is 16 bytes
    input.read(reinterpret_cast<char*>(tag.data()), tag.size());
    
    // Read ciphertext
    std::vector<uint8_t> ciphertext((std::istreambuf_iterator<char>(input)),
                                    std::istreambuf_iterator<char>());
    input.close();
    
    std::vector<uint8_t> plaintext;
    decryptData(ciphertext, iv, tag, plaintext);
    
    std::ofstream output(outputPath, std::ios::binary);
    if (!output) {
        throw std::runtime_error("Failed to open output file for decryption");
    }
    
    output.write(reinterpret_cast<const char*>(plaintext.data()), plaintext.size());
}

std::vector<uint8_t> FileEncryptor::generateIV() {
    std::vector<uint8_t> iv(EncryptionKey::IV_SIZE);
    if (RAND_bytes(iv.data(), iv.size()) != 1) {
        throw std::runtime_error("Failed to generate IV");
    }
    return iv;
}

void FileEncryptor::encryptData(const std::vector<uint8_t>& plaintext,
                               const std::vector<uint8_t>& iv,
                               std::vector<uint8_t>& ciphertext,
                               std::vector<uint8_t>& tag) {
    EVP_CIPHER_CTX* ctx = EVP_CIPHER_CTX_new();
    if (!ctx) {
        throw std::runtime_error("Failed to create cipher context");
    }
    
    try {
        if (EVP_EncryptInit_ex(ctx, EVP_aes_256_gcm(), nullptr, nullptr, nullptr) != 1) {
            throw std::runtime_error("Failed to initialize encryption");
        }
        
        if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_SET_IVLEN, iv.size(), nullptr) != 1) {
            throw std::runtime_error("Failed to set IV length");
        }
        
        if (EVP_EncryptInit_ex(ctx, nullptr, nullptr, mKey.getKey().data(), iv.data()) != 1) {
            throw std::runtime_error("Failed to set key and IV");
        }
        
        ciphertext.resize(plaintext.size());
        int len;
        
        if (EVP_EncryptUpdate(ctx, ciphertext.data(), &len, plaintext.data(), plaintext.size()) != 1) {
            throw std::runtime_error("Failed to encrypt data");
        }
        
        int finalLen;
        if (EVP_EncryptFinal_ex(ctx, ciphertext.data() + len, &finalLen) != 1) {
            throw std::runtime_error("Failed to finalize encryption");
        }
        
        ciphertext.resize(len + finalLen);
        
        tag.resize(16);
        if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_GET_TAG, 16, tag.data()) != 1) {
            throw std::runtime_error("Failed to get authentication tag");
        }
        
    } catch (...) {
        EVP_CIPHER_CTX_free(ctx);
        throw;
    }
    
    EVP_CIPHER_CTX_free(ctx);
}

void FileEncryptor::decryptData(const std::vector<uint8_t>& ciphertext,
                               const std::vector<uint8_t>& iv,
                               const std::vector<uint8_t>& tag,
                               std::vector<uint8_t>& plaintext) {
    EVP_CIPHER_CTX* ctx = EVP_CIPHER_CTX_new();
    if (!ctx) {
        throw std::runtime_error("Failed to create cipher context");
    }
    
    try {
        if (EVP_DecryptInit_ex(ctx, EVP_aes_256_gcm(), nullptr, nullptr, nullptr) != 1) {
            throw std::runtime_error("Failed to initialize decryption");
        }
        
        if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_SET_IVLEN, iv.size(), nullptr) != 1) {
            throw std::runtime_error("Failed to set IV length");
        }
        
        if (EVP_DecryptInit_ex(ctx, nullptr, nullptr, mKey.getKey().data(), iv.data()) != 1) {
            throw std::runtime_error("Failed to set key and IV");
        }
        
        plaintext.resize(ciphertext.size());
        int len;
        
        if (EVP_DecryptUpdate(ctx, plaintext.data(), &len, ciphertext.data(), ciphertext.size()) != 1) {
            throw std::runtime_error("Failed to decrypt data");
        }
        
        if (EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_GCM_SET_TAG, tag.size(), const_cast<uint8_t*>(tag.data())) != 1) {
            throw std::runtime_error("Failed to set authentication tag");
        }
        
        int finalLen;
        if (EVP_DecryptFinal_ex(ctx, plaintext.data() + len, &finalLen) != 1) {
            throw std::runtime_error("Failed to finalize decryption - authentication failed");
        }
        
        plaintext.resize(len + finalLen);
        
    } catch (...) {
        EVP_CIPHER_CTX_free(ctx);
        throw;
    }
    
    EVP_CIPHER_CTX_free(ctx);
}

void BackupManifest::writeToFile(const std::filesystem::path& path) const {
    nlohmann::json j;
    j["encrypted"] = encrypted;
    j["encryptionMethod"] = encryptionMethod;
    j["salt"] = salt;
    j["timestamp"] = timestamp;
    j["version"] = version;
    
    std::ofstream file(path);
    if (!file) {
        throw std::runtime_error("Failed to write manifest file");
    }
    
    file << j.dump(4);
}

BackupManifest BackupManifest::readFromFile(const std::filesystem::path& path) {
    std::ifstream file(path);
    if (!file) {
        throw std::runtime_error("Failed to read manifest file");
    }
    
    nlohmann::json j;
    file >> j;
    
    BackupManifest manifest;
    manifest.encrypted = j.value("encrypted", false);
    manifest.encryptionMethod = j.value("encryptionMethod", "");
    manifest.salt = j.value("salt", std::vector<uint8_t>());
    manifest.timestamp = j.value("timestamp", 0ULL);
    manifest.version = j.value("version", "");
    
    return manifest;
}

} // namespace etcpp