#include "breez-service.hpp"
#include <QDebug>
#include <QDir>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QSet>
#include <QStandardPaths>
#include <QThread>
#include <QUrlQuery>
#include <QtConcurrent>
#include <QRandomGenerator>

// Breez SDK includes (only available when compiled with the SDK)
#ifdef HAVE_BREEZ_SDK
#include <breez_sdk/breez_sdk.hpp>
#endif

using namespace breez_sdk;

BreezService &BreezService::instance() {
  static BreezService instance;
  return instance;
}

BreezService::BreezService(QObject *parent)
    : QObject(parent), m_initialized(false),
      m_workingDir(
          QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) +
          "/breez"),
      m_pollingTimer(this), m_networkManager(new QNetworkAccessManager(this)) {
  // Ensure working directory exists
  if (!QDir().mkpath(m_workingDir)) {
    qWarning() << "Failed to create Breez working directory:" << m_workingDir;
  } else {
    qDebug() << "Breez working directory:" << m_workingDir;
  }

  // Set up polling timer for payment checks with exponential backoff
  m_pollingTimer.setSingleShot(false);
  m_pollingTimer.setInterval(5000); // Start with 5 seconds
  connect(&m_pollingTimer, &QTimer::timeout, this,
          &BreezService::checkForPayments);

  // Set up retry timer for failed operations
  m_retryTimer.setSingleShot(true);
  connect(&m_retryTimer, &QTimer::timeout, this,
          &BreezService::retryInitialization);

  // Set up network configuration
  m_networkManager->setRedirectPolicy(QNetworkRequest::NoLessSafeRedirectPolicy);
  m_networkManager->setTransferTimeout(30000); // 30 second timeout

  // Connect network error handling
  connect(m_networkManager, &QNetworkAccessManager::sslErrors, this,
          [](QNetworkReply *reply, const QList<QSslError> &errors) {
            qWarning() << "SSL Errors occurred:";
            for (const auto &error : errors) {
              qWarning() << "- " << error.errorString();
            }
            // In production, we should NOT ignore SSL errors.
            // reply->ignoreSslErrors();
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

bool BreezService::initialize(const QString &apiKey, const QString &sparkUrl,
                              const QString &sparkAccessKey,
                              const QString &network) {
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

  // Store credentials
  m_apiKey = apiKey;
  // Spark args ignored for Greenlight
  
  qInfo() << "Initializing BreezService (Greenlight)";
  qInfo() << "network=" << network;

  // Ensure working directory exists
  QDir dir(m_workingDir);
  if (!dir.exists() && !dir.mkpath(".")) {
    m_lastError =
        QString("Failed to create working directory: %1").arg(m_workingDir);
    qCritical() << m_lastError;
    emit errorOccurred(m_lastError);
    return false;
  }

  // Start initialization using breez SDK
#ifdef HAVE_BREEZ_SDK



    try {
        // Ensure working directory exists
        if (!QDir().mkpath(m_workingDir)) {
             qWarning() << "Failed to create working dir:" << m_workingDir;
             return false;
        }

        // Load or generate seed
        QString seedPath = m_workingDir + "/seed.dat";
        QByteArray seedBytes;
        QFile seedFile(seedPath);
        if (seedFile.exists() && seedFile.open(QIODevice::ReadOnly)) {
            seedBytes = seedFile.readAll();
            seedFile.close();
        }
        
        if (seedBytes.size() != 32) {
            qInfo() << "Generating new seed...";
            seedBytes.resize(32);
            // Use a securely seeded random generator for the wallet seed
            QRandomGenerator secureRng = QRandomGenerator::securelySeeded();
            for (int i = 0; i < 32; ++i) {
                seedBytes[i] = static_cast<char>(secureRng.generate());
            }
            
            if (seedFile.open(QIODevice::WriteOnly)) {
                seedFile.write(seedBytes);
                seedFile.close();
            }
        }

        // Create default config for Greenlight (EnvironmentType::PRODUCTION)
        breez_sdk::NodeConfig node_config;
        // Variant 1 (Greenlight) - 6 bytes: 00 00 00 01 00 00
        node_config.raw_data = {0, 0, 0, 1, 0, 0};
        
        auto config = breez_sdk::SDK::default_config(
            breez_sdk::EnvironmentType::PRODUCTION,
            apiKey.toStdString(),
            node_config
        );

        // Set working dir
        config.working_dir = (m_workingDir + "/breez_sdk").toStdString();

        // Connect
        std::vector<uint8_t> seedVec(seedBytes.begin(), seedBytes.end());
        m_sdk = breez_sdk::SDK::connect(config, seedVec, this); 
        
        // Also set as payment listener explicitly if needed
        m_sdk->set_payment_listener(this);
        
        m_initialized = true;
        emit serviceReady(true);
        
        // Start polling
        m_pollingTimer.start();
        
        // Check if we should show backup reminder (first time setup)
        QString reminderFlagPath = m_workingDir + "/backup_reminder_shown";
        if (!QFile::exists(reminderFlagPath)) {
            // Emit signal to show backup reminder (handled by UI layer)
            emit backupReminderNeeded(m_workingDir + "/seed.dat");
        }
        
    } catch (const std::exception& e) {
        qWarning() << "Failed to initialize Breez SDK:" << e.what();
        m_lastError = e.what();
        emit errorOccurred(m_lastError);
        return false;
    // Legacy code removed.
    } // End of initialize method
    return true;

#else
  qWarning()
      << "BreezService not compiled with Breez SDK. Initialization skipped.";
  emit serviceReady(false);
  return false;
#endif
}

bool BreezService::sendLightningPayment(const QString &bolt11) {
  if (bolt11.isEmpty()) {
    qWarning() << "sendLightningPayment: empty invoice";
    QMetaObject::invokeMethod(
        this, [this]() { emit sendCompleted(false, "Empty invoice"); },
        Qt::QueuedConnection);
    return false;
  }

#ifdef HAVE_BREEZ_SDK
  if (!m_initialized || !m_sdk) {
    qWarning() << "sendLightningPayment: Breez SDK not initialized";
    QMetaObject::invokeMethod(
        this,
        [this]() { emit sendCompleted(false, "Breez SDK not initialized"); },
        Qt::QueuedConnection);
    return false;
  }

  // Run the send operation asynchronously
  (void)QtConcurrent::run([this, bolt11]() {
    try {
      // Use SDK to send payment
      breez_sdk::SendPaymentRequest req;
      req.bolt11 = bolt11.toStdString();
      auto result = m_sdk->send_payment(req);

      if (result.success) {
        QString txid = QString::fromStdString(result.payment_id);
        QMetaObject::invokeMethod(
            this, [this, txid]() { emit sendCompleted(true, txid); },
            Qt::QueuedConnection);
      } else {
        QString err = QString::fromStdString(result.error_message);
        QMetaObject::invokeMethod(
            this, [this, err]() { emit sendCompleted(false, err); },
            Qt::QueuedConnection);
      }
    } catch (const std::exception &e) {
      QString err =
          QString("Exception sending lightning payment: %1").arg(e.what());
      qWarning() << err;
      QMetaObject::invokeMethod(
          this, [this, err]() { emit sendCompleted(false, err); },
          Qt::QueuedConnection);
    }
  });

  return true;
#else
  Q_UNUSED(bolt11)
  // Not compiled with SDK — fail fast but simulate async behavior for UI
  (void)QtConcurrent::run([this]() {
    QThread::sleep(1);
    QMetaObject::invokeMethod(
        this,
        [this]() {
          emit sendCompleted(false, "Breez SDK not available in this build");
        },
        Qt::QueuedConnection);
  });
  return false;
#endif
}

bool BreezService::sendOnChain(const QString &address, qint64 amountSats,
                               const QString &network) {
  if (address.isEmpty() || amountSats <= 0) {
    qWarning() << "sendOnChain: invalid parameters";
    QMetaObject::invokeMethod(
        this,
        [this]() { emit sendCompleted(false, "Invalid address or amount"); },
        Qt::QueuedConnection);
    return false;
  }

#ifdef HAVE_BREEZ_SDK
  if (!m_initialized || !m_sdk) {
    qWarning() << "sendOnChain: Breez SDK not initialized";
    QMetaObject::invokeMethod(
        this,
        [this]() { emit sendCompleted(false, "Breez SDK not initialized"); },
        Qt::QueuedConnection);
    return false;
  }

  // Run send on-chain asynchronously
  (void)QtConcurrent::run([this, address, amountSats, network]() {
    try {
      breez_sdk::OnChainSendRequest req;
      req.address = address.toStdString();
      req.amount_sat = static_cast<uint64_t>(amountSats);
      if (network.compare("liquid", Qt::CaseInsensitive) == 0) {
        req.network = breez_sdk::Network::LIQUID;
      } else {
        req.network = breez_sdk::Network::BITCOIN;
      }

      auto result = m_sdk->send_on_chain(req);
      if (result.success) {
        QString txid = QString::fromStdString(result.txid);
        QMetaObject::invokeMethod(
            this, [this, txid]() { emit sendCompleted(true, txid); },
            Qt::QueuedConnection);
      } else {
        QString err = QString::fromStdString(result.error_message);
        QMetaObject::invokeMethod(
            this, [this, err]() { emit sendCompleted(false, err); },
            Qt::QueuedConnection);
      }
    } catch (const std::exception &e) {
      QString err = QString("Exception sending on-chain: %1").arg(e.what());
      qWarning() << err;
      QMetaObject::invokeMethod(
          this, [this, err]() { emit sendCompleted(false, err); },
          Qt::QueuedConnection);
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
    QMetaObject::invokeMethod(
        this,
        [this]() {
          emit sendCompleted(false, "Breez SDK not available in this build");
        },
        Qt::QueuedConnection);
  });
  return false;
#endif
}

QString BreezService::createInvoice(qint64 amountSats,
                                    const QString &description, int expirySec) {
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

  qDebug() << "Creating invoice for" << amountSats << "sats, expires in"
           << expirySec << "seconds";

  try {
    // Create invoice request
    CreateInvoiceRequest req;
    // Set amount to 0 to allow sender to specify the amount
    req.amount_msat = 0;
    req.description = description.toStdString();
    req.expiry = expirySec;

    // Add amount as part of the description if specified
    if (amountSats > 0) {
      req.description =
          (description + "\nSuggested amount: " + QString::number(amountSats) +
           " sats")
              .toStdString();
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

QVariantMap BreezService::fullNodeInfo() const {
  QVariantMap result;
  if (!m_initialized || !m_sdk) {
    return result;
  }

  try {
    auto info = m_sdk->node_info();
    result["id"] = QString::fromStdString(info.id);
    result["block_height"] = (uint)info.block_height;
    result["max_payable_msat"] = (qulonglong)info.max_payable_msat;
    result["max_receivable_msat"] = (qulonglong)info.max_receivable_msat;
    result["inbound_liquidity_msats"] = (qulonglong)info.inbound_liquidity_msats;
    result["channels_balance_msat"] = (qulonglong)info.channels_balance_msat;
    result["onchain_balance_msat"] = (qulonglong)info.onchain_balance_msat;
    result["connected_peers_count"] = (int)info.connected_peers.size();
  } catch (const std::exception &e) {
    qWarning() << "Error getting full node info:" << e.what();
  }

  return result;
}

qint64 BreezService::balance() const {
  if (!m_initialized || !m_sdk) {
    return 0;
  }

  try {
    auto info = m_sdk->node_info();
    return (info.onchain_balance_msat / 1000) +
           (info.channels_balance_msat / 1000);
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
      pmt["amount"] = (qulonglong)(payment.amount_msat / 1000); // Convert to satoshis
      pmt["hash"] = QString::fromStdString(payment.id); // Use ID as hash
      pmt["memo"] = QString::fromStdString(payment.description);
      pmt["timestamp"] = (qint64)payment.payment_time;
      
      QString statusStr = "unknown";
      if (payment.status == breez_sdk::PaymentStatus::COMPLETE) statusStr = "complete";
      else if (payment.status == breez_sdk::PaymentStatus::PENDING) statusStr = "pending";
      else if (payment.status == breez_sdk::PaymentStatus::FAILED) statusStr = "failed";
      pmt["status"] = statusStr;
      pmt["type"] = (payment.payment_type == breez_sdk::PaymentType::RECEIVED) ? "received" : "sent";

      result.append(pmt);
    }
  } catch (const std::exception &e) {
    qWarning() << "Failed to get payment history:" << e.what();
  }

  return result;
}

void BreezService::setupPaymentListener() {
#ifdef HAVE_BREEZ_SDK
  if (!m_sdk)
    return;

  // Listener implementation requires a class inheriting from EventListener.
  // For now, we disable this as we need to implement the wrapper.
  // m_sdk->set_payment_listener(...);
  qWarning() << "Payment listener not implemented yet";
#endif
}

void BreezService::onPaymentReceived(const InvoicePaid &payment) {
  qint64 amountSats = payment.amount_msat / 1000; // Convert to satoshis
  QString paymentHash = QString::fromStdString(payment.payment_hash);
  QString memo = QString::fromStdString(payment.description);

  qInfo() << "Payment received:" << amountSats << "sats, hash:" << paymentHash
          << "memo:" << memo;

  emit paymentReceived(amountSats, paymentHash, memo);
}

void BreezService::checkForPayments() {
  if (!m_initialized || !m_sdk) {
    qDebug() << "Skipping payment check - service not initialized";
    return;
  }

  // Rate limiting - don't check too frequently
  static QDateTime lastCheck;
  QDateTime now = QDateTime::currentDateTime();
  if (lastCheck.isValid() &&
      lastCheck.msecsTo(now) < 2000) { // 2 second minimum between checks
    return;
  }
  lastCheck = now;

  // Use QtConcurrent to fetch payments in background
  (void)QtConcurrent::run([this]() {
    try {
      // Check for new payments
      auto payments = m_sdk->list_payments(ListPaymentsRequest{});

      // Process results on main thread
      QMetaObject::invokeMethod(this, [this, payments]() {
        for (const auto &payment : payments) {
          if (payment.status == PaymentStatus::COMPLETE &&
              payment.payment_type == PaymentType::RECEIVED) {
            QString paymentId = QString::fromStdString(payment.id);
            if (!m_processedPayments.contains(paymentId)) {
              m_processedPayments.insert(paymentId);
              InvoicePaid paid;
              paid.amount_msat = payment.amount_msat;
              paid.payment_hash = payment.id; // Use ID as hash if hash not available
              paid.description = payment.description;
              onPaymentReceived(paid);
            }
          }
        }

        // Clean up old processed payments
        if (m_processedPayments.size() > 1000) {
          m_processedPayments.clear();
        }
      }, Qt::QueuedConnection);

    } catch (const std::exception &e) {
      qCritical() << QString("Error checking for payments: %1").arg(e.what());
      QMetaObject::invokeMethod(this, [this, e_what = QString::fromStdString(e.what())]() {
          m_pollingTimer.setInterval(qMin(m_pollingTimer.interval() * 2, 60000));
          emit errorOccurred(QString("Error checking for payments: %1").arg(e_what));
      }, Qt::QueuedConnection);
    }
  });
}

void BreezService::on_event(const breez_sdk::SdkEvent& e) {
    // In a real implementation, we'd deserialize the event
    // For now, we'll just trigger a check for payments when any event occurs
    // This is much better than polling every 5 seconds!
    qDebug() << "Breez SDK event received, checking for payments...";
    
    // We can't call checkForPayments directly if it's not thread-safe or if it's on a different thread
    // But checkForPayments uses m_sdk which is thread-safe.
    QMetaObject::invokeMethod(this, "checkForPayments", Qt::QueuedConnection);
}

void BreezService::retryInitialization() {
  if (m_retryCount >= 3) {
    qWarning() << "Max retry attempts reached for Breez initialization";
    emit errorOccurred("Failed to initialize after multiple attempts");
    return;
  }

  m_retryCount++;
  qInfo() << "Retrying Breez initialization, attempt" << m_retryCount;

  // Try to re-initialize with stored credentials
  initialize(m_apiKey, m_sparkUrl, m_sparkAccessKey, "bitcoin");
}
