# Breez SDK Spark API Integration Guide

## Overview
This document provides guidance for integrating the Breez SDK Spark (nodeless) implementation into the OBS QR Donations plugin.

## Key Differences: Spark vs Greenlight

| Feature | Spark (Current) | Greenlight (Deprecated) |
|---------|-----------------|-------------------------|
| Node Required | No | Yes (cloud-hosted) |
| Channel Management | No | Yes (automated) |
| Architecture | Nodeless Layer 2 | Lightning Network |
| API Namespace | `breez_sdk_spark` | `breez_sdk` |
| Configuration | Spark operators | Greenlight nodes |

## API Changes Required

### 1. Include Headers
```cpp
// OLD (Greenlight)
#include <breez_sdk/breez_sdk.h>

// NEW (Spark)
#include <breez_sdk_spark/breez_sdk_spark.h>
```

### 2. Initialization
```cpp
// OLD (Greenlight)
breez_sdk::Config config;
config.working_dir = workingDir.toStdString();
config.api_key = apiKey.toStdString();
config.network = breez_sdk::Network::BITCOIN;

// NEW (Spark)
breez_sdk_spark::Config config;
config.working_dir = workingDir.toStdString();
config.api_key = apiKey.toStdString();
config.network = breez_sdk_spark::Network::BITCOIN;
// No Greenlight-specific config needed
```

### 3. Invoice Creation
```cpp
// OLD (Greenlight)
breez_sdk::ReceivePaymentRequest req;
req.amount_msat = amountSats * 1000;
req.description = description.toStdString();
auto invoice = sdk->receive_payment(req);

// NEW (Spark)
breez_sdk_spark::ReceivePaymentRequest req;
req.amount_msat = amountSats * 1000;
req.description = description.toStdString();
auto invoice = sdk->receive_payment(req);
// Same API, different namespace
```

### 4. Payment Listening
```cpp
// OLD (Greenlight)
sdk->addEventListener([](const breez_sdk::PaymentReceivedEvent& event) {
    // Handle payment
});

// NEW (Spark)
sdk->addEventListener([](const breez_sdk_spark::PaymentReceivedEvent& event) {
    // Handle payment - same structure
});
```

### 5. Spark-Specific Features

#### Spark Addresses
```cpp
// Generate a Spark address for receiving
breez_sdk_spark::ReceiveOnChainRequest req;
req.network = breez_sdk_spark::Network::BITCOIN;
auto address = sdk->receive_on_chain(req);
QString sparkAddress = QString::fromStdString(address.address);
```

#### BTKN Token Support
```cpp
// Send Spark tokens (BTKN)
breez_sdk_spark::SendTokenRequest req;
req.token_address = tokenAddress.toStdString();
req.amount = amount;
auto result = sdk->send_token(req);
```

## Implementation Checklist

- [ ] Replace all `breez_sdk::` namespace references with `breez_sdk_spark::`
- [ ] Remove Greenlight-specific configuration (node config, channel management)
- [ ] Update error handling for Spark-specific errors
- [ ] Add support for Spark addresses
- [ ] Add support for BTKN tokens (optional)
- [ ] Update payment event handlers
- [ ] Test invoice generation
- [ ] Test payment reception
- [ ] Update UI to show Spark features

## Compilation Flags

The following flags are set when Spark SDK is detected:
- `HAVE_BREEZ_SDK` - Breez SDK is available
- `BREEZ_SDK_SPARK` - Using Spark implementation (not Greenlight)
- `ENABLE_SPARK_WALLET` - Spark wallet features enabled

Use these in conditional compilation:
```cpp
#ifdef BREEZ_SDK_SPARK
    // Spark-specific code
#elif defined(HAVE_BREEZ_SDK)
    // Generic Breez SDK code (Greenlight)
#else
    // Stub implementation
#endif
```

## Testing

1. **Build with Stub** (default):
   ```bash
   cmake -B build -S . -DBREEZ_USE_STUB=ON
   cmake --build build
   ```

2. **Build with Spark SDK**:
   ```bash
   # First, build the Spark SDK
   scripts/setup_spark_sdk.bat  # Windows
   # or
   scripts/setup_spark_sdk.sh   # Linux/Mac
   
   # Then build the plugin
   cmake -B build -S . -DBREEZ_USE_STUB=OFF
   cmake --build build
   ```

3. **Test Invoice Generation**:
   - Open OBS
   - Add QR Donations source
   - Enable Lightning
   - Enter Breez API key
   - Click "Generate Invoice"
   - Verify QR code displays

## Troubleshooting

### SDK Not Found
```
Breez SDK Spark not found at third_party/breez_sdk
```
**Solution**: Run `git clone https://github.com/breez/spark-sdk.git third_party/breez_sdk`

### Build Failed
```
cargo build failed
```
**Solution**: 
1. Install Rust: https://rustup.rs/
2. Run `cargo build --release` in `third_party/breez_sdk`

### Linking Errors
```
undefined reference to breez_sdk_spark::...
```
**Solution**: Ensure the Spark SDK library is built and CMake can find it

## Resources

- Spark SDK Repo: https://github.com/breez/spark-sdk
- Documentation: https://sdk-doc-spark.breez.technology/
- API Reference: https://breez.github.io/spark-sdk/breez_sdk_spark/index.html
