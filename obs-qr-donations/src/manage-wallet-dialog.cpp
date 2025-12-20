#include "manage-wallet-dialog.hpp"
#include "breez-service.hpp"
#include "send-payment-dialog.hpp"
#include <QLabel>
#include <QPushButton>
#include <QTableWidget>
#include <QHeaderView>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QDateTime>
#include <QMessageBox>
#include <QApplication>
#include <QClipboard>

#include <QFileDialog>
#include <QFile>
#include <QStandardPaths>

ManageWalletDialog::ManageWalletDialog(QWidget *parent)
    : QDialog(parent) {
  setupUI();
  setWindowTitle("Lightning Wallet Manager");
  setMinimumSize(600, 550);
  
  // Connect to BreezService signals
  connect(&BreezService::instance(), &BreezService::sendCompleted, this, &ManageWalletDialog::onSendCompleted);
  connect(&BreezService::instance(), &BreezService::paymentReceived, this, [this](){ updateInfo(); });
  
  updateInfo();
}

void ManageWalletDialog::setupUI() {
  auto *mainLayout = new QVBoxLayout(this);
  mainLayout->setSpacing(15);

  // Header
  auto *headerLabel = new QLabel("<h2>⚡ Wallet Management</h2>", this);
  headerLabel->setAlignment(Qt::AlignCenter);
  mainLayout->addWidget(headerLabel);

  // Node Info Group
  auto *infoLayout = new QVBoxLayout();
  infoLayout->setSpacing(5);

  m_nodeIdLabel = new QLabel("<b>Node ID:</b> -", this);
  m_nodeIdLabel->setTextInteractionFlags(Qt::TextSelectableByMouse);
  m_nodeIdLabel->setWordWrap(true);
  infoLayout->addWidget(m_nodeIdLabel);

  m_balanceLabel = new QLabel("<b>Lightning Balance:</b> - sats", this);
  m_balanceLabel->setStyleSheet("font-size: 16px; color: #4CAF50;");
  infoLayout->addWidget(m_balanceLabel);

  m_onchainBalanceLabel = new QLabel("<b>On-chain Balance:</b> - sats", this);
  infoLayout->addWidget(m_onchainBalanceLabel);

  m_liquidityLabel = new QLabel("<b>Inbound Liquidity:</b> - sats", this);
  infoLayout->addWidget(m_liquidityLabel);

  mainLayout->addLayout(infoLayout);

  // Status Label
  m_statusLabel = new QLabel(this);
  m_statusLabel->setWordWrap(true);
  m_statusLabel->setStyleSheet("color: #2196F3; font-weight: bold;");
  m_statusLabel->setVisible(false);
  mainLayout->addWidget(m_statusLabel);

  // Actions Row
  auto *actionLayout = new QHBoxLayout();
  
  m_sendButton = new QPushButton("💸 Send Payment", this);
  m_sendButton->setStyleSheet("padding: 8px; font-weight: bold;");
  connect(m_sendButton, &QPushButton::clicked, this, &ManageWalletDialog::onSendPayment);
  actionLayout->addWidget(m_sendButton);

  m_backupButton = new QPushButton("💾 Backup Seed", this);
  m_backupButton->setStyleSheet("padding: 8px;");
  connect(m_backupButton, &QPushButton::clicked, this, &ManageWalletDialog::onBackupWallet);
  actionLayout->addWidget(m_backupButton);

  m_refreshButton = new QPushButton("🔄 Refresh", this);
  m_refreshButton->setStyleSheet("padding: 8px;");
  connect(m_refreshButton, &QPushButton::clicked, this, &ManageWalletDialog::onRefresh);
  actionLayout->addWidget(m_refreshButton);

  mainLayout->addLayout(actionLayout);

  // Payments Table
  mainLayout->addWidget(new QLabel("<b>Recent Payments:</b>", this));
  
  m_paymentsTable = new QTableWidget(0, 4, this);
  m_paymentsTable->setHorizontalHeaderLabels({"Date", "Type", "Amount", "Status"});
  m_paymentsTable->horizontalHeader()->setSectionResizeMode(QHeaderView::Stretch);
  m_paymentsTable->setEditTriggers(QAbstractItemView::NoEditTriggers);
  m_paymentsTable->setSelectionBehavior(QAbstractItemView::SelectRows);
  mainLayout->addWidget(m_paymentsTable);

  // Close Button
  auto *closeButton = new QPushButton("Close", this);
  connect(closeButton, &QPushButton::clicked, this, &QDialog::accept);
  mainLayout->addWidget(closeButton, 0, Qt::AlignRight);
}

void ManageWalletDialog::updateInfo() {
  updateNodeInfo();
  updatePayments();
}

void ManageWalletDialog::updateNodeInfo() {
  QVariantMap info = BreezService::instance().fullNodeInfo();
  
  if (info.isEmpty()) {
    m_nodeIdLabel->setText("<b>Node ID:</b> (Not Initialized)");
    return;
  }

  QString nodeId = info["id"].toString();
  m_nodeIdLabel->setText(QString("<b>Node ID:</b> %1").arg(nodeId));
  
  qint64 lightningBalance = info["channels_balance_msat"].toLongLong() / 1000;
  m_balanceLabel->setText(QString("<b>Lightning Balance:</b> %1 sats").arg(lightningBalance));
  
  qint64 onchainBalance = info["onchain_balance_msat"].toLongLong() / 1000;
  m_onchainBalanceLabel->setText(QString("<b>On-chain Balance:</b> %1 sats").arg(onchainBalance));
  
  qint64 inboundLiquidity = info["inbound_liquidity_msats"].toLongLong() / 1000;
  m_liquidityLabel->setText(QString("<b>Inbound Liquidity:</b> %1 sats").arg(inboundLiquidity));
}

void ManageWalletDialog::updatePayments() {
  QVariantList history = BreezService::instance().paymentHistory();
  m_paymentsTable->setRowCount(0);
  
  for (const QVariant &v : history) {
    QVariantMap payment = v.toMap();
    int row = m_paymentsTable->rowCount();
    m_paymentsTable->insertRow(row);
    
    qint64 timestamp = payment["timestamp"].toLongLong();
    QDateTime dt = QDateTime::fromSecsSinceEpoch(timestamp);
    QString type = payment["type"].toString();
    qint64 amountSats = payment["amount"].toLongLong();
    QString status = payment["status"].toString();

    m_paymentsTable->setItem(row, 0, new QTableWidgetItem(dt.toString("yyyy-MM-dd HH:mm")));
    m_paymentsTable->setItem(row, 1, new QTableWidgetItem(type == "received" ? "📥 Received" : "📤 Sent"));
    m_paymentsTable->setItem(row, 2, new QTableWidgetItem(QString("%1 sats").arg(amountSats)));
    m_paymentsTable->setItem(row, 3, new QTableWidgetItem(status));
    
    if (type == "received") {
      m_paymentsTable->item(row, 1)->setForeground(Qt::green);
    } else {
      m_paymentsTable->item(row, 1)->setForeground(Qt::red);
    }
  }
}

void ManageWalletDialog::onRefresh() {
  updateInfo();
}

void ManageWalletDialog::onSendPayment() {
  SendPaymentDialog dialog(this);
  dialog.setBalance(BreezService::instance().balance());
  if (dialog.exec() == QDialog::Accepted) {
    QString invoice = dialog.getInvoice();
    if (BreezService::instance().sendLightningPayment(invoice)) {
        m_statusLabel->setText("⏳ Sending payment...");
        m_statusLabel->setStyleSheet("color: #2196F3; font-weight: bold;");
        m_statusLabel->setVisible(true);
        m_sendButton->setEnabled(false);
    }
  }
}

void ManageWalletDialog::onSendCompleted(bool ok, const QString &txid_or_err) {
    m_sendButton->setEnabled(true);
    if (ok) {
        m_statusLabel->setText("✅ Payment sent successfully!");
        m_statusLabel->setStyleSheet("color: #4CAF50; font-weight: bold;");
        updateInfo();
    } else {
        m_statusLabel->setText(QString("❌ Payment failed: %1").arg(txid_or_err));
        m_statusLabel->setStyleSheet("color: #F44336; font-weight: bold;");
    }
    m_statusLabel->setVisible(true);
}

void ManageWalletDialog::onBackupWallet() {
  QString seedPath = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation) + "/breez/seed.dat";
  
  QString defaultName = QString("lightning-wallet-backup-%1.dat")
                            .arg(QDateTime::currentDateTime().toString("yyyyMMdd-HHmmss"));
  QString defaultPath = QStandardPaths::writableLocation(QStandardPaths::DocumentsLocation) + "/" + defaultName;

  QString backupPath = QFileDialog::getSaveFileName(this, "Save Wallet Backup", defaultPath, "Wallet Backup (*.dat)");

  if (backupPath.isEmpty()) return;

  if (QFile::copy(seedPath, backupPath)) {
      QFile::setPermissions(backupPath, QFile::ReadOwner | QFile::WriteOwner);
      QMessageBox::information(this, "Backup Successful", "Wallet backup saved to:\n" + backupPath);
  } else {
      QMessageBox::critical(this, "Backup Failed", "Failed to copy seed file. Ensure the wallet is initialized.");
  }
}
