# OBS QR Donations Plugin

A native OBS Studio plugin that enables streamers to receive cryptocurrency donations via QR codes, with built-in Lightning Network support through Breez SDK.

## ✨ New Features

- ⚡ **Lightning Network Support**: Generate and display Lightning invoices with Breez Spark SDK
- 🎆 **Visual Donation Effects**: Particle animations and notifications when payments are received
- 📱 **Fully Responsive**: Automatically adapts to any screen size or aspect ratio (16:9, 4:3, 21:9, 9:16)
- 🎨 **Customizable Display**: Toggle between Lightning, Bitcoin, and Liquid payment methods
- 💰 **Payment Notifications**: Beautiful color-coded visual effects based on donation amount
- 🚀 **Easy Installation**: One-click installer script for Windows, Linux, and Mac

## 🚀 Quick Install

### For Users (Easy!)

1. **Download** the latest release
2. **Run the installer**:
   - Windows: Double-click `install.bat`
   - Linux/Mac: Run `./install.sh` in terminal
3. **Open OBS** and add "QR Donations" source
4. **Configure** your cryptocurrency addresses
5. **Done!** Start receiving donations

**See [USER_GUIDE.md](USER_GUIDE.md) for detailed usage instructions.**

### For Developers (Build from Source)

See [BUILD_GUIDE.md](BUILD_GUIDE.md) for complete build instructions.

Quick build:
```bash
cmake -B build -S . -DBREEZ_USE_STUB=ON
cmake --build build --config Release
install.bat    # Windows
./install.sh   # Linux/Mac
```

## Features

- ⚡ **Lightning Network Support**: Instant payments with Breez SDK
- 🔗 **Bitcoin On-Chain Support**: Display Bitcoin addresses with proper BIP21 URI formatting
 - 🔗 **Bitcoin On-Chain Support**: Display Bitcoin addresses with proper BIP21 URI formatting
 - 🌊 **Liquid (L-BTC) Support**: Display Liquid on-chain addresses and generated QR codes; Lightning invoices can be created on Liquid using Breez when supported
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
   - Note: Breez SDK is optional. If the Breez SDK is not present the plugin will still build and work with Bitcoin on-chain addresses; Lightning/Spark features require the Breez SDK to be available.

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
       - If building with Breez/Spark support, specify the Breez SDK path:
          -DBREEZ_SDK_PATH=/path/to/third_party/breez_sdk
   ```

4. Build the plugin:
   ```
   cmake --build . --config Release
   ```

5. Install the plugin:
   ```
   cmake --install . --config Release
   ```
   
Try the Breez test after installing:

1. Add the `QR Donations` source in OBS.
2. Open source properties and enable Lightning.
3. Enter your Breez API key and Spark details.
4. Click "Test Breez Connection" and confirm the success message.


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
   - Enable "Use Lightning Network" (requires Breez SDK)
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

Important: Each user should obtain their own Breez API key (free) and enter it in the `Breez API Key` setting. The plugin will not embed or share developer API keys; this keeps user funds and rate limits isolated and secure.

Important: You cannot enable Lightning without a valid Breez API key. The plugin enforces this: the Lightning toggle will automatically turn off if the API key is empty and a status is shown in the source properties.


When Lightning is enabled, fields are added to the source properties for Breez API Key, Spark URL and Spark Access Key.

Test Connection:
- Use the "Test Breez Connection" button in the source properties to validate your Breez API Key and Spark settings. The test shows a success/failure message and sets the "Breez Test Status" field with the latest result.

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

- **Bitcoin**: `1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa`
- **Ethereum**: `0x71C7656EC7ab88b098defB751B7401B5f6d8976F`
- **Litecoin**: `Lg6rHZx5w1Z4LgYzKq3rKq5hJ3Nk5XqN2J`
