# OBS QR Donations Plugin

A native OBS Studio plugin that enables streamers to receive cryptocurrency donations via QR codes, with built-in Lightning Network support through Breez SDK.

## Features

- ⚡ **Lightning Network Support**: Generate and display Lightning invoices with Breez SDK
- 🔗 **Bitcoin On-Chain Support**: Display Bitcoin addresses with proper BIP21 URI formatting
- 🚀 **Multi-Asset Support**: Display QR codes for various cryptocurrencies (Bitcoin, Ethereum, Litecoin, etc.)
- 🎨 **Customizable Display**: Toggle between Lightning and Bitcoin payment methods
- 🖼️ **High-Quality QR Codes**: Clean, scalable QR code generation
- 💰 **Payment Notifications**: Get real-time notifications for received payments
- 🎭 **Streamer-Focused**: Special modes for live streaming (auto-copy on stream start)
- 🔒 **Secure**: No third-party servers - all transactions are peer-to-peer

## Payment Methods

### Lightning Network (via Breez SDK)
- Instant, low-fee Bitcoin payments
- Support for amount-less invoices (sender specifies amount)
- Auto-refreshing invoices
- Built-in payment tracking

### On-Chain Cryptocurrencies
- Bitcoin (BTC)
- Ethereum (ETH)
- Litecoin (LTC)
- Bitcoin Cash (BCH)

## Requirements

- OBS Studio 28.0 or later
- Qt 6.x (with Widgets and Network modules)
- CMake 3.16 or later
- C++17 compatible compiler
- qrencode library
- Breez SDK (included)
- Spark wallet (for Lightning Network)

## Building from Source

### Windows

1. Install the following prerequisites:
   - [OBS Studio](https://obsproject.com/)
   - [CMake](https://cmake.org/download/)
   - [Qt 6.x](https://www.qt.io/download-qt-installer)
   - [Visual Studio 2019 or later](https://visualstudio.microsoft.com/downloads/)
   - [vcpkg](https://vcpkg.io/en/getting-started.html) (for qrencode)

2. Install dependencies using vcpkg:
   ```
   vcpkg install qrencode:x64-windows
   ```

3. Configure the build:
   ```
   mkdir build
   cd build
   cmake .. -DCMAKE_TOOLCHAIN_FILE=[path-to-vcpkg]/scripts/buildsystems/vcpkg.cmake
   ```

4. Build the plugin:
   ```
   cmake --build . --config Release
   ```

5. Install the plugin:
   ```
   cmake --install . --config Release
   ```

### Linux

1. Install dependencies:
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install -y build-essential cmake qt6-base-dev libobs-dev libqrencode-dev
   ```

2. Build the plugin:
   ```bash
   mkdir build
   cd build
   cmake ..
   make -j$(nproc)
   sudo make install
   ```

## Getting Started

### Basic Setup

1. **Add the plugin to OBS Studio**:
   - Click the "+" button in the Sources box
   - Select "QR Donations"
   - Click "OK"

2. **Configure Lightning Network (Recommended)**:
   - Enable "Use Lightning Network"
   - Enter your Breez API key
   - Set your Spark wallet connection details
   - Configure default donation amount (optional)

3. **Configure Bitcoin On-Chain**:
   - Enter your Bitcoin address
   - Choose QR code size and style
   - Set a default amount (optional)

### For Streamers

1. **Before Going Live**:
   - Open the QR Donations settings
   - Click "Generate New Invoice"
   - Payment details will be automatically copied to your clipboard

2. **During Stream**:
   - Viewers can scan the QR code to donate
   - You'll receive notifications for new donations
   - Invoices automatically refresh after payment

3. **After Stream**:
   - All transactions are logged for your records
   - Export donation history if needed

## Advanced Features

### Breez SDK Integration
- Connect to your own Spark wallet
- Customize invoice expiration time
- Set minimum/maximum donation amounts
- Add custom memos to invoices

### OBS Integration
- Add as a browser source or window capture
- Supports chroma key for transparent backgrounds
- Custom CSS styling available
- Scene collection aware settings

## Troubleshooting

### Common Issues
1. **QR Code Not Updating**
   - Ensure the plugin is properly refreshed in OBS
   - Check for error messages in the OBS log

2. **Lightning Payments Not Working**
   - Verify your Breez API key
   - Check your Spark wallet connection
   - Ensure your node has sufficient liquidity

3. **Plugin Crashes**
   - Update to the latest version
   - Check system requirements
   - Contact support with logs

## Support

For issues and feature requests, please visit our [GitHub repository](https://github.com/your-repo/obs-qr-donations).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) before submitting pull requests.

## Donate

If you find this plugin useful, consider supporting development:
- **Lightning**: [Your Lightning Address]
- **Bitcoin**: [Your Bitcoin Address]

---

*This is an unofficial plugin and is not affiliated with OBS Project or Breez.*

### Adding Custom Icons

1. Place your icon in the `resources/icons/` directory
2. Update `resources.qrc` to include your icon
3. Add the asset to `AssetManager::initialize()` in `asset-manager.cpp`

### Styling

The UI can be customized by modifying the stylesheet in `qr-donations.ui`. The plugin uses a dark theme by default to match OBS Studio's interface.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For support, please [open an issue](https://github.com/yourusername/obs-qr-donations/issues) on GitHub.

## Donate

If you find this plugin useful, consider supporting its development:

- **Bitcoin**: bc1qycuvag85k5ndmrjlpufq3mlsr28ylnffv7k92522afrej7lhag8q0czmwp

