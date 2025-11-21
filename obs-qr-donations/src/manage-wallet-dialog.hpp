#pragma once

#include <QDialog>
#include <QString>

class QComboBox;
class QLineEdit;
class QSpinBox;
class QPushButton;
class QLabel;

class ManageWalletDialog : public QDialog {
    Q_OBJECT
public:
    explicit ManageWalletDialog(QWidget *parent = nullptr);
    ~ManageWalletDialog() override;

private slots:
    void onSendClicked();
    void onSendCompleted(bool ok, const QString &txid_or_err);

private:
    QComboBox *methodCombo = nullptr; // "Lightning" or "On-Chain"
    QLineEdit *destEdit = nullptr;    // bolt11 or address
    QSpinBox *amountSpin = nullptr;   // sats (for on-chain)
    QPushButton *sendBtn = nullptr;
    QLabel *statusLabel = nullptr;
};
