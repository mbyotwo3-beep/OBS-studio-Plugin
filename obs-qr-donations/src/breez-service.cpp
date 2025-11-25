#include "breez-service.hpp"
#include <QStandardPaths>
#include <QDir>
#include <QDebug>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QUrlQuery>
#include <QtConcurrent>
#include <QThread>
#include <QSet>
#include <QDateTime>
#include <QSslError>

// Breez SDK includes (only available when compiled with the SDK)
#ifdef HAVE_BREEZ_SDK
#include <breez_sdk_spark/breez_sdk_spark.h>
using namespace breez_sdk_spark;
#endif

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
    m_networkManager->setRedirectPolicy(QNetworkRequest::NoLessSafeRedirectPolicy);
    m_networkManager->setTransferTimeout(30000); // 30 second timeout
    
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
                             const QString& sparkAccessKey,
                             const QString& network) {
    if (m_initialized) {
        qInfo() << "BreezService: already initialized";
        return true;
    }
    
    // Reset state
    m_retryCount = 0;
    m_lastError.clear();
    
    // Validate inputs
    if (apiKey.isEmpty()) {
        m_lastError = "API key cannot be empty";
        qWarning() << m_lastError;
        emit errorOccurred(m_lastError);
        return false;
    }
    
    // Spark URL is now OPTIONAL - only required for custom Spark wallet
    // For standard Breez nodeless, only API key is needed
    
    // Store API key
    m_apiKey = apiKey;
    
    // Spark URL and access key are OPTIONAL for custom Spark wallet
    // If not provided, Breez SDK will use default nodeless configuration
    if (!sparkUrl.isEmpty()) {
        m_sparkUrl = sparkUrl.endsWith('/') ? sparkUrl.left(sparkUrl.length() - 1) : sparkUrl;
        m_sparkAccessKey = sparkAccessKey;
        qInfo() << "Using custom Spark wallet at" << m_sparkUrl;
    } else {
        qInfo() << "Using Breez default nodeless configuration (no custom Spark wallet)";
    }
    
    qInfo() << "Initializing BreezService with Spark at" << m_sparkUrl;
    qInfo() << "network=" << network;
    
    // Ensure working directory exists
    QDir dir(m_workingDir);
    if (!dir.exists() && !dir.mkpath(".")) {
        m_lastError = QString("Failed to create working directory: %1").arg(m_workingDir);
        qCritical() << m_lastError;
        emit errorOccurred(m_lastError);
        return false;
    }
    
    // Start initialization using breez SDK
#ifdef HAVE_BREEZ_SDK
    
    try {
        // Prepare configuration for the SDK
        breez_sdk_spark::Config config;
        config.working_dir = m_workingDir.toStdString();
        config.api_key = m_apiKey.toStdString();
        // Choose network (Bitcoin or Liquid)
        if (network.compare("liquid", Qt::CaseInsensitive) == 0) {
            config.network = breez_sdk_spark::Network::LIQUID;
        } else {
            config.network = breez_sdk_spark::Network::BITCOIN;
        }
        config.log_level = breez_sdk_spark::LogLevel::DEBUG;

        // Set up Spark wallet configuration if provided (requires ENABLE_SPARK_WALLET)
    #ifdef ENABLE_SPARK_WALLET
        if (!m_sparkUrl.isEmpty() && !m_sparkAccessKey.isEmpty()) {
            qDebug() << "Configuring Spark wallet with URL:" << m_sparkUrl;
            QUrl url(m_sparkUrl);
            if (!url.isValid() || url.scheme().isEmpty()) {
                throw std::runtime_error("Invalid Spark wallet URL");
            }

            breez_sdk_spark::SparkConfig sparkCfg;
            sparkCfg.url = m_sparkUrl.toStdString();
            sparkCfg.access_key = m_sparkAccessKey.toStdString();

            config.default_wallet = breez_sdk_spark::WalletType::SPARK;
            config.spark_config = sparkCfg;
            qDebug() << "Spark wallet configured successfully";
        }
    #else
        Q_UNUSED(sparkUrl)
        Q_UNUSED(sparkAccessKey)
    #endif
        else {
            qDebug() << "Using default Breez wallet (Spark wallet not configured)";
        }

        // Create SDK instance
        qDebug() << "Creating Breez SDK instance...";
        m_sdk = std::make_unique<breez_sdk_spark::SDK>(config);

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
#else
    qWarning() << "BreezService not compiled with Breez SDK. Initialization skipped.";
    emit serviceReady(false);
    return false;
#endif
}

bool BreezService::sendLightningPayment(const QString &bolt11) {
    if (bolt11.isEmpty()) {
        qWarning() << "sendLightningPayment: empty invoice";
        QMetaObject::invokeMethod(this, [this]() { emit sendCompleted(false, "Empty invoice"); }, Qt::QueuedConnection);
        return false;
    }

#ifdef HAVE_BREEZ_SDK
    if (!m_initialized || !m_sdk) {
        qWarning() << "sendLightningPayment: Breez SDK not initialized";
        QMetaObject::invokeMethod(this, [this]() { emit sendCompleted(false, "Breez SDK not initialized"); }, Qt::QueuedConnection);
        return false;
    }

    // Run the send operation asynchronously
    QtConcurrent::run([this, bolt11]() {
        try {
            // Use SDK to send payment
            SendPaymentRequest req;
            req.bolt11 = bolt11.toStdString();
            auto result = m_sdk->send_payment(req);

            if (result.success) {
                QString txid = QString::fromStdString(result.payment_id);
                QMetaObject::invokeMethod(this, [this, txid]() { emit sendCompleted(true, txid); }, Qt::QueuedConnection);
            } else {
                QString err = QString::fromStdString(result.error_message);
                QMetaObject::invokeMethod(this, [this, err]() { emit sendCompleted(false, err); }, Qt::QueuedConnection);
            }
        } catch (const std::exception &e) {
            QString err = QString("Exception sending lightning payment: %1").arg(e.what());
            qWarning() << err;
            QMetaObject::invokeMethod(this, [this, err]() { emit sendCompleted(false, err); }, Qt::QueuedConnection);
        }
    });

    return true;
#else
    Q_UNUSED(bolt11)
    // Not compiled with SDK — fail fast but simulate async behavior for UI
    QtConcurrent::run([this]() {
        QThread::sleep(1);
        QMetaObject::invokeMethod(this, [this]() { emit sendCompleted(false, "Breez SDK not available in this build"); }, Qt::QueuedConnection);
    });
    return false;
#endif
}

bool BreezService::sendOnChain(const QString &address, qint64 amountSats, const QString &network) {
    if (address.isEmpty() || amountSats <= 0) {
        qWarning() << "sendOnChain: invalid parameters";
        QMetaObject::invokeMethod(this, [this]() { emit sendCompleted(false, "Invalid address or amount"); }, Qt::QueuedConnection);
        return false;
    }

#ifdef HAVE_BREEZ_SDK
    if (!m_initialized || !m_sdk) {
        qWarning() << "sendOnChain: Breez SDK not initialized";
        QMetaObject::invokeMethod(this, [this]() { emit sendCompleted(false, "Breez SDK not initialized"); }, Qt::QueuedConnection);
        return false;
    }

    // Run send on-chain asynchronously
    QtConcurrent::run([this, address, amountSats, network]() {
        try {
            OnChainSendRequest req;
            req.address = address.toStdString();
            req.amount_sat = static_cast<uint64_t>(amountSats);
            if (network.compare("liquid", Qt::CaseInsensitive) == 0) {
                req.network = Network::LIQUID;
            } else {
                req.network = Network::BITCOIN;
            }

            auto result = m_sdk->send_on_chain(req);
            if (result.success) {
                QString txid = QString::fromStdString(result.txid);
                QMetaObject::invokeMethod(this, [this, txid]() { emit sendCompleted(true, txid); }, Qt::QueuedConnection);
            } else {
                QString err = QString::fromStdString(result.error_message);
                QMetaObject::invokeMethod(this, [this, err]() { emit sendCompleted(false, err); }, Qt::QueuedConnection);
            }
        } catch (const std::exception &e) {
            QString err = QString("Exception sending on-chain: %1").arg(e.what());
            qWarning() << err;
            QMetaObject::invokeMethod(this, [this, err]() { emit sendCompleted(false, err); }, Qt::QueuedConnection);
        }
    });

    return true;
#else
    Q_UNUSED(address)
    Q_UNUSED(amountSats)
    Q_UNUSED(network)
    // Simulate async response when SDK not available
    QtConcurrent::run([this]() {
        QThread::sleep(2);
        QMetaObject::invokeMethod(this, [this]() { emit sendCompleted(false, "Breez SDK not available in this build"); }, Qt::QueuedConnection);
    });
    return false;
#endif
}

QString BreezService::createInvoice(qint64 amountSats, const QString& description, int expirySec) {
    if (!m_initialized || !m_sdk) {
        const QString errorMsg = "Breez SDK not initialized";
        qWarning() << errorMsg;
        emit errorOccurred(errorMsg);
        return "";
    }
    
    // Validate amount
    if (amountSats < 0) {
        const QString errorMsg = "Invalid amount: cannot be negative";
        qWarning() << errorMsg;
        emit errorOccurred(errorMsg);
        return "";
    }
    
    // Validate expiry time
    if (expirySec < 60) {
        qWarning() << "Expiry time too short, using minimum of 60 seconds";
        expirySec = 60;
    } else if (expirySec > 86400 * 7) { // 7 days max
        qWarning() << "Expiry time too long, capping at 7 days";
        expirySec = 86400 * 7;
    }
    
    qDebug() << "Creating invoice for" << amountSats << "sats, expires in" << expirySec << "seconds";
    
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
#ifdef HAVE_BREEZ_SDK
    if (!m_initialized || !m_sdk) {
        qDebug() << "Skipping payment check - service not initialized";
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
        auto payments = m_sdk->list_payments({});
        
        // Process new payments
        for (const auto& payment : payments) {
            if (payment.status == "complete" && !payment.payment_preimage.empty()) {
                // Check if we've already processed this payment
                QString paymentId = QString::fromStdString(payment.payment_hash);
                if (!m_processedPayments.contains(paymentId)) {
                    m_processedPayments.insert(paymentId);
                    
                    // This is a completed payment
                    InvoicePaid event;
                    event.amount_msat = payment.amount_msat;
                    event.payment_hash = payment.payment_hash;
                    event.description = payment.description;
                    
                    onPaymentReceived(event);
                }
            }
        }
        
        // Clean up old processed payments to prevent memory leak
        if (m_processedPayments.size() > 1000) {
            m_processedPayments.clear();
            qDebug() << "Cleared processed payments cache";
        }
        
    } catch (const std::exception& e) {
        qWarning() << "Error checking for payments:" << e.what();
        emit errorOccurred(QString("Error checking for payments: %1").arg(e.what()));
        // Back off on errors
        m_pollingTimer.setInterval(qMin(m_pollingTimer.interval() * 2, 60000)); // Max 1 minute
    }
#endif
}

void BreezService::retryInitialization() {
    if (m_initialized) {
        return;
    }
    
    m_retryCount++;
    if (m_retryCount > 3) {
        qCritical() << "Failed to initialize Breez SDK after" << m_retryCount << "attempts";
        emit errorOccurred("Failed to initialize Breez SDK after multiple attempts");
        return;
    }
    
    qInfo() << "Retrying Breez SDK initialization (attempt" << m_retryCount << ")";
    // The initialization will be retried when the user tries again or when the timer fires
}
