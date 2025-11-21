#include "manage-wallet-dialog.hpp"
#include "breez-service.hpp"
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QComboBox>
#include <QLineEdit>
#include <QSpinBox>
#include <QPushButton>
#include <QLabel>
#include <QMessageBox>

ManageWalletDialog::ManageWalletDialog(QWidget *parent)
    : QDialog(parent)
{
    setWindowTitle("Manage Wallet");
    setMinimumSize(400, 180);

    methodCombo = new QComboBox(this);
    methodCombo->addItem("Lightning");
    methodCombo->addItem("On-Chain (Bitcoin)");
    methodCombo->addItem("On-Chain (Liquid)");

    destEdit = new QLineEdit(this);
    destEdit->setPlaceholderText("Bolt11 invoice or on-chain address");

    amountSpin = new QSpinBox(this);
    amountSpin->setMinimum(0);
    amountSpin->setMaximum(1000000000); // arbitrary limit in sats
    amountSpin->setSuffix(" sats");
    amountSpin->setValue(0);

    // Simulation warning when stub simulation is enabled
    QLabel *simWarning = new QLabel(this);
#ifdef BREEZ_STUB_SIMULATE
    simWarning->setText("Demo Mode: Sends are simulated — no real funds will be transferred.");
    simWarning->setStyleSheet("color: #856404; background-color: #FFF3CD; padding: 6px; border-radius: 4px; font-weight: bold;");
    simWarning->setWordWrap(true);
    simWarning->setVisible(true);
#else
    simWarning->setVisible(false);
#endif

    sendBtn = new QPushButton("Send", this);
    statusLabel = new QLabel(this);
    statusLabel->setWordWrap(true);
    statusLabel->setVisible(false);

    QVBoxLayout *main = new QVBoxLayout(this);
    main->addWidget(new QLabel("Method:", this));
    main->addWidget(methodCombo);
    main->addWidget(simWarning);
    main->addWidget(new QLabel("Destination (bolt11 or address):", this));
    main->addWidget(destEdit);
    main->addWidget(new QLabel("Amount (satoshis, for on-chain):", this));
    main->addWidget(amountSpin);

    QHBoxLayout *btnRow = new QHBoxLayout();
    btnRow->addStretch(1);
    btnRow->addWidget(sendBtn);
    main->addLayout(btnRow);
    main->addWidget(statusLabel);

    connect(sendBtn, &QPushButton::clicked, this, &ManageWalletDialog::onSendClicked);
    connect(&BreezService::instance(), &BreezService::sendCompleted, this, &ManageWalletDialog::onSendCompleted, Qt::QueuedConnection);
}

ManageWalletDialog::~ManageWalletDialog() = default;

void ManageWalletDialog::onSendClicked() {
    QString method = methodCombo->currentText();
    QString dest = destEdit->text().trimmed();
    qint64 amount = static_cast<qint64>(amountSpin->value());

    statusLabel->setVisible(false);

    if (method.startsWith("Lightning")) {
        if (dest.isEmpty()) {
            QMessageBox::warning(this, "Send", "Please enter a Lightning bolt11 invoice.");
            return;
        }
        sendBtn->setEnabled(false);
        statusLabel->setText("Sending Lightning payment...");
        statusLabel->setVisible(true);
        // Call BreezService
        BreezService::instance().sendLightningPayment(dest);
    } else {
        // On-chain
        if (dest.isEmpty() || amount <= 0) {
            QMessageBox::warning(this, "Send", "Please enter an address and a positive amount in sats.");
            return;
        }
        sendBtn->setEnabled(false);
        statusLabel->setText("Sending on-chain transaction...");
        statusLabel->setVisible(true);

        QString network = method.contains("Liquid") ? QStringLiteral("liquid") : QStringLiteral("bitcoin");
        BreezService::instance().sendOnChain(dest, amount, network);
    }
}

void ManageWalletDialog::onSendCompleted(bool ok, const QString &txid_or_err) {
    sendBtn->setEnabled(true);
    if (ok) {
        statusLabel->setText(QString("Send succeeded: %1").arg(txid_or_err));
        QMessageBox::information(this, "Send Succeeded", QString("Transaction id: %1").arg(txid_or_err));
        accept(); // close dialog on success
    } else {
        statusLabel->setText(QString("Send failed: %1").arg(txid_or_err));
        QMessageBox::critical(this, "Send Failed", txid_or_err);
    }
}
