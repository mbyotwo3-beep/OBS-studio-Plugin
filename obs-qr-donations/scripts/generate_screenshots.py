#!/usr/bin/env python3
"""
Screenshot Generator for OBS QR Donations Plugin

This script automates the process of taking screenshots of the plugin UI
for documentation purposes.
"""
import os
import sys
import time
import subprocess
from pathlib import Path
from PyQt6.QtWidgets import QApplication, QWidget, QVBoxLayout, QPushButton, QLabel
from PyQt6.QtCore import Qt, QTimer
from PyQt6.QtGui import QPixmap, QGuiApplication

# Configuration
SCREENSHOTS_DIR = Path("docs/screenshots")
SCREENSHOTS_DIR.mkdir(parents=True, exist_ok=True)

class ScreenshotTool(QWidget):    
    def __init__(self):
        super().__init__()
        self.setWindowTitle("OBS QR Donations - Screenshot Tool")
        self.setup_ui()
        
    def setup_ui(self):
        layout = QVBoxLayout()
        
        # Add instructions
        instructions = QLabel(
            "This tool will help capture screenshots for documentation.\n"
            "1. Position the OBS QR Donations window as desired\n"
            "2. Click 'Capture' to take a screenshot\n"
            "3. Enter a name for the screenshot when prompted"
        )
        layout.addWidget(instructions)
        
        # Add capture button
        self.capture_btn = QPushButton("Capture Screenshot")
        self.capture_btn.clicked.connect(self.capture_screenshot)
        layout.addWidget(self.capture_btn)
        
        # Add status label
        self.status_label = QLabel("Ready to capture")
        layout.addWidget(self.status_label)
        
        self.setLayout(layout)
    
    def capture_screenshot(self):
        # Get the active window (should be OBS with the plugin)
        window = QApplication.activeWindow()
        if not window:
            self.status_label.setText("Error: No active window found")
            return
            
        # Get screenshot
        screen = QGuiApplication.primaryScreen()
        pixmap = screen.grabWindow(window.winId())
        
        # Generate filename with timestamp
        timestamp = time.strftime("%Y%m%d-%H%M%S")
        filename = f"qr-donations-{timestamp}.png"
        filepath = SCREENSHOTS_DIR / filename
        
        # Save the screenshot
        if pixmap.save(str(filepath), "PNG"):
            self.status_label.setText(f"Screenshot saved to: {filepath}")
            
            # Open the screenshot in default viewer
            if sys.platform == 'win32':
                os.startfile(filepath)
            elif sys.platform == 'darwin':
                subprocess.run(['open', filepath])
            else:
                subprocess.run(['xdg-open', filepath])
        else:
            self.status_label.setText("Failed to save screenshot")

def main():
    app = QApplication(sys.argv)
    
    # Set application style
    app.setStyle('Fusion')
    
    # Create and show the main window
    window = ScreenshotTool()
    window.show()
    
    sys.exit(app.exec())

if __name__ == "__main__":
    main()
