#pragma once

#include <QObject>
#include <QString>
#include <QTimer>
#include <memory>

// Forward declarations
namespace breez_sdk {
class SDK;
struct PaymentReceivedEvent;
}

class BreezHandler : public QObject {
    Q_OBJECT
    
public:
    static BreezHandler& instance();
    
    // Initialize the Breez SDK with API key and configuration
    bool initialize(const QString& apiKey, const QString& workingDir);
    
    // Generate a payment request
    QString createInvoice(qint64 amountSats, const QString& description = "", 
                         int expirySec = 3600);
    
    // Check if Breez SDK is initialized and ready
    bool isReady() const { return m_initialized; }
    
    // Get current node info
    QString nodeInfo() const;
    
    // Get current balance in satoshis
    qint64 balance() const;
    
signals:
    void paymentReceived(qint64 amountSats, const QString& paymentHash);
    void paymentFailed(const QString& error);
    void serviceReady(bool ready);
    
private slots:
    void checkForPayments();
    
private:
    BreezHandler(QObject *parent = nullptr);
    ~BreezHandler();
    
    // Disable copy and move
    BreezHandler(const BreezHandler&) = delete;
    BreezHandler& operator=(const BreezHandler&) = delete;
    BreezHandler(BreezHandler&&) = delete;
    BreezHandler& operator=(BreezHandler&&) = delete;
    
    void setupPaymentListener();
    void onPaymentReceived(const breez_sdk::PaymentReceivedEvent& payment);
    
    std::unique_ptr<breez_sdk::SDK> m_sdk;
    QTimer m_pollingTimer;
    bool m_initialized = false;
    QString m_apiKey;
    QString m_workingDir;
};
