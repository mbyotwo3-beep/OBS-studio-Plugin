#  Quick Breez Spark Setup Guide

## 3-Minute Setup

### What You Need
1. Breez API Key from https://breez.technology (it's free!)
2. OBS Studio with QR Donations plugin installed

### Setup Steps

1. **In OBS Studio:**
   - Sources → + → **QR Donations**
   - Right-click → **Properties**

2. **Enable Lightning:**
   - ☑ **Enable Lightning Network (Breez SDK - Nodeless)**

3. **Enter API Key:**
   ```
   Breez API Key: [YOUR_API_KEY_HERE]
   ```
   
   **That's it!** Spark URL and Access Key are optional (for advanced custom Spark wallet setups only).

4. **Test Connection:**
   - Click **🔌 Test Lightning Connection**
   - Should show: ✅ "Breez initialized successfully"

5. **Done!**
   - Lightning QR codes will generate automatically
   - Payments detected in real-time
   - Balance updates automatically

## What the Plugin Does

✅ **Generates Lightning invoices** when viewers want to donate  
✅ **Displays QR codes** for easy scanning  
✅ **Monitors payments** automatically  
✅ **Shows balance** in real-time  
✅ **Triggers effects** when donations received  

## Full Features

| Feature | Supported |
|---------|-----------|
| Receive Lightning payments | ✅ |
| Generate custom invoices | ✅ |
| Send Lightning payments | ✅ |
| Send on-chain | ✅ |
| Check balance | ✅ |
| Payment history | ✅ |
| Multi-network (BTC/Liquid) | ✅ |

## Troubleshooting

**Can't enable Lightning?**
- Make sure you entered the Breez API Key first

**Test connection fails?**
- Double-check API key (no extra spaces)
- Verify Spark URL is correct
- Ensure internet connection works

**No payments detected?**
- Check OBS logs for errors
- Verify Lightning is enabled
- Test with small amount first

## Need More Info?

See full documentation: [BREEZ_SPARK_GUIDE.md](../docs/BREEZ_SPARK_GUIDE.md)

---

**That's it!** Your OBS stream can now accept Lightning donations with zero configuration beyond entering your API key.
