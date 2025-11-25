
## How It Works

### Receiving Payments

Once configured, the plugin will:

1. **Generate Lightning Invoices** automatically when viewers want to donate
2. **Display QR codes** with the Lightning invoice
3. **Monitor for payments** in real-time
4. **Trigger effects** when payments are received (if enabled)
5. **Update balance** automatically

### Invoice Generation

The plugin creates invoices with:
- **Amount**: Specified by the donor or default amount
- **Memo**: "Donation via OBS Stream" (or custom message)
- **Expiry**: 1 hour (default)

### Payment Notifications

When a payment is received:
- `paymentReceived` signal triggers
- UI updates with new balance
- Visual effects play (if enabled)
- Payment hash and amount logged

## Full Wallet Functionality

### ✅ What the Plugin Can Do

| Feature | Status | Description |
|---------|--------|-------------|
| **Receive Payments** | ✅ Full Support | Generate invoices, receive Lightning payments |
| **Invoice Creation** | ✅ Full Support | Create custom invoices with amount/memo |
| **Balance Query** | ✅ Full Support | Check current wallet balance |
| **Payment History** | ✅ Full Support | View past received payments |
| **Send Payments** | ✅ Full Support | Pay Lightning invoices (bolt11) |
| **On-Chain Send** | ✅ Full Support | Send to Bitcoin/Liquid addresses |
| **Real-time Monitoring** | ✅ Full Support | Automatic payment detection |
| **Multi-Network** | ✅ Full Support | Bitcoin and Liquid Lightning |

### Advanced Features

#### Send Lightning Payment
```cpp
BreezService::instance().sendLightningPayment("lnbc...");
```

#### Send On-Chain
```cpp
BreezService::instance().sendOnChain("bc1...", 10000, "bitcoin");
```

#### Check Balance
```cpp
qint64 balance = BreezService::instance().balance();
```

#### Get Payment History
```cpp
QVariantList history = BreezService::instance().paymentHistory();
```

## Settings Reference

### Plugin Settings (OBS Properties)

| Setting | Type | Required | Description |
|---------|------|----------|-------------|
| `enable_lightning` | Boolean | - | Master switch for Lightning functionality |
| `breez_api_key` | String | ✅ YES | Your Breez API key (only required field!) |
| `spark_url` | String | ❌ No | Custom Spark wallet endpoint (advanced - leave empty for default) |
| `spark_access_key` | String | ❌ No | Custom Spark access key (advanced - leave empty for default) |
| `asset` | Dropdown | - | "BTC" or "L-BTC" (determines network) |

**Important:** For most users, you only need the `breez_api_key`. The Spark URL and Access Key are optional advanced settings for custom Spark wallet configurations.

### Test Connection

The **"Test Breez Connection"** button:
- Validates your credentials
- Initializes the Breez SDK
- Connects to the Spark wallet
- Returns status message

**Success:** "Breez initialized successfully"  
**Failure:** "Invalid credentials or SDK not present"

## Troubleshooting

### "Breez API key required"

**Problem:** Lightning enabled but no API key provided

**Solution:**
```
1. Disable "Enable Lightning"
2. Enter your Breez API Key
3. Enter Spark URL and Access Key
4. Click "Test Breez Connection"
5. If successful, enable "Enable Lightning"
```

### "Breez initialization failed"

**Problem:** Invalid credentials or network issues

**Solutions:**
- Verify API key is correct (no extra spaces)
- Check Spark URL format
- Ensure Spark Access Key is valid
- Test internet connection
- Check OBS log for detailed errors

### Stub Mode vs Real SDK

The plugin has two modes:

**Stub Mode** (for testing)
- Simulates Lightning functionality
- No real API key needed
- For development/testing only
- Build with: `-DBREEZ_USE_STUB=ON`

**Real SDK Mode** (for production)
- Uses actual Breez Spark SDK
- Requires valid API credentials
- Handles real Lightning payments
- Build with: `-DBREEZ_USE_STUB=OFF` (default)

### Checking Current Mode

Look for in OBS logs:
```
[QR Donations] Using Breez stub implementation  // Stub mode
[QR Donations] Breez SDK initialized           // Real SDK mode
```

## Security Best Practices

### API Key Security

1. **Never share** your API key publicly
2. **Don't commit** API keys to version control
3. **Use environment variables** for testing
4. **Rotate keys** periodically

### Wallet Security

1. **Monitor balance** regularly
2. **Set spending limits** in Breez dashboard
3. **Enable notifications** for large payments
4. **Keep backups** of access keys

## Example Configuration

### Minimal Setup (Recommended - API Key Only)
```ini
enable_lightning = true
breez_api_key = "sk_live_abc123xyz..."
asset = "BTC"
```

### Advanced Setup with Custom Spark Wallet (Optional)
```ini
enable_lightning = true
breez_api_key = "sk_live_abc123xyz..."
spark_url = "https://your-custom-spark.server"
spark_access_key = "spark_access_abc123..."
asset = "BTC"
```

### Full Setup with On-Chain Fallback
```ini
enable_lightning = true
breez_api_key = "sk_live_abc123xyz..."
asset = "BTC"
bitcoin_address = "bc1q..."  // Fallback on-chain address
show_balance = true
enable_effects = true
```

## Testing Your Setup

### Step 1: Configure
- Enter all credentials
- Click "Test Breez Connection"
- Verify success message

### Step 2: Generate Invoice
- Request a small donation
- Check QR code displays
- Verify invoice is valid

### Step 3: Test Payment
- Use a Lightning wallet to pay test invoice
- Watch for payment notification
- Verify balance updates

### Step 4: Production
- Start streaming
- Donations work automatically!

## API Integration Code

If you're integrating with the plugin programmatically:

```cpp
// Initialize service
BreezService::instance().initialize(
    "your_api_key",
    "https://spark.breez.technology", 
    "your_access_key",
    "bitcoin"  // or "liquid"
);

// Create invoice
QString invoice = BreezService::instance().createInvoice(
    10000,  // 10,000 sats
    "Stream donation",
    3600    // 1 hour expiry
);

// Listen for payments
QObject::connect(&BreezService::instance(), &BreezService::paymentReceived,
    [](qint64 sats, const QString& hash, const QString& memo) {
        qDebug() << "Received" << sats << "sats!";
    }
);
```

## Support

### Getting Help

- **Plugin Issues**: Check BUILD_GUIDE.md and OBS logs
- **Breez Issues**: Contact Breez support
- **Lightning Issues**: See Breez documentation

### Useful Links

- Breez Website: https://breez.technology/
- Breez SDK Docs: https://github.com/breez/breez-sdk
- Spark Documentation: https://breez.technology/spark
- OBS Plugin Repo: [Your GitHub repo]

## Summary

The OBS QR Donations plugin provides **full Lightning wallet functionality** via Breez Spark:

✅ **No node required** - Just paste your API key  
✅ **Automatic invoicing** - Generate QR codes instantly  
✅ **Real-time payments** - Get notified immediately  
✅ **Full control** - Send, receive, check balance  
✅ **Multi-network** - Bitcoin and Liquid support  

Simply configure your Breez Spark credentials in the OBS source properties and you're ready to accept Lightning donations!
