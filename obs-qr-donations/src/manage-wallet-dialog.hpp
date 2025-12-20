#ifndef MANAGE_WALLET_DIALOG_HPP
#define MANAGE_WALLET_DIALOG_HPP

#include <QDialog>
#include <QString>
#include <QVector>

class QLabel;
class QPushButton;
class QTableWidget;

class ManageWalletDialog : public QDialog {
  Q_OBJECT

public:
  explicit ManageWalletDialog(QWidget *parent = nullptr);
  ~ManageWalletDialog() override = default;

  void updateInfo();

private slots:
  void onRefresh();
  void onSendPayment();
  void onBackupWallet();
  void onSendCompleted(bool ok, const QString &txid_or_err);

private:
  void setupUI();
  void updateNodeInfo();
  void updatePayments();

  QLabel *m_nodeIdLabel = nullptr;
  QLabel *m_balanceLabel = nullptr;
  QLabel *m_onchainBalanceLabel = nullptr;
  QLabel *m_liquidityLabel = nullptr;
  QLabel *m_statusLabel = nullptr;
  QTableWidget *m_paymentsTable = nullptr;
  QPushButton *m_refreshButton = nullptr;
  QPushButton *m_sendButton = nullptr;
  QPushButton *m_backupButton = nullptr;

};

#endif // MANAGE_WALLET_DIALOG_HPP
