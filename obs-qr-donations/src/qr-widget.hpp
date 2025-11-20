#pragma once

#include <QWidget>
#include <QImage>
#include <QString>
#include <memory>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QTabWidget>
#include <QPushButton>

// Forward declarations
class QLabel;
class QVBoxLayout;
class QPushButton;

class QLabel;
class QVBoxLayout;

class QRDonationsWidget : public QWidget {
    Q_OBJECT
    
public:
    explicit QRDonationsWidget(QWidget *parent = nullptr);
    ~QRDonationsWidget() override;
    
    // Set wallet information
    void setWalletInfo(const QString &asset, const QString &nodeUrl, const QString &apiKey);
    
    // Set display options
    void setDisplayOptions(bool showBalance, bool showAssetSymbol);
    
    // Generate new invoices for both payment methods
    void generateInvoices();
    
    // Get the current QR code as an image
    QImage getQRCodeImage() const;
    
    // Get the current payment requests
    QString getLightningInvoice() const;
    QString getBitcoinAddress() const;
    
    // Get the current amount in satoshis
    qint64 getAmountSats() const;
    
    // Set the amount for invoices
    void setAmountSats(qint64 amount);
    
public slots:
    // Handle payment received
    void onPaymentReceived(qint64 amountSats, const QString &paymentHash, const QString &memo);
    
    // Handle stream status changes
    void onStreamStatusChanged(bool streaming);
    
protected:
    void resizeEvent(QResizeEvent *event) override;
    void paintEvent(QPaintEvent *event) override;
    
private slots:
    void updateQRCode();
    void onGenerateClicked();
    void onCopyLightningClicked();
    void onCopyBitcoinClicked();
    void onTabChanged(int index);
    
private:
    void updateLayout();
    void showError(const QString &message);
    void setLoading(bool loading);
    
    struct Private {
        std::string currentAsset;
        std::string bitcoinAddress;
        std::string lightningInvoice;
        bool showBalance = true;
        bool showAssetSymbol = true;
        qint64 amountSats = 0;
        QDateTime invoiceExpiry;
        
        // UI Elements
        QTabWidget *tabWidget = nullptr;
        QLabel *lightningQRLabel = nullptr;
        QLabel *bitcoinQRLabel = nullptr;
        QLabel *lightningInvoiceLabel = nullptr;
        QLabel *bitcoinAddressLabel = nullptr;
        QLabel *assetLabel = nullptr;
        QLabel *balanceLabel = nullptr;
        QLabel *amountHintLabel = nullptr;
        QPushButton *copyLightningBtn = nullptr;
        QPushButton *copyBitcoinBtn = nullptr;
        
        // Loading state
        QLabel *lightningLoadingLabel = nullptr;
        QLabel *lightningStatusLabel = nullptr;
        bool isLoading = false;
        
        QVBoxLayout *mainLayout = nullptr;
    };
    std::unique_ptr<Private> d;
};
