#ifndef SEND_PAYMENT_DIALOG_HPP
#define SEND_PAYMENT_DIALOG_HPP

#include <QDialog>
#include <QString>

class QLabel;
class QLineEdit;
class QPushButton;
class QTextEdit;

class SendPaymentDialog : public QDialog {
  Q_OBJECT

public:
  explicit SendPaymentDialog(QWidget *parent = nullptr);
  ~SendPaymentDialog() override = default;

  void setBalance(qint64 balanceSats);
  QString getInvoice() const;

private slots:
  void onInvoiceChanged();
  void onSendClicked();
  void onCancelClicked();

private:
  void setupUI();
  bool validateInvoice(const QString &invoice);
  void parseInvoice(const QString &invoice);

  QTextEdit *m_invoiceInput = nullptr;
  QLabel *m_balanceLabel = nullptr;
  QLabel *m_amountLabel = nullptr;
  QLabel *m_descriptionLabel = nullptr;
  QLabel *m_statusLabel = nullptr;
  QPushButton *m_sendButton = nullptr;
  QPushButton *m_cancelButton = nullptr;

  QString m_invoice;
  qint64 m_amountMsat = 0;
  qint64 m_currentBalanceSats = 0;
  QString m_description;
  bool m_validInvoice = false;
};

#endif // SEND_PAYMENT_DIALOG_HPP
