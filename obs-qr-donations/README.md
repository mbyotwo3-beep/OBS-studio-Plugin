# OBS QR Donations Plugin

A professional OBS Studio plugin that enables streamers to receive cryptocurrency donations via QR codes, featuring built-in Lightning Network support via Breez SDK (Greenlight).

## Features

- ⚡ **Lightning Network**: Instant, low-fee Bitcoin donations.
- 🔗 **On-Chain Support**: Bitcoin (BTC) and Liquid (L-BTC) support with BIP21 URI formatting.
- 💰 **Wallet Manager**: Real-time balance, node info, and payment history.
- 💸 **Send Payments**: Send Lightning payments directly from OBS.
- 🔒 **Secure**: Securely seeded wallet generation and strict SSL validation.
- 🎭 **Streamer-Focused**: Auto-copy payment details on stream start and non-blocking notifications.
- 📁 **Robust Backup**: Built-in one-click wallet backup functionality.

## Requirements

- **OBS Studio 28.0+**
- **Qt 6.x**
- **qrencode** library
- **Breez SDK** (included in `third_party/`)

## Installation

### Linux (Ubuntu/Debian)
1. Install dependencies:
   ```bash
   sudo apt-get update
   sudo apt-get install -y build-essential cmake qt6-base-dev libobs-dev libqrencode-dev
   ```
2. Build and install:
   ```bash
   mkdir build && cd build
   cmake .. -DBREEZ_API_KEY="your_optional_api_key"
   make -j$(nproc)
   sudo make install
   ```

### Windows
1. Install OBS Studio, CMake, Qt 6, and Visual Studio.
2. Use `vcpkg` to install `qrencode:x64-windows`.
3. Configure with CMake and build the solution in Visual Studio.

## Getting Started

1. **Add Source**: In OBS, add a new "QR Donations" source.
2. **Enable Lightning**: Open source properties and check "Enable Lightning".
   - *Note: A unique wallet is created automatically on first run.*
3. **Configure API Key**: Enter your Breez API key (or use the default if provided during build).
4. **Test Connection**: Click "Test Breez Connection" to verify setup.

## How to Use

### Receiving Donations
- The QR code automatically rotates between enabled payment methods (Lightning, Bitcoin, Liquid).
- When a donation is received, a non-blocking flash notification appears on the widget.
- Payment details are automatically copied to your clipboard when you start streaming.

### Managing Your Wallet
- Click **⚙️ Manage Lightning Wallet** in the source properties.
- **View Balance:** See your Lightning and On-chain balances.
- **History:** Review all incoming and outgoing transactions.
- **Send Funds:** Use the "Send Payment" feature to pay Lightning invoices.
- **Backup:** Click **Backup Wallet** to save your `seed.dat` to a safe location.

## 🔐 Security & Backup

**CRITICAL:** Your wallet is stored in a `seed.dat` file. If you lose this file, you lose your funds.
- **Location:** `~/.config/obs-studio/plugin_config/obs-qr-donations/seed.dat`
- **Action:** Use the **Backup Wallet** button in the Wallet Manager immediately after setup.
- **Storage:** Store your backup in multiple secure locations (USB, encrypted cloud).

## Support & Contributing

- **Issues:** [Open an issue on GitHub](https://github.com/yourusername/obs-qr-donations/issues)
- **License:** MIT License

---
*This is an unofficial plugin and is not affiliated with the OBS Project or Breez.*
