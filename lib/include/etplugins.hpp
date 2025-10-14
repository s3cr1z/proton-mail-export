// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#pragma once

#include <string>
#include <filesystem>
#include <memory>
#include <vector>
#include <map>
#include <functional>

namespace etcpp {

struct ExportFormat {
    std::string name;
    std::string extension;
    std::string description;
    std::string mimeType;
};

class ExportContext {
public:
    ExportContext(const std::filesystem::path& outputDir, const ExportFormat& format);
    
    const std::filesystem::path& getOutputDirectory() const { return mOutputDir; }
    const ExportFormat& getFormat() const { return mFormat; }
    
    void setMetadata(const std::string& key, const std::string& value);
    std::string getMetadata(const std::string& key) const;
    
private:
    std::filesystem::path mOutputDir;
    ExportFormat mFormat;
    std::map<std::string, std::string> mMetadata;
};

class IExporter {
public:
    virtual ~IExporter() = default;
    
    virtual std::string getName() const = 0;
    virtual std::string getVersion() const = 0;
    virtual std::vector<ExportFormat> getSupportedFormats() const = 0;
    
    virtual void initialize(const ExportContext& context) = 0;
    virtual void exportMessage(const EmailMessage& message, const ExportContext& context) = 0;
    virtual void finalize(const ExportContext& context) = 0;
    
    virtual std::string getConfigSchema() const { return "{}"; } // JSON schema
    virtual void configure(const std::string& config) {}
};

// PDF Exporter Implementation
class PDFExporter : public IExporter {
public:
    std::string getName() const override { return "PDF Exporter"; }
    std::string getVersion() const override { return "1.0.0"; }
    std::vector<ExportFormat> getSupportedFormats() const override;
    
    void initialize(const ExportContext& context) override;
    void exportMessage(const EmailMessage& message, const ExportContext& context) override;
    void finalize(const ExportContext& context) override;
    
    std::string getConfigSchema() const override;
    void configure(const std::string& config) override;
    
private:
    struct PDFConfig {
        bool includeHeaders = true;
        bool includeAttachments = false;
        std::string fontFamily = "Arial";
        int fontSize = 12;
        bool enableBookmarks = true;
    } mConfig;
    
    void createPDFFromMessage(const EmailMessage& message, const std::filesystem::path& outputPath);
    std::string generateHTMLFromMessage(const EmailMessage& message);
};

class PluginManager {
public:
    static PluginManager& getInstance();
    
    void registerExporter(std::unique_ptr<IExporter> exporter);
    void loadPlugin(const std::filesystem::path& pluginPath);
    void loadPluginsFromDirectory(const std::filesystem::path& pluginDir);
    
    std::vector<std::string> getAvailableExporters() const;
    std::vector<ExportFormat> getSupportedFormats() const;
    
    IExporter* getExporter(const std::string& name) const;
    IExporter* getExporterForFormat(const std::string& format) const;
    
private:
    PluginManager() = default;
    std::map<std::string, std::unique_ptr<IExporter>> mExporters;
    
    void registerBuiltinExporters();
};

// Plugin Interface (C-style for dynamic loading)
extern "C" {
    typedef IExporter* (*CreateExporterFunc)();
    typedef void (*DestroyExporterFunc)(IExporter*);
    typedef const char* (*GetPluginInfoFunc)();
}

#define EXPORT_PLUGIN(ExporterClass) \
    extern "C" { \
        IExporter* createExporter() { return new ExporterClass(); } \
        void destroyExporter(IExporter* exporter) { delete exporter; } \
        const char* getPluginInfo() { return #ExporterClass " Plugin v1.0"; } \
    }

} // namespace etcpp
