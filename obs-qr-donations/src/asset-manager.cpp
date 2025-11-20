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
    
    // Add supported assets
    // Format: id, name, symbol, decimals, qrPrefix, iconPath
    assets = {
        {"BTC", "Bitcoin", "₿", 8, "bitcoin:", "qrc:/icons/btc.png"},
        {"ETH", "Ethereum", "Ξ", 18, "ethereum:", "qrc:/icons/eth.png"},
        {"LTC", "Litecoin", "Ł", 8, "litecoin:", "qrc:/icons/ltc.png"},
        {"XRP", "Ripple", "XRP", 6, "ripple:", "qrc:/icons/xrp.png"},
        {"BCH", "Bitcoin Cash", "BCH", 8, "bitcoincash:", "qrc:/icons/bch.png"},
        {"XLM", "Stellar", "XLM", 7, "stellar:", "qrc:/icons/xlm.png"},
        {"DOGE", "Dogecoin", "Ð", 8, "dogecoin:", "qrc:/icons/doge.png"},
        {"USDT", "Tether", "USDT", 6, "tether:", "qrc:/icons/usdt.png"},
        {"USDC", "USD Coin", "USDC", 6, "ethereum:", "qrc:/icons/usdc.png"},
        {"SOL", "Solana", "◎", 9, "solana:", "qrc:/icons/sol.png"}
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
