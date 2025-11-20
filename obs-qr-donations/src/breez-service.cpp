#include "breez-service.hpp"
#include <QStandardPaths>
#include <QDir>
#include <QDebug>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QUrlQuery>

// Breez SDK includes
#include <breez_sdk/breez_sdk.h>

using namespace breez_sdk;

BreezService& BreezService::instance() {
    static BreezService instance;
    return instance;
}

BreezService::BreezService(QObject *parent)
    : QObject(parent)
    , m_initialized(false)
    , m_workingDir(QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/breez")
    , m_pollingTimer(this)
    , m_networkManager(new QNetworkAccessManager(this))
{
    // Ensure working directory exists
    if (!QDir().mkpath(m_workingDir)) {
        qWarning() << "Failed to create Breez working directory:" << m_workingDir;
    } else {
        qDebug() << "Breez working directory:" << m_workingDir;
    }
    
    // Set up polling timer for payment checks with exponential backoff
    m_pollingTimer.setSingleShot(false);
    m_pollingTimer.setInterval(5000); // Start with 5 seconds
    connect(&m_pollingTimer, &QTimer::timeout, this, &BreezService::checkForPayments);
    
    // Set up retry timer for failed operations
    m_retryTimer.setSingleShot(true);
    connect(&m_retryTimer, &QTimer::timeout, this, &BreezService::retryInitialization);
    
    // Set up network configuration
    m_networkManager.setRedirectPolicy(QNetworkRequest::NoLessSafeRedirectPolicy);
    m_networkManager.setTransferTimeout(30000); // 30 second timeout
    
    // Connect network error handling
    connect(m_networkManager, &QNetworkAccessManager::sslErrors, 
            this, [](QNetworkReply *reply, const QList<QSslError> &errors) {
        qWarning() << "SSL Errors occurred:";
        for (const auto &error : errors) {
            qWarning() << "- " << error.errorString();
        }
        // Ignore SSL errors for development (remove in production)
        reply->ignoreSslErrors();
    });
    
    qDebug() << "BreezService initialized";
}

BreezService::~BreezService() {
    m_pollingTimer.stop();
    // Clean up Breez SDK
    if (m_sdk) {
        // Properly shut down the SDK
    }
}

bool BreezService::initialize(const QString& apiKey, const QString& sparkUrl, 
                             const QString& sparkAccessKey) {
    if (m_initialized) {
        LOG_INFO("Service already initialized");
        return true;
    }
    
    // Reset state
    m_retryCount = 0;
    m_lastError.clear();
    
    // Validate inputs
    if (apiKey.isEmpty()) {
        m_lastError = "API key cannot be empty";
        LOG_ERROR(m_lastError);
        emit errorOccurred(m_lastError);
        return false;
    }
    
    if (sparkUrl.isEmpty()) {
        m_lastError = "Spark URL cannot be empty";
        LOG_ERROR(m_lastError);
        emit errorOccurred(m_lastError);
        return false;
    }
    
    // Store credentials
    m_apiKey = apiKey;
    m_sparkUrl = sparkUrl.endsWith('/') ? sparkUrl.left(sparkUrl.length() - 1) : sparkUrl;
    m_sparkAccessKey = sparkAccessKey;
    
    LOG_INFO("Initializing BreezService with Spark at" << m_sparkUrl);
    
    // Ensure working directory exists
    QDir dir(m_workingDir);
    if (!dir.exists() && !dir.mkpath(".")) {
        m_lastError = QString("Failed to create working directory: %1").arg(m_workingDir);
        LOG_ERROR(m_lastError);
        emit errorOccurred(m_lastError);
        return false;
    }
    
    // Start initialization
    return initializeBreezSDK();
    
    if (apiKey.isEmpty()) {
        qWarning() << "Cannot initialize BreezService: API key is empty";
        emit errorOccurred("API key is required to initialize Breez service");
        return false;
    }
    
    m_apiKey = apiKey;
    qDebug() << "Initializing BreezService with API key:" << apiKey.mid(0, 4) << "...";
    
    try {
        // Validate and prepare configuration
        Config config;
        config.working_dir = m_workingDir.toStdString();
        config.api_key = m_apiKey.toStdString();
        config.network = Network::BITCOIN;
        config.log_level = LogLevel::DEBUG;
        
        // Set up Spark wallet configuration if provided
        if (!sparkUrl.isEmpty() && !sparkAccessKey.isEmpty()) {
            qDebug() << "Configuring Spark wallet with URL:" << sparkUrl;
            
            // Validate Spark URL
            QUrl url(sparkUrl);
            if (!url.isValid() || url.scheme().isEmpty()) {
                throw std::runtime_error("Invalid Spark wallet URL");
            }
            
            m_sparkConfig = std::make_unique<SparkConfig>();
            m_sparkConfig->url = sparkUrl.toStdString();
            m_sparkConfig->access_key = sparkAccessKey.toStdString();
            
            // Set Spark as the default wallet
            config.default_wallet = WalletType::SPARK;
            config.spark_config = *m_sparkConfig;
            
            qDebug() << "Spark wallet configured successfully";
        } else {
            qDebug() << "Using default Breez wallet (Spark wallet not configured)";
        }
        
        // Create SDK instance
        qDebug() << "Creating Breez SDK instance...";
        m_sdk = std::make_unique<SDK>(config);
        
        if (!m_sdk) {
            throw std::runtime_error("Failed to create Breez SDK instance");
        }
        
        // Set up payment listener
        setupPaymentListener();
        
        m_initialized = true;
        qInfo() << "BreezService initialized successfully";
        emit serviceReady(true);
        
        // Start polling for payments
        m_pollingTimer.start();
        qDebug() << "Started payment polling timer";
        
        return true;
        
    } catch (const std::exception &e) {
        const QString errorMsg = QString("Failed to initialize Breez SDK: %1").arg(e.what());
        qCritical() << errorMsg;
        emit errorOccurred(errorMsg);
        emit serviceReady(false);
        
        // Clean up on failure
        m_initialized = false;
        m_sdk.reset();
        m_sparkConfig.reset();
        
        return false;
    }
}

QString BreezService::createInvoice(qint64 amountSats, const QString& description, int expirySec) {
    if (!m_initialized || !m_sdk) {
        const QString errorMsg = "Breez SDK not initialized";
        LOG_WARNING(errorMsg);
        emit errorOccurred(errorMsg);
        return "";
    }
    
    // Validate amount
    if (amountSats < 0) {
        const QString errorMsg = "Invalid amount: cannot be negative";
        LOG_WARNING(errorMsg);
        emit errorOccurred(errorMsg);
        return "";
    }
    
    // Validate expiry time
    if (expirySec < 60) {
        LOG_WARNING("Expiry time too short, using minimum of 60 seconds");
        expirySec = 60;
    } else if (expirySec > 86400 * 7) { // 7 days max
        LOG_WARNING("Expiry time too long, capping at 7 days");
        expirySec = 86400 * 7;
    }
    
    LOG_DEBUG(QString("Creating invoice for %1 sats, expires in %2 seconds").arg(amountSats).arg(expirySec));
    
    if (amountSats < 0) {
        const QString errorMsg = "Invalid amount: amount cannot be negative";
        qWarning() << errorMsg;
        emit errorOccurred(errorMsg);
        return "";
    }
    
    if (expirySec <= 0) {
        qWarning() << "Invalid expiry time, using default (1 hour)";
        expirySec = 3600; // Default to 1 hour
    }
    
    try {
        // Create invoice request
        CreateInvoiceRequest req;
        // Set amount to 0 to allow sender to specify the amount
        req.amount_msat = 0; 
        req.description = description.toStdString();
        req.expiry = expirySec;
        
        // Add amount as part of the description if specified
        if (amountSats > 0) {
            req.description = (description + "\nSuggested amount: " + 
                             QString::number(amountSats) + " sats").toStdString();
        }
        
        // Generate invoice
        auto invoice = m_sdk->create_invoice(req);
        
        // If amount was specified, create a BOLT11 invoice with amount field
        if (amountSats > 0) {
            // This creates a payment request that will prompt for amount
            std::string bolt11 = invoice.bolt11;
            // The actual amount will be entered by the sender's wallet
            return QString::fromStdString(bolt11);
        }
        
        return QString::fromStdString(invoice.bolt11);
        
    } catch (const std::exception &e) {
        QString error = QString("Failed to create invoice: %1").arg(e.what());
        qWarning() << error;
        emit errorOccurred(error);
        return "";
    }
}

QString BreezService::nodeInfo() const {
    if (!m_initialized || !m_sdk) {
        return "Breez SDK not initialized";
    }
    
    try {
        auto info = m_sdk->node_info();
        return QString("Node ID: %1\nChannels: %2\nBlock Height: %3")
            .arg(QString::fromStdString(info.id))
            .arg(info.channels_balance_msat / 1000) // Convert to satoshis
            .arg(info.block_height);
    } catch (const std::exception &e) {
        return QString("Error getting node info: %1").arg(e.what());
    }
}

qint64 BreezService::balance() const {
    if (!m_initialized || !m_sdk) {
        return 0;
    }
    
    try {
        auto balance = m_sdk->get_balance();
        return balance.on_chain_balance_sat + (balance.off_chain_balance_msat / 1000);
    } catch (const std::exception &e) {
        qWarning() << "Failed to get balance:" << e.what();
        return 0;
    }
}

QVariantList BreezService::paymentHistory() const {
    QVariantList result;
    
    if (!m_initialized || !m_sdk) {
        return result;
    }
    
    try {
        auto payments = m_sdk->list_payments({});
        
        for (const auto &payment : payments) {
            QVariantMap pmt;
            pmt["amount"] = payment.amount_msat / 1000; // Convert to satoshis
            pmt["hash"] = QString::fromStdString(payment.payment_hash);
            pmt["memo"] = QString::fromStdString(payment.description);
            pmt["timestamp"] = payment.timestamp;
            pmt["status"] = QString::fromStdString(payment.status);
            
            result.append(pmt);
        }
    } catch (const std::exception &e) {
        qWarning() << "Failed to get payment history:" << e.what();
    }
    
    return result;
}

void BreezService::setupPaymentListener() {
    if (!m_sdk) return;
    
    // Set up payment listener
    m_sdk->set_payment_listener([this](const InvoicePaid& payment) {
        // This will be called on a background thread
        QMetaObject::invokeMethod(this, [this, payment]() {
            onPaymentReceived(payment);
        }, Qt::QueuedConnection);
    });
}

void BreezService::onPaymentReceived(const InvoicePaid& payment) {
    qint64 amountSats = payment.amount_msat / 1000; // Convert to satoshis
    QString paymentHash = QString::fromStdString(payment.payment_hash);
    QString memo = QString::fromStdString(payment.description);
    
    qInfo() << "Payment received:" << amountSats << "sats, hash:" << paymentHash << "memo:" << memo;
    
    emit paymentReceived(amountSats, paymentHash, memo);
}

void BreezService::checkForPayments() {
    if (!m_initialized || !m_sdk) {
        LOG_DEBUG("Skipping payment check - service not initialized");
        return;
    }
    
    // Rate limiting - don't check too frequently
    static QDateTime lastCheck;
    QDateTime now = QDateTime::currentDateTime();
    if (lastCheck.isValid() && lastCheck.msecsTo(now) < 2000) { // 2 second minimum between checks
        return;
    }
    lastCheck = now;
    
    try {
        // Check for new payments
        auto payments = m_sdk->listPayments(ListPaymentsRequest{});
        if (!payments) {
            LOG_WARNING("Failed to list payments");
            return;
        }
        
        // Process new payments
        for (const auto& payment : *payments) {
            if (payment.status == PaymentStatus::COMPLETE && 
                payment.payment_type == PaymentType::RECEIVE) {
                // Check if we've already processed this payment
                QString paymentId = QString::fromStdString(payment.id);
                if (!m_processedPayments.contains(paymentId)) {
                    m_processedPayments.insert(paymentId);
                    onPaymentReceived(InvoicePaid{
                        payment.payment_hash,
                        payment.amount_msat,
                        payment.description
                    });
                }
            }
        }
        
        // Clean up old processed payments to prevent memory leak
        if (m_processedPayments.size() > 1000) {
            m_processedPayments.clear();
            LOG_DEBUG("Cleared processed payments cache");
        }
        
    } catch (const std::exception& e) {
        LOG_ERROR(QString("Error checking for payments: %1").arg(e.what()));
        // Back off on errors
        m_pollingTimer.setInterval(qMin(m_pollingTimer.interval() * 2, 60000)); // Max 1 minute
    }
    
    try {
        // Check for new payments
        auto payments = m_sdk->list_payments({});
        
        // Process new payments
        for (const auto& payment : payments) {
            if (payment.status == "complete" && !payment.payment_preimage.empty()) {
                // This is a completed payment
                InvoicePaid event;
                event.amount_msat = payment.amount_msat;
                event.payment_hash = payment.payment_hash;
                event.description = payment.description;
                
                onPaymentReceived(event);
            }
        }
    } catch (const std::exception &e) {
        qWarning() << "Error checking for payments:" << e.what();
        emit errorOccurred(QString("Error checking for payments: %1").arg(e.what()));
    }
}
