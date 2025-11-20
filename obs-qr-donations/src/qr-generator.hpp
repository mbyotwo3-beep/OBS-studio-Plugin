#pragma once

#include <QImage>
#include <string>

class QRGenerator {
public:
    // Generate a QR code from the given text with the specified size
    static QImage generateQRCode(const std::string &text, int width, int height);
    
    // Generate a QR code with a specific error correction level
    enum class ErrorCorrectionLevel {
        Low,      // 7% of data can be restored
        Medium,   // 15% of data can be restored
        Quartile, // 25% of data can be restored
        High      // 30% of data can be restored
    };
    
    static QImage generateQRCode(const std::string &text, int width, int height, 
                               ErrorCorrectionLevel level);

private:
    // Private constructor to prevent instantiation
    QRGenerator() = delete;
    ~QRGenerator() = delete;
    
    // Disable copy and move
    QRGenerator(const QRGenerator&) = delete;
    QRGenerator& operator=(const QRGenerator&) = delete;
    QRGenerator(QRGenerator&&) = delete;
    QRGenerator& operator=(QRGenerator&&) = delete;
};
