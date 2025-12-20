#ifndef BACKUP_REMINDER_DIALOG_HPP
#define BACKUP_REMINDER_DIALOG_HPP

#include <QDialog>
#include <QString>

class QLabel;
class QPushButton;
class QCheckBox;

class BackupReminderDialog : public QDialog {
  Q_OBJECT

public:
  explicit BackupReminderDialog(const QString &seedPath, QWidget *parent = nullptr);
  ~BackupReminderDialog() override = default;

  bool dontShowAgain() const { return m_dontShowAgain; }
  bool backupCompleted() const { return m_backupCompleted; }

private slots:
  void onBackupNow();
  void onRemindLater();
  void onDontShowAgain(int state);

private:
  void setupUI();
  bool performBackup();

  QString m_seedPath;
  bool m_dontShowAgain = false;
  bool m_backupCompleted = false;

  QLabel *m_warningLabel = nullptr;
  QLabel *m_instructionsLabel = nullptr;
  QPushButton *m_backupButton = nullptr;
  QPushButton *m_laterButton = nullptr;
  QCheckBox *m_dontShowCheckbox = nullptr;
};

#endif // BACKUP_REMINDER_DIALOG_HPP
