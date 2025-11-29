#include "qr-widget.hpp"
#include "qr-generator.hpp"
#include "breez-service.hpp"
#include "manage-wallet-dialog.hpp"
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QLabel>
#include <QPainter>
#include <QStyleOption>
#include <QApplication>
#include <QTabWidget>
#include <QPushButton>
#include <QClipboard>
#include <QMessageBox>
#include <QDateTime>

struct QRDonationsWidget::Private {
    // Payment methods
    enum PaymentTab {
        TAB_LIGHTNING = 0,
        TAB_BITCOIN = 1,
        TAB_LIQUID = 2
    };
    
    std::string currentAsset;
    std::string bitcoinAddress;
    std::string liquidAddress;
    std::string lightningInvoice;
    
    bool showBalance = true;
    bool showAssetSymbol = true;
    qint64 amountSats = 0;  // 0 means let sender specify amount
    
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
    QPushButton *copyLightningBtn = nullptr;
    QPushButton *copyBitcoinBtn = nullptr;
    QPushButton *copyLiquidBtn = nullptr;
    
    // Timestamp for invoice expiration
    QDateTime invoiceExpiry;
    
    QVBoxLayout *mainLayout = nullptr;
    
    Private() {
        // Initialize components
    }
    
    ~Private() {
        // Cleanup if needed
    }
};

QRDonationsWidget::QRDonationsWidget(QWidget *parent)
    : QWidget(parent)
    , d(new Private)
{
    setWindowTitle("QR Donations");
    setMinimumSize(400, 500);
    
    // Create tab widget
    d->tabWidget = new QTabWidget(this);
    
    // Create Lightning tab
    QWidget *lightningTab = new QWidget();
    QVBoxLayout *lightningLayout = new QVBoxLayout(lightningTab);
    
    // Create a container for the QR code with loading overlay
    QWidget *qrContainer = new QWidget(lightningTab);
    QVBoxLayout *qrLayout = new QVBoxLayout(qrContainer);
    qrLayout->setContentsMargins(0, 0, 0, 0);
    
    d->lightningQRLabel = new QLabel(qrContainer);
    d->lightningQRLabel->setAlignment(Qt::AlignCenter);
    d->lightningQRLabel->setMinimumSize(250, 250);
    
    // Create a simple loading indicator text overlay
    d->lightningLoadingLabel = new QLabel(qrContainer);
    d->lightningLoadingLabel->setAlignment(Qt::AlignCenter);
    d->lightningLoadingLabel->setStyleSheet("background-color: rgba(255, 255, 255, 200); font-weight: bold;");
    d->lightningLoadingLabel->setText("Generating invoice...");
    d->lightningLoadingLabel->setVisible(false);
    
    qrLayout->addWidget(d->lightningQRLabel);
    qrLayout->addWidget(d->lightningLoadingLabel, 0, Qt::AlignCenter);
    
    d->lightningInvoiceLabel = new QLabel(lightningTab);
    d->lightningInvoiceLabel->setWordWrap(true);
    d->lightningInvoiceLabel->setTextInteractionFlags(Qt::TextSelectableByMouse);
    d->lightningInvoiceLabel->setStyleSheet(
        "QLabel { background-color: #f0f0f0; padding: 8px; border-radius: 4px; }"
        "QLabel:disabled { color: #888; }"
    );
    
    d->copyLightningBtn = new QPushButton("Copy Invoice", lightningTab);
    // Keep copy button simple; no icon required for a cleaner look
    d->copyLightningBtn->setEnabled(false);
    connect(d->copyLightningBtn, &QPushButton::clicked, this, &QRDonationsWidget::onCopyLightningClicked);
    
    // Status label for errors/messages
    d->lightningStatusLabel = new QLabel(lightningTab);
    d->lightningStatusLabel->setWordWrap(true);
    d->lightningStatusLabel->setStyleSheet("color: #d32f2f; font-size: 12px;");
    d->lightningStatusLabel->setVisible(false);
    
    lightningLayout->addWidget(qrContainer, 1);
    lightningLayout->addWidget(new QLabel("<b>Lightning Invoice:</b>", lightningTab));
    lightningLayout->addWidget(d->lightningInvoiceLabel);
    lightningLayout->addWidget(d->lightningStatusLabel);
    lightningLayout->addWidget(d->copyLightningBtn);
    
    // Create Bitcoin tab
    QWidget *bitcoinTab = new QWidget();
    QVBoxLayout *bitcoinLayout = new QVBoxLayout(bitcoinTab);
    
    d->bitcoinQRLabel = new QLabel(bitcoinTab);
    d->bitcoinQRLabel->setAlignment(Qt::AlignCenter);
    d->bitcoinQRLabel->setMinimumSize(250, 250);
    
    d->bitcoinAddressLabel = new QLabel(bitcoinTab);
    d->bitcoinAddressLabel->setWordWrap(true);
    d->bitcoinAddressLabel->setTextInteractionFlags(Qt::TextSelectableByMouse);
    d->bitcoinAddressLabel->setStyleSheet("background-color: #f0f0f0; padding: 8px; border-radius: 4px;");
    
    d->copyBitcoinBtn = new QPushButton("Copy Address", bitcoinTab);
    connect(d->copyBitcoinBtn, &QPushButton::clicked, this, &QRDonationsWidget::onCopyBitcoinClicked);
    
    bitcoinLayout->addWidget(d->bitcoinQRLabel, 1);
    bitcoinLayout->addWidget(new QLabel("Bitcoin Address:", bitcoinTab));
    bitcoinLayout->addWidget(d->bitcoinAddressLabel);
    bitcoinLayout->addWidget(d->copyBitcoinBtn);
    
    // Create Liquid tab (on-chain Liquid network)
    QWidget *liquidTab = new QWidget();
    QVBoxLayout *liquidLayout = new QVBoxLayout(liquidTab);

    d->liquidQRLabel = new QLabel(liquidTab);
    d->liquidQRLabel->setAlignment(Qt::AlignCenter);
    d->liquidQRLabel->setMinimumSize(250, 250);

    d->liquidAddressLabel = new QLabel(liquidTab);
    d->liquidAddressLabel->setWordWrap(true);
    d->liquidAddressLabel->setTextInteractionFlags(Qt::TextSelectableByMouse);
    d->liquidAddressLabel->setStyleSheet("background-color: #f0f0f0; padding: 8px; border-radius: 4px;");

    d->copyLiquidBtn = new QPushButton("Copy Liquid Address", liquidTab);
    connect(d->copyLiquidBtn, &QPushButton::clicked, this, &QRDonationsWidget::onCopyLiquidClicked);

    liquidLayout->addWidget(d->liquidQRLabel, 1);
    liquidLayout->addWidget(new QLabel("Liquid Address:", liquidTab));
    liquidLayout->addWidget(d->liquidAddressLabel);
    liquidLayout->addWidget(d->copyLiquidBtn);
    
    // Add tabs
    // Keep tabs simple and text-only to reduce missing resource issues
    d->tabWidget->addTab(lightningTab, "Lightning");
    d->tabWidget->addTab(bitcoinTab, "Bitcoin");
    d->tabWidget->addTab(liquidTab, "Liquid");
    connect(d->tabWidget, &QTabWidget::currentChanged, this, &QRDonationsWidget::onTabChanged);
    
    // Info labels
    d->assetLabel = new QLabel(this);
    d->assetLabel->setAlignment(Qt::AlignCenter);
    d->assetLabel->setStyleSheet("font-weight: bold; font-size: 16px;");
    
    d->balanceLabel = new QLabel(this);
    d->balanceLabel->setAlignment(Qt::AlignCenter);
    d->balanceLabel->setStyleSheet("color: #4CAF50; font-size: 14px;");
    
    d->amountHintLabel = new QLabel(this);
    d->amountHintLabel->setAlignment(Qt::AlignCenter);
    d->amountHintLabel->setStyleSheet("color: #2196F3; font-style: italic; font-size: 12px;");
    d->amountHintLabel->setWordWrap(true);
    
    // Set up main layout
    QVBoxLayout *mainLayout = new QVBoxLayout(this);
    d->currentMethodLabel = new QLabel(this);
    d->currentMethodLabel->setAlignment(Qt::AlignCenter);
    d->currentMethodLabel->setStyleSheet("font-weight: bold; font-size: 14px; color: #2196F3; padding: 4px;");
    // Simulation banner (visible when the stub is compiled with simulation enabled)
    d->simulationLabel = new QLabel(this);
    d->simulationLabel->setAlignment(Qt::AlignCenter);
    d->simulationLabel->setStyleSheet("background-color: #FFF3CD; color: #856404; padding: 6px; border: 1px solid #FFE8A1; border-radius: 4px; font-weight: bold;");
#ifdef BREEZ_STUB_SIMULATE
    d->simulationLabel->setText("Demo Mode: Payments are SIMULATED — no real funds will be transferred");
    d->simulationLabel->setVisible(true);
#else
    d->simulationLabel->setVisible(false);
#endif
    
    // Set up rotation timer (10s) which cycles Liquid -> Lightning -> Bitcoin
    d->rotationTimer = new QTimer(this);
    d->rotationTimer->setInterval(10000);
    d->rotationTimer->setSingleShot(false);
    d->rotationIndex = 0;
    connect(d->rotationTimer, &QTimer::timeout, this, [this]() {
        // rotation order: Liquid (2) -> Lightning (0) -> Bitcoin (1)
        const QVector<int> order = {2, 0, 1};
        int attempts = 0;
        do {
            d->rotationIndex = (d->rotationIndex + 1) % order.size();
            int idx = order[d->rotationIndex];
            // Only switch to enabled tabs
            if (idx == 0 && !d->lightningInvoice.empty()) { d->tabWidget->setCurrentIndex(0); d->currentMethodLabel->setText("Lightning"); break; }
            if (idx == 1 && !d->bitcoinAddress.empty()) { d->tabWidget->setCurrentIndex(1); d->currentMethodLabel->setText("Bitcoin"); break; }
            if (idx == 2 && !d->liquidAddress.empty()) { d->tabWidget->setCurrentIndex(2); d->currentMethodLabel->setText("Liquid"); break; }
            attempts++;
        } while (attempts < order.size());
    });
    
    mainLayout->addWidget(d->currentMethodLabel);
    mainLayout->addWidget(d->simulationLabel);
    mainLayout->addWidget(d->assetLabel);
    mainLayout->addWidget(d->tabWidget, 1);
    mainLayout->addWidget(d->balanceLabel);
    mainLayout->addWidget(d->amountHintLabel);
    
    // Manage Wallet button for outgoing sends (post-stream)
    QPushButton *manageWalletBtn = new QPushButton("Manage Wallet", this);
    connect(manageWalletBtn, &QPushButton::clicked, this, &QRDonationsWidget::onManageWalletClicked);
    mainLayout->addWidget(manageWalletBtn);
    mainLayout->setSpacing(10);
    mainLayout->setContentsMargins(15, 15, 15, 15);
    
    // Create flash overlay (hidden by default)
    d->flashOverlay = new QLabel(this);
    d->flashOverlay->setAlignment(Qt::AlignCenter);
    d->flashOverlay->setStyleSheet(
        "QLabel { "
        "  background-color: rgba(76, 175, 80, 200); "
        "  color: white; "
        "  font-size: 18px; "
        "  font-weight: bold; "
        "  padding: 20px; "
        "  border-radius: 8px; "
        "}"
    );
    d->flashOverlay->hide();
    d->flashOverlay->raise(); // Keep on top
    
    // Create timer for flash effect
    d->flashTimer = new QTimer(this);
    d->flashTimer->setSingleShot(true);
    connect(d->flashTimer, &QTimer::timeout, this, [this]() {
        if (d->flashOverlay) {
            d->flashOverlay->hide();
            // Reset background flash
            setStyleSheet("");
        }
    });
    
    // Initialize with default values
    setAddress("BTC", "");
    // Start rotation
    d->rotationTimer->start();
}

void QRDonationsWidget::onCopyLiquidClicked() {
    QClipboard *clipboard = QGuiApplication::clipboard();
    if (clipboard && !d->liquidAddress.empty()) {
        clipboard->setText(QString::fromStdString(d->liquidAddress));
        QMessageBox::information(this, "Copied", "Liquid address copied to clipboard");
    } else {
        QMessageBox::warning(this, "Copy", "No Liquid address available to copy");
    }
}

QString QRDonationsWidget::getLiquidAddress() const {
    return QString::fromStdString(d->liquidAddress);
}

QString QRDonationsWidget::getBitcoinAddress() const {
    return QString::fromStdString(d->bitcoinAddress);
}

QString QRDonationsWidget::getLightningInvoice() const {
    return QString::fromStdString(d->lightningInvoice);
}

void QRDonationsWidget::onCopyBitcoinClicked() {
    QClipboard *clipboard = QGuiApplication::clipboard();
    if (clipboard && !d->bitcoinAddress.empty()) {
        clipboard->setText(QString::fromStdString(d->bitcoinAddress));
        QMessageBox::information(this, "Copied", "Address copied to clipboard");
    } else {
        QMessageBox::warning(this, "Copy", "No address available to copy");
    }
}

QRDonationsWidget::~QRDonationsWidget() = default;

void QRDonationsWidget::setAddress(const std::string &asset, const std::string &address) {
    d->currentAsset = asset;
    d->bitcoinAddress = address;
    
    // Update UI
    d->assetLabel->setText(QString::fromStdString(asset));
    d->bitcoinAddressLabel->setText(QString::fromStdString(address));
    
    // Generate new invoices
    generateInvoices();
    updateLayout();
}

void QRDonationsWidget::setBitcoinAddress(const std::string &address) {
    d->bitcoinAddress = address;
    if (d->bitcoinAddressLabel) d->bitcoinAddressLabel->setText(QString::fromStdString(address));
    updateQRCode();
}

void QRDonationsWidget::setLiquidAddress(const std::string &address) {
    d->liquidAddress = address;
    if (d->liquidAddressLabel) d->liquidAddressLabel->setText(QString::fromStdString(address));
    updateQRCode();
}

void QRDonationsWidget::setDisplayOptions(bool showBalance, bool showAssetSymbol) {
    if (d->showBalance != showBalance || d->showAssetSymbol != showAssetSymbol) {
        d->showBalance = showBalance;
        d->showAssetSymbol = showAssetSymbol;
        updateLayout();
    }
}

void QRDonationsWidget::setAmount(qint64 amountSats) {
    if (d->amountSats != amountSats) {
        d->amountSats = amountSats;
        updateQRCode();
    }
}

qint64 QRDonationsWidget::getAmountSats() const {
    return d->amountSats;
}

QImage QRDonationsWidget::getQRCodeImage() const {
    return d->qrCode;
}

void QRDonationsWidget::resizeEvent(QResizeEvent *event) {
    QWidget::resizeEvent(event);
    updateQRCode();
}

void QRDonationsWidget::paintEvent(QPaintEvent *event) {
    QStyleOption opt;
    opt.initFrom(this);
    QPainter p(this);
    style()->drawPrimitive(QStyle::PE_Widget, &opt, &p, this);
}

void QRDonationsWidget::generateInvoices() {
    if (d->currentAsset.empty()) {
        return;
    }
    
    // Show loading state
    setLoading(true);
    
    // Clear previous invoice and status
    d->lightningInvoice.clear();
    d->lightningInvoiceLabel->clear();
    d->lightningStatusLabel->clear();
    d->lightningStatusLabel->setVisible(false);
    
    // Generate Lightning invoice in a separate thread to keep UI responsive
    QtConcurrent::run([this]() {
        try {
            if (BreezService::instance().isReady()) {
                QString description = QString("Donation for %1 stream").arg(
                    QString::fromStdString(d->currentAsset));
                
                // Generate invoice (this might take some time)
                QString invoice = BreezService::instance().createInvoice(
                    d->amountSats,
                    description,
                    86400  // 24h expiry
                );
                
                // Update UI in the main thread
                QMetaObject::invokeMethod(this, [this, invoice]() {
                    d->lightningInvoice = invoice.toStdString();
                    d->lightningInvoiceLabel->setText(invoice);
                    d->invoiceExpiry = QDateTime::currentDateTime().addSecs(86400);
                    d->copyLightningBtn->setEnabled(true);
                    updateQRCode();
                    setLoading(false);
                });
            } else {
                throw std::runtime_error("Breez service is not ready");
            }
        } catch (const std::exception &e) {
            QString errorMsg = QString("Failed to generate invoice: %1").arg(e.what());
            qWarning() << errorMsg;
            
            // Update UI in the main thread
            QMetaObject::invokeMethod(this, [this, errorMsg]() {
                d->lightningStatusLabel->setText(errorMsg);
                d->lightningStatusLabel->setVisible(true);
                d->copyLightningBtn->setEnabled(false);
                setLoading(false);
            });
        }
    });
    
    // Update Bitcoin QR code immediately
    updateQRCode();
}

void QRDonationsWidget::updateQRCode() {
    if (d->currentAsset.empty()) {
        return;
    }
    
    // Update Lightning QR code
    if (!d->lightningInvoice.empty()) {
        QImage lightningQR = QRGenerator::generateQRCode(
            d->lightningInvoice,
            d->lightningQRLabel->width() - 20,
            d->lightningQRLabel->height() - 20
        );
        d->lightningQRLabel->setPixmap(QPixmap::fromImage(lightningQR));
    }
    
    // Update Bitcoin QR code
    if (!d->bitcoinAddress.empty()) {
        QString qrText = QString::fromStdString(d->bitcoinAddress);
        if (d->amountSats > 0) {
            double btcAmount = d->amountSats / 100000000.0;
            qrText = QString("bitcoin:%1?amount=%2&label=Donation")
                .arg(QString::fromStdString(d->bitcoinAddress))
                .arg(btcAmount, 0, 'f', 8);
        }
        
        QImage bitcoinQR = QRGenerator::generateQRCode(
            qrText.toStdString(),
            d->bitcoinQRLabel->width() - 20,
            d->bitcoinQRLabel->height() - 20
        );
        d->bitcoinQRLabel->setPixmap(QPixmap::fromImage(bitcoinQR));
    }

    // Update Liquid QR code
    if (!d->liquidAddress.empty()) {
        QString qrText = QString::fromStdString(d->liquidAddress);
        if (d->amountSats > 0) {
            double btcAmount = d->amountSats / 100000000.0;
            qrText = QString("liquid:%1?amount=%2&label=Donation")
                .arg(QString::fromStdString(d->liquidAddress))
                .arg(btcAmount, 0, 'f', 8);
        }

        QImage liquidQR = QRGenerator::generateQRCode(
            qrText.toStdString(),
            d->liquidQRLabel->width() - 20,
            d->liquidQRLabel->height() - 20
        );
        d->liquidQRLabel->setPixmap(QPixmap::fromImage(liquidQR));
    }
    
    // Update amount hint
    if (d->amountSats > 0) {
        double btcAmount = d->amountSats / 100000000.0;
        d->amountHintLabel->setText(QString("Amount: %1 BTC (%2 sats)")
                                 .arg(btcAmount, 0, 'f', 8)
                                 .arg(d->amountSats));
    } else {
        d->amountHintLabel->setText("Scan and enter amount in your wallet");
    }
}

void QRDonationsWidget::setLoading(bool loading) {
    if (d->isLoading == loading) {
        return;
    }
    
    d->isLoading = loading;
    
    // Update UI elements based on loading state
    if (d->lightningLoadingLabel) {
        d->lightningLoadingLabel->setVisible(loading);
    }
    
    if (d->lightningQRLabel) {
        d->lightningQRLabel->setEnabled(!loading);
    }
    
    if (d->lightningInvoiceLabel) {
        d->lightningInvoiceLabel->setEnabled(!loading);
    }
    
    if (d->copyLightningBtn) {
        d->copyLightningBtn->setEnabled(!loading && !d->lightningInvoice.empty());
    }
    
    // Show/hide loading indicator
    if (d->lightningStatusLabel) {
        if (loading) {
            d->lightningStatusLabel->clear();
            d->lightningStatusLabel->setVisible(false);
        }
    }
    
    // Update cursor
    setCursor(loading ? Qt::BusyCursor : Qt::ArrowCursor);
}

void QRDonationsWidget::updateLayout() {
    // Show/hide elements based on settings
    d->assetLabel->setVisible(d->showAssetSymbol && !d->currentAsset.empty());
    d->balanceLabel->setVisible(d->showBalance);
    
    // Update copy buttons state
    if (d->copyLightningBtn) {
        d->copyLightningBtn->setEnabled(!d->isLoading && !d->lightningInvoice.empty());
    }
    
    if (d->copyBitcoinBtn) {
        d->copyBitcoinBtn->setEnabled(!d->bitcoinAddress.empty());
    }
    if (d->copyLiquidBtn) {
        d->copyLiquidBtn->setEnabled(!d->liquidAddress.empty());
    }
    
    // Update tab icons based on payment method availability
    if (d->tabWidget) {
        d->tabWidget->setTabEnabled(0, !d->lightningInvoice.empty());
        d->tabWidget->setTabEnabled(1, !d->bitcoinAddress.empty());
        d->tabWidget->setTabEnabled(2, !d->liquidAddress.empty());
    }
    
    // Update balance display if needed
    if (d->showBalance) {
        // In a real implementation, you would fetch the balance from a service
        d->balanceLabel->setText("Balance: 0.0 " + QString::fromStdString(d->currentAsset));
    }
    
    // Adjust layout
    d->mainLayout->update();
    update();
}

void QRDonationsWidget::setLightningStatus(const QString &status, bool ok) {
    if (!d->lightningStatusLabel) return;

    d->lightningStatusLabel->setText(status);
    d->lightningStatusLabel->setStyleSheet(ok ? "color: #4CAF50; font-size: 12px;" : "color: #d32f2f; font-size: 12px;");
    d->lightningStatusLabel->setVisible(!status.isEmpty());
    update();
}

void QRDonationsWidget::onPaymentReceived(qint64 amountSats, const QString &paymentHash, const QString &memo) {
    // Update UI to show the payment with flash effect instead of blocking message box
    QString message = QString("🎉 Received %1 sats!").arg(amountSats);
    if (!memo.isEmpty()) {
        message += QString("\n%1").arg(memo);
    }
    
    // Show flash overlay
    if (d->flashOverlay) {
        d->flashOverlay->setText(message);
        
        // Position overlay in the center of the widget
        int overlayWidth = width() * 0.8;
        int overlayHeight = 100;
        d->flashOverlay->setGeometry(
            (width() - overlayWidth) / 2,
            (height() - overlayHeight) / 2,
            overlayWidth,
            overlayHeight
        );
        
        d->flashOverlay->show();
        d->flashOverlay->raise();
    }
    
    // Flash background
    setStyleSheet("QRDonationsWidget { background-color: rgba(76, 175, 80, 50); }");
    
    // Hide overlay and reset background after 4 seconds
    if (d->flashTimer) {
        d->flashTimer->start(4000);
    }
    
    // Generate a new invoice for the next payment
    generateInvoices();
}

void QRDonationsWidget::onStreamStatusChanged(bool streaming) {
    if (streaming) {
        // Generate new invoices when going live
        generateInvoices();
        
        // Show a notification with payment details
        QString message = "Stream is now live!\n\n";
        
        if (!d->lightningInvoice.empty()) {
            message += "Lightning Invoice: " + QString::fromStdString(d->lightningInvoice) + "\n\n";
        }
        
        if (!d->bitcoinAddress.empty()) {
            message += "Bitcoin Address: " + QString::fromStdString(d->bitcoinAddress) + "\n\n";
        }

        if (!d->liquidAddress.empty()) {
            message += "Liquid Address: " + QString::fromStdString(d->liquidAddress) + "\n\n";
        }
        
        message += "These details have been copied to your clipboard.";
        
        // Copy to clipboard
        QClipboard *clipboard = QGuiApplication::clipboard();
        if (clipboard) {
            QString clipboardText;
            if (!d->lightningInvoice.empty()) {
                clipboardText += "Lightning: " + QString::fromStdString(d->lightningInvoice) + "\n";
            }
            if (!d->bitcoinAddress.empty()) {
                clipboardText += "Bitcoin: " + QString::fromStdString(d->bitcoinAddress) + "\n";
            }

            if (!d->liquidAddress.empty()) {
                clipboardText += "Liquid: " + QString::fromStdString(d->liquidAddress);
            }
            }
            clipboard->setText(clipboardText);
        }
        
        QMessageBox::information(this, "Stream Live", message);
    } else {
        // Clear invoices when stream ends
        d->lightningInvoice.clear();
        d->lightningInvoiceLabel->clear();
        d->lightningQRLabel->clear();
        updateQRCode();
    }
}

void QRDonationsWidget::onManageWalletClicked() {
    ManageWalletDialog dlg(this);
    dlg.exec();
}
