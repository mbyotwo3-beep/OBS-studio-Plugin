#include "breez-handler.hpp"
#include <QStandardPaths>
#include <QDir>
#include <QDebug>

// Breez SDK includes
#include <breez_sdk/breez_sdk.h>

using namespace breez_sdk;

BreezHandler& BreezHandler::instance() {
    static BreezHandler instance;
    return instance;
}

BreezHandler::BreezHandler(QObject *parent)
    : QObject(parent)
    , m_initialized(false)
    , m_workingDir(QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/breez") {
    
    // Ensure working directory exists
    QDir().mkpath(m_workingDir);
    
    // Set up polling timer for payment checks
    connect(&m_pollingTimer, &QTimer::timeout, this, &BreezHandler::checkForPayments);
    m_pollingTimer.start(5000); // Check every 5 seconds
}

BreezHandler::~BreezHandler() {
    m_pollingTimer.stop();
    // Clean up Breez SDK
    if (m_sdk) {
        // Properly shut down the SDK
    }
}

bool BreezHandler::initialize(const QString& apiKey, const QString& workingDir) {
    if (m_initialized) {
        return true;
    }
    
    if (!workingDir.isEmpty()) {
        m_workingDir = workingDir;
        QDir().mkpath(m_workingDir);
    }
    
    m_apiKey = apiKey;
    
    try {
        // Initialize Breez SDK configuration
        Config config;
        config.working_dir = m_workingDir.toStdString();
        config.api_key = m_apiKey.toStdString();
        config.network = Network::BITCOIN;
        config.log_level = LogLevel::DEBUG;
        
        // Create SDK instance
        m_sdk = std::make_unique<SDK>(config);
        
        // Set up payment listener
        setupPaymentListener();
        
        m_initialized = true;
        emit serviceReady(true);
        return true;
        
    } catch (const std::exception &e) {
        qWarning() << "Failed to initialize Breez SDK:" << e.what();
        emit serviceReady(false);
        return false;
    }
}

QString BreezHandler::createInvoice(qint64 amountSats, const QString& description, int expirySec) {
    if (!m_initialized || !m_sdk) {
        qWarning() << "Breez SDK not initialized";
        return "";
    }
    
    try {
        // Create invoice request
        CreateInvoiceRequest req;
        req.amount_msat = amountSats * 1000; // Convert to millisatoshis
        req.description = description.toStdString();
        req.expiry = expirySec;
        
        // Generate invoice
        auto invoice = m_sdk->create_invoice(req);
        return QString::fromStdString(invoice.bolt11);
        
    } catch (const std::exception &e) {
        qWarning() << "Failed to create invoice:" << e.what();
        return "";
    }
}

QString BreezHandler::nodeInfo() const {
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

qint64 BreezHandler::balance() const {
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

void BreezHandler::setupPaymentListener() {
    if (!m_sdk) return;
    
    // Set up payment listener
    m_sdk->set_payment_listener([this](const PaymentReceivedEvent& event) {
        // This will be called on a background thread
        QMetaObject::invokeMethod(this, [this, event]() {
            onPaymentReceived(event);
        }, Qt::QueuedConnection);
    });
}

void BreezHandler::onPaymentReceived(const PaymentReceivedEvent& payment) {
    qint64 amountSats = payment.amount_msat / 1000; // Convert to satoshis
    QString paymentHash = QString::fromStdString(payment.payment_hash);
    
    qInfo() << "Payment received:" << amountSats << "sats, hash:" << paymentHash;
    
    emit paymentReceived(amountSats, paymentHash);
}

void BreezHandler::checkForPayments() {
    if (!m_initialized || !m_sdk) {
        return;
    }
    
    try {
        // Check for new payments
        auto payments = m_sdk->list_payments(ListPaymentsRequest{});
        
        // Process new payments
        for (const auto& payment : payments) {
            if (payment.status == "complete" && !payment.payment_preimage.empty()) {
                // This is a completed payment
                PaymentReceivedEvent event;
                event.amount_msat = payment.amount_msat;
                event.payment_hash = payment.payment_hash;
                
                onPaymentReceived(event);
            }
        }
    } catch (const std::exception &e) {
        qWarning() << "Error checking for payments:" << e.what();
    }
}
