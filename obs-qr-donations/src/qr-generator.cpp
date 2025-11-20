#include "qr-generator.hpp"
#include <qrencode.h>
#include <QPainter>
#include <QColor>
#include <QDebug>

// Helper function to convert error correction level to QR_ECLEVEL
static QRcode *generateQRCodeData(const std::string &text, QRGenerator::ErrorCorrectionLevel level) {
    // Convert error correction level
    QRecLevel ecLevel;
    switch (level) {
        case QRGenerator::ErrorCorrectionLevel::Low:      ecLevel = QR_ECLEVEL_L; break;
        case QRGenerator::ErrorCorrectionLevel::Medium:   ecLevel = QR_ECLEVEL_M; break;
        case QRGenerator::ErrorCorrectionLevel::Quartile: ecLevel = QR_ECLEVEL_Q; break;
        case QRGenerator::ErrorCorrectionLevel::High:     ecLevel = QR_ECLEVEL_H; break;
        default:                                          ecLevel = QR_ECLEVEL_M;
    }
    
    // Generate QR code
    QRcode *qr = QRcode_encodeString(
        text.c_str(), // Text to encode
        0,            // Version (0 = auto)
        ecLevel,      // Error correction level
        QR_MODE_8,    // 8-bit data mode
        1             // Case sensitive
    );
    
    return qr;
}

QImage QRGenerator::generateQRCode(const std::string &text, int width, int height) {
    return generateQRCode(text, width, height, ErrorCorrectionLevel::Medium);
}

QImage QRGenerator::generateQRCode(const std::string &text, int width, int height, 
                                 ErrorCorrectionLevel level) {
    if (text.empty()) {
        return QImage();
    }
    
    // Generate QR code data
    QRcode *qr = generateQRCodeData(text, level);
    if (!qr) {
        qWarning() << "Failed to generate QR code for text:" << QString::fromStdString(text);
        return QImage();
    }
    
    // Calculate scaling factor
    const int qrSize = qr->width > 0 ? qr->width : 1;
    const int scaleWidth = width / qrSize;
    const int scaleHeight = height / qrSize;
    const int scale = qMin(scaleWidth, scaleHeight);
    
    // Create image
    QImage image(qrSize * scale, qrSize * scale, QImage::Format_ARGB32);
    image.fill(Qt::white);
    
    // Draw QR code
    QPainter painter(&image);
    painter.setPen(Qt::NoPen);
    painter.setBrush(Qt::black);
    
    for (int y = 0; y < qrSize; y++) {
        for (int x = 0; x < qrSize; x++) {
            if (qr->data[y * qrSize + x] & 0x01) {
                painter.drawRect(x * scale, y * scale, scale, scale);
            }
        }
    }
    
    // Add quiet zone (margin)
    if (scale > 2) {
        const int margin = scale * 2; // 2 modules margin
        QImage paddedImage(image.width() + margin * 2, image.height() + margin * 2, 
                          QImage::Format_ARGB32);
        paddedImage.fill(Qt::white);
        
        QPainter p(&paddedImage);
        p.drawImage(margin, margin, image);
        p.end();
        
        image = paddedImage;
    }
    
    // Clean up
    QRcode_free(qr);
    
    return image.scaled(width, height, Qt::KeepAspectRatio, Qt::SmoothTransformation);
}
