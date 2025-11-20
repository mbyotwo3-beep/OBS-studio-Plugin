#pragma once

#include <string>
#include <vector>
#include <map>

struct AssetInfo {
    std::string id;          // e.g., "BTC", "ETH"
    std::string name;        // e.g., "Bitcoin", "Ethereum"
    std::string symbol;      // e.g., "₿", "Ξ"
    int decimals;            // Number of decimal places
    std::string qrPrefix;    // URI prefix for QR codes (e.g., "bitcoin:")
    std::string iconPath;    // Path to icon file
    
    // Default constructor
    AssetInfo() : decimals(8) {}
    
    // Constructor with parameters
    AssetInfo(const std::string &id, const std::string &name, 
              const std::string &symbol, int decimals = 8,
              const std::string &qrPrefix = "", 
              const std::string &iconPath = "")
        : id(id), name(name), symbol(symbol), decimals(decimals),
          qrPrefix(qrPrefix), iconPath(iconPath) {}
};

class AssetManager {
public:
    // Get the singleton instance
    static AssetManager& instance();
    
    // Initialize the asset manager
    void initialize();
    
    // Get all supported assets
    const std::vector<AssetInfo>& getSupportedAssets() const;
    
    // Get asset info by ID
    bool getAssetInfo(const std::string &id, AssetInfo &outInfo) const;
    
    // Check if an asset is supported
    bool isAssetSupported(const std::string &id) const;
    
    // Disable copy and move
    AssetManager(const AssetManager&) = delete;
    AssetManager& operator=(const AssetManager&) = delete;
    AssetManager(AssetManager&&) = delete;
    AssetManager& operator=(AssetManager&&) = delete;
    
private:
    AssetManager() = default;
    ~AssetManager() = default;
    
    std::vector<AssetInfo> assets;
    std::map<std::string, size_t> assetIndex;
};
