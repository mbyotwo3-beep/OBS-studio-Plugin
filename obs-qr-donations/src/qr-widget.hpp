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
    
    // Set wallet address & asset
    void setAddress(const std::string &asset, const std::string &address);
    
    // Set display options
    void setDisplayOptions(bool showBalance, bool showAssetSymbol);
    
    // Generate new invoices for both payment methods
    void generateInvoices();
    
    // Get the current QR code as an image
    QImage getQRCodeImage() const;
    
    // Get the current payment requests
    QString getLightningInvoice() const;
    QString getBitcoinAddress() const;
    QString getLiquidAddress() const;
    
    // Get the current amount in satoshis
    qint64 getAmountSats() const;
    
    // Set the amount for invoices (in satoshis)
    void setAmount(qint64 amountSats);

    // Set per-network addresses
    void setBitcoinAddress(const std::string &address);
    void setLiquidAddress(const std::string &address);

    // Set display of Lightning/Breez service status in the widget
    void setLightningStatus(const QString &status, bool ok = true);
    
public slots:
    // Handle payment received
    void onPaymentReceived(qint64 amountSats, const QString &paymentHash, const QString &memo);
    
    // Handle stream status changes
    void onStreamStatusChanged(bool streaming);
    
    // Open the manage wallet dialog for outgoing sends
    void onManageWalletClicked();
    
protected:
    void resizeEvent(QResizeEvent *event) override;
    void paintEvent(QPaintEvent *event) override;
    
private slots:
    void updateQRCode();
    void onGenerateClicked();
    void onCopyLightningClicked();
    void onCopyBitcoinClicked();
    void onCopyLiquidClicked();
    void onTabChanged(int index);
    
private:
    void updateLayout();
    void showError(const QString &message);
    void setLoading(bool loading);
    
    struct Private {
        std::string currentAsset;
        std::string bitcoinAddress;
        std::string liquidAddress;
        std::string lightningInvoice;
        bool showBalance = true;
        bool showAssetSymbol = true;
        qint64 amountSats = 0;
        QDateTime invoiceExpiry;
        
        // UI Elements
        QTabWidget *tabWidget = nullptr;
        QLabel *lightningQRLabel = nullptr;
        QLabel *bitcoinQRLabel = nullptr;
        QLabel *liquidQRLabel = nullptr;
        QLabel *lightningInvoiceLabel = nullptr;
        QLabel *bitcoinAddressLabel = nullptr;
        QLabel *liquidAddressLabel = nullptr;
        QLabel *assetLabel = nullptr;
        QLabel *balanceLabel = nullptr;
        QLabel *amountHintLabel = nullptr;
        QLabel *currentMethodLabel = nullptr; // shows which method (Liquid/Lightning/Bitcoin) is currently displayed
        QLabel *simulationLabel = nullptr; // shows when running in simulated stub mode
        QPushButton *copyLightningBtn = nullptr;
        QPushButton *copyBitcoinBtn = nullptr;
        QPushButton *copyLiquidBtn = nullptr;
        
        // Loading state
        QLabel *lightningLoadingLabel = nullptr;
        QLabel *lightningStatusLabel = nullptr;
        bool isLoading = false;
        
        QVBoxLayout *mainLayout = nullptr;
        QTimer *rotationTimer = nullptr;
        int rotationIndex = 0;
        
        // Flash effect for donation feedback
        QLabel *flashOverlay = nullptr;
        QTimer *flashTimer = nullptr;
    };
    std::unique_ptr<Private> d;
};
