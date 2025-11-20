"""OBS QR Donations Plugin Tester

A simple GUI to test the OBS QR Donations plugin functionality.
"""
import sys
import qrcode
import json
from io import BytesIO
from PyQt6.QtWidgets import (QApplication, QMainWindow, QVBoxLayout, QWidget, QHBoxLayout,
                           QLabel, QLineEdit, QPushButton, QTextEdit, QTabWidget, QFileDialog,
                           QSpinBox, QDoubleSpinBox, QComboBox, QMessageBox)
from PyQt6.QtCore import Qt, QSize
from PyQt6.QtGui import QPixmap, QImage

class PluginTester(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("OBS QR Donations Tester")
        self.setup_ui()
    
    def setup_ui(self):
        # Create tabs
        tabs = QTabWidget()
        
        # QR Code Tab
        qr_tab = QWidget()
        qr_layout = QVBoxLayout()
        
        # QR Code Generation Section
        qr_form = QWidget()
        form_layout = QVBoxLayout()
        
        # Payment Details
        details_group = QWidget()
        details_layout = QVBoxLayout()
        details_layout.addWidget(QLabel("Payment Details"))
        
        # Amount Input
        amount_layout = QHBoxLayout()
        amount_layout.addWidget(QLabel("Amount:"))
        self.amount_input = QDoubleSpinBox()
        self.amount_input.setRange(0.01, 1000000)
        self.amount_input.setValue(10.0)
        amount_layout.addWidget(self.amount_input)
        
        # Currency Selection
        self.currency_combo = QComboBox()
        self.currency_combo.addItems(["BTC", "USD", "EUR", "SAT", "mSAT"])
        amount_layout.addWidget(self.currency_combo)
        details_layout.addLayout(amount_layout)
        
        # Memo/Note
        memo_layout = QHBoxLayout()
        memo_layout.addWidget(QLabel("Memo:"))
        self.memo_input = QLineEdit("Donation for your content!")
        memo_layout.addWidget(self.memo_input)
        details_layout.addLayout(memo_layout)
        
        # Generate QR Button
        self.generate_btn = QPushButton("Generate QR Code")
        self.generate_btn.clicked.connect(self.generate_qr_code)
        details_layout.addWidget(self.generate_btn)
        
        details_group.setLayout(details_layout)
        form_layout.addWidget(details_group)
        
        # QR Code Display
        self.qr_label = QLabel()
        self.qr_label.setAlignment(Qt.AlignmentFlag.AlignCenter)
        self.qr_label.setMinimumSize(300, 300)
        self.qr_label.setStyleSheet("background-color: white; border: 1px solid #ccc;")
        form_layout.addWidget(self.qr_label)
        
        # Save QR Button
        self.save_btn = QPushButton("Save QR Code")
        self.save_btn.clicked.connect(self.save_qr_code)
        self.save_btn.setEnabled(False)
        form_layout.addWidget(self.save_btn)
        
        qr_form.setLayout(form_layout)
        qr_layout.addWidget(qr_form)
        qr_tab.setLayout(qr_layout)
        tabs.addTab(qr_tab, "QR Code")
        
        # Lightning Network Tab (Placeholder)
        ln_tab = QWidget()
        ln_layout = QVBoxLayout()
        ln_layout.addWidget(QLabel("Lightning Network Tester (Coming Soon)"))
        ln_tab.setLayout(ln_layout)
        tabs.addTab(ln_tab, "Lightning Network")
        
        # Set central widget
        central = QWidget()
        layout = QVBoxLayout()
        layout.addWidget(tabs)
        central.setLayout(layout)
        self.setCentralWidget(central)

    def generate_qr_code(self):
        """Generate a QR code with the payment details."""
        try:
            # Prepare payment data
            payment_data = {
                'amount': self.amount_input.value(),
                'currency': self.currency_combo.currentText(),
                'memo': self.memo_input.text(),
                'timestamp': 'now'
            }
            
            # Convert to JSON string
            qr_data = json.dumps(payment_data, indent=2)
            
            # Generate QR code
            qr = qrcode.QRCode(
                version=1,
                error_correction=qrcode.constants.ERROR_CORRECT_L,
                box_size=10,
                border=4,
            )
            qr.add_data(qr_data)
            qr.make(fit=True)
            
            # Create an image from the QR Code instance
            qr_img = qr.make_image(fill_color="black", back_color="white")
            
            # Convert to QPixmap for display
            buffer = BytesIO()
            qr_img.save(buffer, format='PNG')
            qimg = QImage()
            qimg.loadFromData(buffer.getvalue())
            pixmap = QPixmap.fromImage(qimg)
            
            # Scale the pixmap to fit the label while maintaining aspect ratio
            self.qr_label.setPixmap(pixmap.scaled(
                self.qr_label.size() - QSize(20, 20), 
                Qt.AspectRatioMode.KeepAspectRatio,
                Qt.TransformationMode.SmoothTransformation
            ))
            
            # Store the QR data for saving
            self.current_qr_data = qr_data
            self.save_btn.setEnabled(True)
            
        except Exception as e:
            QMessageBox.critical(self, "Error", f"Failed to generate QR code: {str(e)}")
    
    def save_qr_code(self):
        """Save the generated QR code to a file."""
        if not hasattr(self, 'current_qr_data'):
            return
            
        file_path, _ = QFileDialog.getSaveFileName(
            self,
            "Save QR Code",
            "qr_donation.png",
            "PNG Images (*.png);;All Files (*)"
        )
        
        if file_path:
            try:
                # Generate the QR code again to save
                qr = qrcode.QRCode(
                    version=1,
                    error_correction=qrcode.constants.ERROR_CORRECT_L,
                    box_size=10,
                    border=4,
                )
                qr.add_data(self.current_qr_data)
                qr.make(fit=True)
                qr_img = qr.make_image(fill_color="black", back_color="white")
                qr_img.save(file_path)
                QMessageBox.information(self, "Success", f"QR code saved to {file_path}")
            except Exception as e:
                QMessageBox.critical(self, "Error", f"Failed to save QR code: {str(e)}")

def main():
    app = QApplication(sys.argv)
    app.setStyle('Fusion')  # Use Fusion style for a more modern look
    window = PluginTester()
    window.resize(600, 700)
    window.show()
    sys.exit(app.exec())

if __name__ == "__main__":
    main()
