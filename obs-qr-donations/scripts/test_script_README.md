# OBS QR Donations Test Script

This script provides a graphical interface to test QR code generation for the OBS QR Donations plugin.

## Features

- Generate QR codes with custom amounts and currencies (BTC, USD, EUR, SAT, mSAT)
- Add custom memos/notes to payments
- Save generated QR codes as PNG files
- Simple and intuitive interface

## Requirements

- Python 3.7+
- PyQt6
- qrcode
- Pillow (for image handling)

## Installation

1. Navigate to the `scripts` directory:
   ```bash
   cd path/to/obs-qr-donations/scripts
   ```

2. Install the required packages:
   ```bash
   pip install -r requirements.txt
   ```

## Usage

Run the test script:
```bash
python test_plugin.py
```

### How to Use

1. **Enter Payment Details**:
   - Enter the amount in the amount field
   - Select the currency from the dropdown
   - (Optional) Add a memo/note

2. **Generate QR Code**:
   - Click the "Generate QR Code" button
   - The QR code will appear in the display area

3. **Save QR Code**:
   - Click "Save QR Code" to save the generated QR code
   - Choose a location and filename (default: qr_donation.png)

## Notes

- The QR code contains payment information in JSON format
- The Lightning Network tab is currently a placeholder for future development
- For best results, use the generated QR codes in well-lit conditions

## Troubleshooting

- If the QR code doesn't generate, check the error message and ensure all required packages are installed
- Make sure you have write permissions in the directory where you're trying to save the QR code
- If the interface appears too small, you can resize the window as needed
