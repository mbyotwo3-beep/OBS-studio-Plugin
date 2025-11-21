#include "asset-manager.hpp"
#include <QDir>
#include <QFileInfo>
#include <QStandardPaths>

// Static member initialization
AssetManager& AssetManager::instance() {
    static AssetManager instance;
    return instance;
}

void AssetManager::initialize() {
    // Only initialize once
    if (!assets.empty()) {
        return;
    }
    
    // Support Bitcoin and Liquid to start — Lightning invoices are created using Breez
    assets = {
        {"BTC", "Bitcoin", "₿", 8, "bitcoin:", "qrc:/icons/btc.png"},
        {"L-BTC", "Liquid Bitcoin", "ŁBTC", 8, "liquid:", "qrc:/icons/btc.png"}
    };
    
    // Create index for faster lookups
    for (size_t i = 0; i < assets.size(); ++i) {
        assetIndex[assets[i].id] = i;
    }
}

const std::vector<AssetInfo>& AssetManager::getSupportedAssets() const {
    return assets;
}

bool AssetManager::getAssetInfo(const std::string &id, AssetInfo &outInfo) const {
    auto it = assetIndex.find(id);
    if (it != assetIndex.end() && it->second < assets.size()) {
        outInfo = assets[it->second];
        return true;
    }
    return false;
}

bool AssetManager::isAssetSupported(const std::string &id) const {
    return assetIndex.find(id) != assetIndex.end();
}
