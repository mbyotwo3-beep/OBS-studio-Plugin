# OBS QR Donations Plugin - User Guide

## 🚀 Quick Start (5 Minutes)

### Step 1: Install the Plugin
Simply run the installer:
- **Windows**: Double-click `install.bat`
- **Linux/Mac**: Run `./install.sh`

The installer automatically finds OBS and puts files in the right place!

### Step 2: Add to Your Stream
1. Open OBS Studio
2. In the **Sources** panel, click the **+** button
3. Select **"QR Donations"**
4. Give it a name (e.g., "Donation QR")
5. Click **OK**

### Step 3: Configure Addresses (30 seconds)
In the properties window:
1. Enter your **Bitcoin address** (required)
2. *Optional*: Enter Liquid address
3. *Optional*: Enable Lightning (see Lightning Setup below)
4. Click **OK**

**That's it!** Your QR code is now showing on stream! 🎉

## 📱 Using the Plugin

### Resize Anywhere
- Drag corners to resize
- Works in any size or shape
- QR code stays scannable automatically

### Payment Methods
The plugin rotates between three methods every 10 seconds:
- ⚡ **Lightning** (instant, low fees)
- ₿ **Bitcoin** (on-chain)
- 💧 **Liquid** (fast, confidential)

### When Someone Donates
Automatic celebration effects:
- 🎆 Particle burst animation
- 💬 Pop-up showing amount
- 🎨 Colors based on donation size
- ⏱️ 4-second animation

| 720p | 300 x 450 px | Top-right corner |
| Full screen alert | 1920 x 400 px | Center |

### Positioning
- **Corner overlay**: Unobtrusive, always visible
- **Full screen**: For donation alerts
- **Vertical**: Works great in 9:16 format

## 🛠️ Troubleshooting

### QR Code Not Showing
- ✅ Check that you entered a valid Bitcoin address
- ✅ Make sure source is visible in your scene
- ✅ Try resizing the source

### Lightning Not Working
- ✅ Verify Breez API key is correct
- ✅ Click "Test Connection" button
- ✅ Check OBS log for error messages
- ✅ Make sure Spark wallet is configured

### Donation Effects Not Showing
- ✅ Effects only show when payment is received
- ✅ Test with small payment first
- ✅ Check that source is on top layer

### Plugin Not in OBS
- ✅ Run the installer again
- ✅ Restart OBS completely
- ✅ Check OBS installed correctly

## 💡 Pro Tips

### Going Live Checklist
1. **Generate fresh Lightning invoice** before stream
2. **Test donation** with small amount
3. **Copy addresses** to stream description
4. **Position** QR code prominently

### For Maximum Donations
- 📍 **Place in top-right** corner (most visible)
- 💬 **Mention it verbally** ("QR in corner for donations!")
- 🎯 **Set specific goals** ("100k sats for new mic!")
- 🙏 **Thank donors** on stream when notification appears

### Stream Transitions
The plugin auto-updates when you:
- Switch scenes (keeps same invoice)
- Go live (generates new invoice)
- End stream (clears sensitive data)

## 📊 Understanding the Display

### What Viewers See
- **QR Code**: Scannable with any crypto wallet
- **Current Method**: Lightning / Bitcoin / Liquid
- **Amount Hint**: Suggested donation (if set)
- **Payment Info**: Address or invoice below QR

### Demo Mode Banner
If you see "Demo Mode: Payments are SIMULATED":
- You're running the **stub version** (no real payments)
- For real donations, build with Breez SDK
- Perfect for testing layouts!

## 🎯 Common Use Cases

### 1. Always-On Corner Display
- Size: 300-400 px wide
- Position: Top-right
- Shows all payment methods
- Unobtrusive but visible

### 2. Donation Alert Overlay
- Size: 1920 x 400 px
- Position: Center screen
- Show only during breaks
- Hide/show with scene switcher

### 3. Mobile Streaming
- Size: 400 x 600 px (vertical)
- Works in 9:16 format
- Perfect for mobile viewers

## 🔐 Security & Privacy

### Your Addresses
- Stored **locally** in OBS config
- Never sent to external servers
- You control your own wallets

### Lightning Invoices
- Generated fresh each stream
- Expire after 24 hours
- Direct peer-to-peer payments

### Best Practices
- ✅ Use dedicated donation wallets
- ✅ Don't share private keys
- ✅ Verify addresses before streaming
- ✅ Keep Breez API key private

## 📞 Getting Help

### Check Logs
OBS → Help → Log Files → View Current Log
Look for "QR Donations" or "Breez" entries

### Common Issues
See **Troubleshooting** section above

### Report Bugs
Include:
- OBS version
- Plugin version  
- Error messages from logs
- Steps to reproduce

## 🎉 You're All Set!

Your donation QR is ready to earn! Remember:
- 📍 Position prominently
- 💬 Mention to viewers
- 🙏 Thank donors
- 🎨 Enjoy the particle effects!

**Happy streaming!** 🚀
