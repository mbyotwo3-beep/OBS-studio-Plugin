#include "backup-reminder-dialog.hpp"
#include <QCheckBox>
#include <QFileDialog>
#include <QLabel>
#include <QMessageBox>
#include <QPushButton>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QFile>
#include <QStandardPaths>
#include <QDateTime>

BackupReminderDialog::BackupReminderDialog(const QString &seedPath, QWidget *parent)
    : QDialog(parent), m_seedPath(seedPath) {
  setupUI();
  setWindowTitle("⚠️ CRITICAL: Backup Your Lightning Wallet");
  setModal(true);
  setMinimumWidth(500);
}

void BackupReminderDialog::setupUI() {
  auto *mainLayout = new QVBoxLayout(this);
  mainLayout->setSpacing(15);

  // Warning icon and title
  m_warningLabel = new QLabel(this);
  m_warningLabel->setText(
      "<h2 style='color: #ff5722;'>⚠️ YOUR WALLET NEEDS TO BE BACKED UP!</h2>");
  m_warningLabel->setWordWrap(true);
  m_warningLabel->setAlignment(Qt::AlignCenter);
  mainLayout->addWidget(m_warningLabel);

  // Critical warning message
  auto *criticalLabel = new QLabel(this);
  criticalLabel->setText(
      "<p style='font-size: 14px; font-weight: bold; color: #d32f2f;'>"
      "If you lose your wallet seed file, you will PERMANENTLY LOSE ACCESS to "
      "all your funds!</p>");
  criticalLabel->setWordWrap(true);
  criticalLabel->setAlignment(Qt::AlignCenter);
  mainLayout->addWidget(criticalLabel);

  // Instructions
  m_instructionsLabel = new QLabel(this);
  m_instructionsLabel->setText(
      "<p style='font-size: 12px;'>"
      "<b>What is the seed file?</b><br>"
      "Your wallet is controlled by a 32-byte seed file stored on your "
      "computer. "
      "This file IS your wallet - it contains all your funds and payment "
      "history.</p>"
      "<p style='font-size: 12px;'>"
      "<b>Why backup?</b><br>"
      "• OS reinstall/crash → Seed deleted → Funds lost<br>"
      "• Hard drive failure → Seed lost → Funds lost<br>"
      "• Accidental deletion → Seed gone → Funds lost</p>"
      "<p style='font-size: 12px;'>"
      "<b>How to backup:</b><br>"
      "1. Click \"Backup Now\" below<br>"
      "2. Save to USB drive or cloud storage (encrypted)<br>"
      "3. Store in multiple safe locations</p>");
  m_instructionsLabel->setWordWrap(true);
  mainLayout->addWidget(m_instructionsLabel);

  // Seed file location
  auto *locationLabel = new QLabel(this);
  locationLabel->setText(QString("<p style='font-size: 11px; color: #666;'>"
                                 "<b>Seed location:</b> %1</p>")
                             .arg(m_seedPath));
  locationLabel->setWordWrap(true);
  mainLayout->addWidget(locationLabel);

  // Buttons
  auto *buttonLayout = new QHBoxLayout();
  buttonLayout->setSpacing(10);

  m_backupButton = new QPushButton("🔐 Backup Now (Recommended)", this);
  m_backupButton->setStyleSheet(
      "QPushButton { background-color: #4CAF50; color: white; font-weight: "
      "bold; padding: 10px; }"
      "QPushButton:hover { background-color: #45a049; }");
  connect(m_backupButton, &QPushButton::clicked, this,
          &BackupReminderDialog::onBackupNow);

  m_laterButton = new QPushButton("⏰ Remind Me Later", this);
  m_laterButton->setStyleSheet("QPushButton { padding: 10px; }");
  connect(m_laterButton, &QPushButton::clicked, this,
          &BackupReminderDialog::onRemindLater);

  buttonLayout->addWidget(m_backupButton);
  buttonLayout->addWidget(m_laterButton);
  mainLayout->addLayout(buttonLayout);

  // Don't show again checkbox
  m_dontShowCheckbox = new QCheckBox("Don't show this reminder again (NOT recommended)", this);
  m_dontShowCheckbox->setStyleSheet("QCheckBox { color: #999; font-size: 10px; }");
  connect(m_dontShowCheckbox, &QCheckBox::stateChanged, this,
          &BackupReminderDialog::onDontShowAgain);
  mainLayout->addWidget(m_dontShowCheckbox);
}

void BackupReminderDialog::onBackupNow() {
  if (performBackup()) {
    m_backupCompleted = true;
    QMessageBox::information(
        this, "Backup Successful",
        "Your wallet seed has been backed up successfully!\n\n"
        "IMPORTANT: Store this backup in a safe location:\n"
        "• USB drive in a safe\n"
        "• Encrypted cloud storage\n"
        "• Multiple physical locations\n\n"
        "Never share this file with anyone!");
    accept();
  }
}

void BackupReminderDialog::onRemindLater() {
  QMessageBox::warning(
      this, "Reminder",
      "Please backup your wallet as soon as possible!\n\n"
      "You can backup anytime by copying the seed file:\n" +
          m_seedPath);
  reject();
}

void BackupReminderDialog::onDontShowAgain(int state) {
  m_dontShowAgain = (state == Qt::Checked);
  
  if (m_dontShowAgain) {
    QMessageBox::warning(
        this, "Warning",
        "Are you sure you want to disable this reminder?\n\n"
        "If you lose your seed file without a backup, "
        "your funds will be PERMANENTLY LOST!\n\n"
        "We strongly recommend keeping this reminder enabled.");
  }
}

bool BackupReminderDialog::performBackup() {
  // Generate default backup filename with timestamp
  QString defaultName = QString("lightning-wallet-backup-%1.dat")
                            .arg(QDateTime::currentDateTime().toString(
                                "yyyyMMdd-HHmmss"));

  QString documentsPath =
      QStandardPaths::writableLocation(QStandardPaths::DocumentsLocation);
  QString defaultPath = documentsPath + "/" + defaultName;

  // Open file dialog
  QString backupPath = QFileDialog::getSaveFileName(
      this, "Save Wallet Backup", defaultPath,
      "Wallet Backup (*.dat);;All Files (*)");

  if (backupPath.isEmpty()) {
    return false; // User cancelled
  }

  // Copy seed file to backup location
  QFile seedFile(m_seedPath);
  if (!seedFile.exists()) {
    QMessageBox::critical(this, "Error",
                          "Seed file not found!\n\nPath: " + m_seedPath);
    return false;
  }

  // Remove existing backup if it exists
  if (QFile::exists(backupPath)) {
    QFile::remove(backupPath);
  }

  // Perform copy
  if (!seedFile.copy(backupPath)) {
    QMessageBox::critical(
        this, "Backup Failed",
        "Failed to copy seed file to backup location.\n\nError: " +
            seedFile.errorString());
    return false;
  }

  // Set restrictive permissions on backup (Unix-like systems)
#ifndef _WIN32
  QFile::setPermissions(backupPath, QFile::ReadOwner | QFile::WriteOwner);
#endif

  return true;
}
