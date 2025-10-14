// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.

#pragma once

#include <string>
#include <functional>
#include <chrono>
#include <atomic>
#include <memory>

namespace etcli {

struct ProgressMetrics {
    uint64_t totalItems = 0;
    uint64_t processedItems = 0;
    uint64_t failedItems = 0;
    uint64_t totalBytes = 0;
    uint64_t processedBytes = 0;
    
    std::chrono::steady_clock::time_point startTime;
    std::chrono::steady_clock::time_point lastUpdateTime;
    
    std::string currentOperation;
    std::string currentItem;
    
    double getProgressPercent() const {
        if (totalItems == 0) return 0.0;
        return (static_cast<double>(processedItems) / totalItems) * 100.0;
    }
    
    std::chrono::seconds getElapsedTime() const {
        return std::chrono::duration_cast<std::chrono::seconds>(
            std::chrono::steady_clock::now() - startTime);
    }
    
    std::chrono::seconds getEstimatedTimeRemaining() const {
        if (processedItems == 0) return std::chrono::seconds(0);
        
        auto elapsed = getElapsedTime();
        auto itemsPerSecond = static_cast<double>(processedItems) / elapsed.count();
        auto remainingItems = totalItems - processedItems;
        
        return std::chrono::seconds(static_cast<long>(remainingItems / itemsPerSecond));
    }
    
    double getBytesPerSecond() const {
        auto elapsed = getElapsedTime();
        if (elapsed.count() == 0) return 0.0;
        return static_cast<double>(processedBytes) / elapsed.count();
    }
};

class IProgressDisplay {
public:
    virtual ~IProgressDisplay() = default;
    
    virtual void start() = 0;
    virtual void update(const ProgressMetrics& metrics) = 0;
    virtual void finish(const ProgressMetrics& metrics) = 0;
    virtual void setError(const std::string& error) = 0;
    
    virtual bool shouldCancel() = 0; // Check for user cancellation
};

// Traditional CLI progress bar
class SimpleProgressBar : public IProgressDisplay {
public:
    SimpleProgressBar(int width = 50);
    
    void start() override;
    void update(const ProgressMetrics& metrics) override;
    void finish(const ProgressMetrics& metrics) override;
    void setError(const std::string& error) override;
    bool shouldCancel() override;
    
private:
    int mWidth;
    std::atomic<bool> mCancelled{false};
    
    void printProgressBar(const ProgressMetrics& metrics);
    std::string formatTime(std::chrono::seconds seconds);
    std::string formatBytes(uint64_t bytes);
};

// Enhanced TUI using Go bubbletea (via C++ wrapper)
class BubbleTeaProgress : public IProgressDisplay {
public:
    BubbleTeaProgress();
    ~BubbleTeaProgress();
    
    void start() override;
    void update(const ProgressMetrics& metrics) override;
    void finish(const ProgressMetrics& metrics) override;
    void setError(const std::string& error) override;
    bool shouldCancel() override;
    
    void setTheme(const std::string& theme); // "default", "dark", "light"
    void enableKeyHandling(bool enabled);
    
private:
    struct Impl;
    std::unique_ptr<Impl> mImpl;
};

// Factory for creating progress displays
class ProgressDisplayFactory {
public:
    enum class Type {
        Simple,
        BubbleTea,
        Quiet
    };
    
    static std::unique_ptr<IProgressDisplay> create(Type type);
    static Type fromString(const std::string& typeName);
    static std::vector<std::string> getAvailableTypes();
};

// Progress manager that handles the display and updates
class ProgressManager {
public:
    ProgressManager(std::unique_ptr<IProgressDisplay> display);
    ~ProgressManager();
    
    void startOperation(const std::string& operationName, uint64_t totalItems, uint64_t totalBytes = 0);
    void updateProgress(uint64_t processedItems, uint64_t processedBytes = 0, const std::string& currentItem = "");
    void incrementProgress(uint64_t itemsDelta = 1, uint64_t bytesDelta = 0);
    void setCurrentOperation(const std::string& operation);
    void addFailedItem();
    void finishOperation();
    void setError(const std::string& error);
    
    bool shouldCancel() const;
    const ProgressMetrics& getMetrics() const { return mMetrics; }
    
private:
    std::unique_ptr<IProgressDisplay> mDisplay;
    ProgressMetrics mMetrics;
    std::atomic<bool> mRunning{false};
    
    void updateDisplay();
};

} // namespace etcli

// C interface for Go bubbletea integration
extern "C" {
    struct BubbleTeaHandle;
    
    BubbleTeaHandle* bubbletea_create();
    void bubbletea_destroy(BubbleTeaHandle* handle);
    void bubbletea_start(BubbleTeaHandle* handle);
    void bubbletea_update(BubbleTeaHandle* handle, const char* json_metrics);
    void bubbletea_finish(BubbleTeaHandle* handle);
    int bubbletea_should_cancel(BubbleTeaHandle* handle);
    void bubbletea_set_theme(BubbleTeaHandle* handle, const char* theme);
}
