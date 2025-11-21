#include "breez-service.hpp"
#include <QDebug>
#include <QtConcurrent>
#include <QThread>

BreezService& BreezService::instance() {
    static BreezService _instance;
    return _instance;
}

BreezService::BreezService(QObject* parent)
    : QObject(parent)
{
    // Stubbed service: nothing to initialize
}

BreezService::~BreezService() = default;

bool BreezService::initialize(const QString& apiKey, const QString& sparkUrl, const QString& sparkAccessKey, const QString& network) {
    Q_UNUSED(apiKey)
    Q_UNUSED(sparkUrl)
    Q_UNUSED(sparkAccessKey)
    Q_UNUSED(network)
    qWarning() << "Breez SDK not available in this build. Using stub implementation.";
#ifdef BREEZ_STUB_SIMULATE
    // When simulating, report ready so UI can generate mock invoices
    emit serviceReady(true);
    return true;
#else
    emit serviceReady(false);
    return false;
#endif
}

QString BreezService::createInvoice(qint64 amountSats, const QString& description, int expirySec) {
    Q_UNUSED(expirySec)
#ifdef BREEZ_STUB_SIMULATE
    // Return a fake bolt11 invoice for UI/testing purposes (clearly a stub)
    QString invoice = QString("lnbc1stub%1").arg(QString::number(qAbs(amountSats)));
    qWarning() << "Returning simulated invoice (stub):" << invoice;
    return invoice;
#else
    Q_UNUSED(amountSats)
    Q_UNUSED(description)
    qWarning() << "Breez SDK not available. Cannot create invoice.";
    return QString();
#endif
}

QString BreezService::nodeInfo() const {
    return QString("Breez SDK unavailable");
}

qint64 BreezService::balance() const {
    return 0;
}

QVariantList BreezService::paymentHistory() const {
    return {};
}

bool BreezService::sendLightningPayment(const QString &bolt11) {
    Q_UNUSED(bolt11)
#ifdef BREEZ_STUB_SIMULATE
    // Simulate an async successful send with a fake payment id
    QtConcurrent::run([this]() {
        QThread::sleep(1);
        emit sendCompleted(true, QString("stub-payment-id-12345"));
    });
    return true;
#else
    QtConcurrent::run([this]() {
        QThread::sleep(1);
        emit sendCompleted(false, QString("Breez SDK not available (stub)."));
    });
    return false;
#endif
}

bool BreezService::sendOnChain(const QString &address, qint64 amountSats, const QString &network) {
    Q_UNUSED(address)
    Q_UNUSED(amountSats)
    Q_UNUSED(network)
#ifdef BREEZ_STUB_SIMULATE
    QtConcurrent::run([this]() {
        QThread::sleep(1);
        emit sendCompleted(true, QString("stub-txid-0xdeadbeef"));
    });
    return true;
#else
    QtConcurrent::run([this]() {
        QThread::sleep(1);
        emit sendCompleted(false, QString("Breez SDK not available (stub)."));
    });
    return false;
#endif
}
