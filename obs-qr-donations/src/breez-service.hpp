#pragma once

#include <QObject>
#include <QString>
#include <QTimer>
#include <memory>

// Forward declarations
namespace breez_sdk {
class SDK;
struct InvoicePaid;
struct SparkConfig;
}

class BreezService : public QObject {
    Q_OBJECT
    
public:
    static BreezService& instance();
    
    // Initialize the Breez service with Spark wallet
    // network: "bitcoin" (default) or "liquid"
    bool initialize(const QString& apiKey, const QString& sparkUrl,
                   const QString& sparkAccessKey,
                   const QString& network = "bitcoin");
    
    // Create a new invoice
    QString createInvoice(qint64 amountSats, const QString& description = "", 
                         int expirySec = 3600);
    
    // Check if Breez service is ready
    bool isReady() const { return m_initialized; }
    
    // Get current node info
    QString nodeInfo() const;
    
    // Get current balance in satoshis
    qint64 balance() const;
    
    // Get payment history
    QVariantList paymentHistory() const;

    // Outgoing payments / withdrawals
    // Pay a Lightning invoice (bolt11) — returns true if send operation started
    bool sendLightningPayment(const QString &bolt11);

    // Send on-chain (bitcoin or liquid) to an address. Returns true if operation started
    bool sendOnChain(const QString &address, qint64 amountSats, const QString &network = "bitcoin");

signals:
    void sendCompleted(bool ok, const QString &txid_or_err);
    
    // Deprecated: old BreezHandler is removed - use BreezService.
    
signals:
    void paymentReceived(qint64 amountSats, const QString& paymentHash, const QString& memo);
    void serviceReady(bool ready);
    void errorOccurred(const QString& error);
    
private slots:
    void checkForPayments();
    void retryInitialization();
    
private:
    BreezService(QObject *parent = nullptr);
    ~BreezService();
    
    // Disable copy and move
    BreezService(const BreezService&) = delete;
    BreezService& operator=(const BreezService&) = delete;
    BreezService(BreezService&&) = delete;
    BreezService& operator=(BreezService&&) = delete;
    
    void setupPaymentListener();
    void onPaymentReceived(const breez_sdk::InvoicePaid& payment);
    
    std::unique_ptr<breez_sdk::SDK> m_sdk;
    std::unique_ptr<breez_sdk::SparkConfig> m_sparkConfig;
    QTimer m_pollingTimer;
    bool m_initialized = false;
    QString m_apiKey;
    QString m_workingDir;
    
    // Additional state for error handling and robustness
    QTimer m_retryTimer;
    int m_retryCount = 0;
    QString m_lastError;
    QString m_sparkUrl;
    QString m_sparkAccessKey;
    QSet<QString> m_processedPayments;
    QNetworkAccessManager *m_networkManager = nullptr;
};
