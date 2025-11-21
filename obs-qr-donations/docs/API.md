# OBS QR Donations Plugin API Documentation

## Table of Contents
- [Overview](#overview)
- [BreezService API](#breezservice-api)
- [QRDonationsWidget API](#qrdonationswidget-api)
- [Event System](#event-system)
- [Error Handling](#error-handling)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Overview

The OBS QR Donations Plugin provides a comprehensive API for integrating cryptocurrency donations into your OBS Studio setup. This document covers the available classes, methods, and events.

## BreezService API

The `BreezService` class handles the Lightning Network integration using the Breez SDK.
Note: Breez SDK support is optional for building the plugin. If not enabled the plugin will still work for on-chain Bitcoin addresses.

### Initialization

```cpp
// Get the singleton instance
BreezService& breez = BreezService::instance();

// Initialize with your credentials
// Optional network argument: "bitcoin" or "liquid" (default "bitcoin")
bool success = breez.initialize(
    "your_api_key",     // Breez API key
    "https://spark:9737", // Spark wallet URL
    "your_access_key"    // Spark access key
    , "bitcoin"
);

Tip: Use the `Test Breez Connection` button in the source properties to validate credentials before enabling Lightning.

Test result:
- Running Test Breez Connection will set the `breez_test_status` field in the source properties. This is persisted in OBS source settings and can be used to display current state or for diagnostics.
```

### Methods

#### `initialize(apiKey, sparkUrl, sparkAccessKey)`
Initialize the Breez service.
- `apiKey`: Your Breez API key
- `sparkUrl`: URL of your Spark wallet (e.g., "https://spark:9737")
- `sparkAccessKey`: Access key for Spark wallet
- Returns: `bool` indicating success

#### `createInvoice(amountSats, description, expirySec)`
Create a new Lightning invoice.
- `amountSats`: Amount in satoshis (0 for amount-less invoices)
- `description`: Payment description
- `expirySec`: Invoice expiry time in seconds
- Returns: `QString` containing the invoice or empty string on error

#### `nodeInfo()`
Get information about the Lightning node.
- Returns: `QString` with node information in JSON format

#### `balance()`
Get the current balance.
- Returns: `qint64` balance in satoshis

## QRDonationsWidget API

The `QRDonationsWidget` class handles the UI and user interactions.

### Methods

#### `setWalletInfo(asset, nodeUrl, apiKey)`
Set wallet information.
- `asset`: Asset symbol (e.g., "BTC", "LTC")
- `nodeUrl`: Node URL (for Lightning)
- `apiKey`: API key (for Lightning)

#### `setDisplayOptions(showBalance, showAssetSymbol)`
Configure display options.
- `showBalance`: Whether to show the balance
- `showAssetSymbol`: Whether to show the asset symbol

#### `setAmount(amountSats)`
Set the default donation amount.
- `amountSats`: Amount in satoshis (0 for any amount)

#### `generateInvoices()`
Generate new invoices for all payment methods.

### Signals

#### `paymentReceived(amountSats, paymentHash, memo)`
Emitted when a payment is received.
- `amountSats`: Amount in satoshis
- `paymentHash`: Payment hash
- `memo`: Payment memo

#### `errorOccurred(message)`
Emitted when an error occurs.
- `message`: Error description

## Event System

The plugin uses Qt's signal/slot mechanism for event handling.

### Example: Handling Payments

```cpp
// Connect to payment received signal
connect(&breez, &BreezService::paymentReceived,
        [](qint64 amount, const QString& hash, const QString& memo) {
    qInfo() << "Received payment:" << amount << "sats";
    qInfo() << "Payment hash:" << hash;
    qInfo() << "Memo:" << memo;
});
```

## Error Handling

All API methods return error codes or empty values on failure. Check the return values and connect to the `errorOccurred` signal for error handling.

## Examples

### Basic Setup

```cpp
// Initialize Breez service
BreezService& breez = BreezService::instance();
if (!breez.initialize("api_key", "https://spark:9737", "access_key")) {
    qCritical() << "Failed to initialize Breez service";
    return;
}

// Create widget
QRDonationsWidget* widget = new QRDonationsWidget();
widget->setWalletInfo("BTC", "https://spark:9737", "api_key");
widget->setAmount(1000); // 1000 sats
widget->show();
```

### Handling Payments

```cpp
// Connect to payment received signal
connect(&breez, &BreezService::paymentReceived,
        [](qint64 amount, const QString& hash, const QString& memo) {
    // Show notification
    QMessageBox::information(
        nullptr,
        "Payment Received",
        QString("Received %1 sats\nMemo: %2").arg(amount).arg(memo)
    );
    
    // Generate new invoice
    BreezService::instance().createInvoice(0, "Donation", 3600);
});
```

## Troubleshooting

### Common Issues

1. **Failed to initialize Breez service**
   - Verify your API key and Spark wallet URL
   - Check network connectivity
   - Ensure Spark wallet is running and accessible

2. **Payments not being detected**
   - Check if the invoice is still valid
   - Verify your node has sufficient inbound liquidity
   - Check the OBS log for error messages

3. **QR code not updating**
   - Ensure the plugin is properly refreshed in OBS
   - Check for error messages in the OBS log

For additional help, please refer to the [GitHub repository](https://github.com/your-repo/obs-qr-donations) or open an issue.
