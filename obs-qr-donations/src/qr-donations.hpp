#pragma once

#include <QSoundEffect>
#include <QWidget>
#include <memory>
#include <obs-frontend-api.h>
#include <obs-module.h>
#include <string>

// Forward declarations
class QRDonationsWidget;
// Donation visual effects have been removed to keep the plugin lightweight.

namespace QRDonations {

class QRDonationsSource : public QObject {
  Q_OBJECT

public:
  explicit QRDonationsSource(obs_data_t *settings, obs_source_t *source);
  ~QRDonationsSource() override;

  void update(obs_data_t *settings);
  void showProperties();
  void hideProperties();
  void render(gs_effect_t *effect);
  uint32_t getWidth() const;
  uint32_t getHeight() const;
  QRDonationsWidget *getWidget() const { return widget; }

  QString getApiKey() const { return apiKey; }
  QString getSparkUrl() const { return sparkUrl; }
  QString getSparkKey() const { return sparkKey; }

public slots:
  void onDonationReceived(double amount, const QString &currency);

private:
  obs_source_t *source;
  QRDonationsWidget *widget;
  // No visual effect; kept for backwards compatibility in settings
  // (previously used for particle animations on donation)
  std::string currentAsset;
  std::string currentAddress;
  bool showBalance;
  bool showAssetSymbol;

  // Audio feedback
  bool enableSound;
  QString soundFilePath;
  QSoundEffect *soundEffect;

  // Breez settings
  QString apiKey;
  QString sparkUrl;
  QString sparkKey;
};

// Source callbacks
static const char *GetSourceName(void *unused);
static void *CreateSource(obs_data_t *settings, obs_source_t *source);
static void DestroySource(void *data);
static void GetSourceDefaults(obs_data_t *settings);
static obs_properties_t *GetSourceProperties(void *data);
static void UpdateSource(void *data, obs_data_t *settings);
static void RenderSource(void *data, gs_effect_t *effect);
static uint32_t GetSourceWidth(void *data);
static uint32_t GetSourceHeight(void *data);

// Plugin initialization
void InitializeSource();

} // namespace QRDonations
