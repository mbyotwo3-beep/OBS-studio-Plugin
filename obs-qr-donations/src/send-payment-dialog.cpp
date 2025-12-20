#include "send-payment-dialog.hpp"
#include <QLabel>
#include <QLineEdit>
#include <QMessageBox>
#include <QPushButton>
#include <QTextEdit>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QRegularExpression>

SendPaymentDialog::SendPaymentDialog(QWidget *parent)
    : QDialog(parent) {
  setupUI();
  setWindowTitle("Send Lightning Payment");
  setModal(true);
  setMinimumWidth(500);
  setMinimumHeight(400);
}

void SendPaymentDialog::setBalance(qint64 balanceSats) {
  m_currentBalanceSats = balanceSats;
  if (m_balanceLabel) {
    m_balanceLabel->setText(QString("<b>Current Balance:</b> %1 sats").arg(balanceSats));
  }
}

void SendPaymentDialog::setupUI() {
  auto *mainLayout = new QVBoxLayout(this);
  mainLayout->setSpacing(15);

  // Title
  auto *titleLabel = new QLabel(this);
  titleLabel->setText("<h2>⚡ Send Lightning Payment</h2>");
  titleLabel->setAlignment(Qt::AlignCenter);
  mainLayout->addWidget(titleLabel);

  // Balance display
  m_balanceLabel = new QLabel(this);
  m_balanceLabel->setText("<b>Current Balance:</b> - sats");
  m_balanceLabel->setStyleSheet("font-size: 14px; color: #4CAF50;");
  m_balanceLabel->setAlignment(Qt::AlignCenter);
  mainLayout->addWidget(m_balanceLabel);

  // Invoice input
  auto *invoiceLabel = new QLabel("Lightning Invoice (BOLT11):", this);
  mainLayout->addWidget(invoiceLabel);

  m_invoiceInput = new QTextEdit(this);
  m_invoiceInput->setPlaceholderText("Paste Lightning invoice here (lnbc...)");
  m_invoiceInput->setMaximumHeight(100);
  connect(m_invoiceInput, &QTextEdit::textChanged, this,
          &SendPaymentDialog::onInvoiceChanged);
  mainLayout->addWidget(m_invoiceInput);

  // Invoice details section
  auto *detailsLabel = new QLabel("<b>Payment Details:</b>", this);
  mainLayout->addWidget(detailsLabel);

  m_amountLabel = new QLabel("Amount: -", this);
  m_amountLabel->setStyleSheet("font-size: 14px; color: #333;");
  mainLayout->addWidget(m_amountLabel);

  m_descriptionLabel = new QLabel("Description: -", this);
  m_descriptionLabel->setStyleSheet("font-size: 12px; color: #666;");
  m_descriptionLabel->setWordWrap(true);
  mainLayout->addWidget(m_descriptionLabel);

  // Status label
  m_statusLabel = new QLabel(this);
  m_statusLabel->setWordWrap(true);
  m_statusLabel->setVisible(false);
  mainLayout->addWidget(m_statusLabel);

  // Warning
  auto *warningLabel = new QLabel(this);
  warningLabel->setText(
      "<p style='color: #ff9800; font-size: 11px;'>"
      "⚠️ <b>Warning:</b> Lightning payments are instant and irreversible. "
      "Please verify the invoice details before sending.</p>");
  warningLabel->setWordWrap(true);
  mainLayout->addWidget(warningLabel);

  // Spacer
  mainLayout->addStretch();

  // Buttons
  auto *buttonLayout = new QHBoxLayout();
  buttonLayout->setSpacing(10);

  m_cancelButton = new QPushButton("Cancel", this);
  connect(m_cancelButton, &QPushButton::clicked, this,
          &SendPaymentDialog::onCancelClicked);

  m_sendButton = new QPushButton("💸 Send Payment", this);
  m_sendButton->setEnabled(false);
  m_sendButton->setStyleSheet(
      "QPushButton { background-color: #4CAF50; color: white; font-weight: "
      "bold; padding: 10px; }"
      "QPushButton:hover { background-color: #45a049; }"
      "QPushButton:disabled { background-color: #ccc; }");
  connect(m_sendButton, &QPushButton::clicked, this,
          &SendPaymentDialog::onSendClicked);

  buttonLayout->addWidget(m_cancelButton);
  buttonLayout->addWidget(m_sendButton);
  mainLayout->addLayout(buttonLayout);
}

void SendPaymentDialog::onInvoiceChanged() {
  QString invoice = m_invoiceInput->toPlainText().trimmed();
  
  if (invoice.isEmpty()) {
    m_amountLabel->setText("Amount: -");
    m_descriptionLabel->setText("Description: -");
    m_statusLabel->setVisible(false);
    m_sendButton->setEnabled(false);
    m_validInvoice = false;
    return;
  }

  if (validateInvoice(invoice)) {
    parseInvoice(invoice);
    m_invoice = invoice;
    m_validInvoice = true;
    
    // Check balance
    qint64 amountSats = m_amountMsat / 1000;
    if (amountSats > 0 && amountSats > m_currentBalanceSats) {
        m_statusLabel->setText(QString("❌ Insufficient balance (%1 sats needed)").arg(amountSats));
        m_statusLabel->setStyleSheet("color: #f44336;");
        m_sendButton->setEnabled(false);
    } else {
        m_statusLabel->setText("✓ Valid Lightning invoice");
        m_statusLabel->setStyleSheet("color: #4CAF50;");
        m_sendButton->setEnabled(true);
    }
    m_statusLabel->setVisible(true);
  } else {
    m_amountLabel->setText("Amount: -");
    m_descriptionLabel->setText("Description: -");
    m_statusLabel->setText("✗ Invalid Lightning invoice");
    m_statusLabel->setStyleSheet("color: #f44336;");
    m_statusLabel->setVisible(true);
    m_sendButton->setEnabled(false);
    m_validInvoice = false;
  }
}

bool SendPaymentDialog::validateInvoice(const QString &invoice) {
  // Basic BOLT11 validation
  // Lightning invoices start with "lnbc" (mainnet) or "lntb" (testnet)
  if (!invoice.startsWith("lnbc") && !invoice.startsWith("lntb") &&
      !invoice.startsWith("lnbcrt")) {
    return false;
  }

  // Minimum length check
  if (invoice.length() < 20) {
    return false;
  }

  return true;
}

void SendPaymentDialog::parseInvoice(const QString &invoice) {
  // Basic parsing - in production, use proper BOLT11 decoder
  // For now, just show that it's a valid invoice
  
  // Try to extract amount from invoice string
  // BOLT11 format: lnbc[amount][multiplier]...
  QRegularExpression amountRegex("lnbc(\\d+)([munp]?)");
  QRegularExpressionMatch match = amountRegex.match(invoice);
  
  if (match.hasMatch()) {
    QString amountStr = match.captured(1);
    QString multiplier = match.captured(2);
    
    qint64 amount = amountStr.toLongLong();
    
    // Convert based on multiplier
    if (multiplier == "m") {
      amount *= 100000; // milli-bitcoin to msat
    } else if (multiplier == "u") {
      amount *= 100; // micro-bitcoin to msat
    } else if (multiplier == "n") {
      amount *= 0.1; // nano-bitcoin to msat
    } else if (multiplier == "p") {
      amount *= 0.0001; // pico-bitcoin to msat
    } else {
      amount *= 100000000; // bitcoin to msat
    }
    
    m_amountMsat = amount;
    qint64 sats = amount / 1000;
    m_amountLabel->setText(QString("Amount: %1 sats").arg(sats));
  } else {
    m_amountLabel->setText("Amount: Any (receiver sets amount)");
    m_amountMsat = 0;
  }
  
  // Description would be extracted from invoice in production
  m_descriptionLabel->setText("Description: (encoded in invoice)");
}

void SendPaymentDialog::onSendClicked() {
  if (!m_validInvoice) {
    QMessageBox::warning(this, "Invalid Invoice",
                         "Please enter a valid Lightning invoice.");
    return;
  }

  // Confirmation dialog
  QString confirmMsg = QString(
      "Are you sure you want to send this payment?\n\n"
      "Amount: %1\n"
      "Invoice: %2...\n\n"
      "This action cannot be undone!")
      .arg(m_amountMsat > 0 ? QString("%1 sats").arg(m_amountMsat / 1000) : "Any amount")
      .arg(m_invoice.left(30));

  auto reply = QMessageBox::question(
      this, "Confirm Payment", confirmMsg,
      QMessageBox::Yes | QMessageBox::No, QMessageBox::No);

  if (reply == QMessageBox::Yes) {
    accept(); // Close dialog and return Accepted
  }
}

void SendPaymentDialog::onCancelClicked() {
  reject();
}

QString SendPaymentDialog::getInvoice() const {
  return m_invoice;
}
