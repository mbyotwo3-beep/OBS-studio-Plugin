#pragma once

#include <obs-module.h>
#include <obs-frontend-api.h>
#include <QWidget>
#include <string>
#include <memory>

// Forward declarations
class QRDonationsWidget;
class DonationEffect;

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
    
public slots:
    void onDonationReceived(double amount, const QString &currency);

private:
    obs_source_t *source;
    QRDonationsWidget *widget;
    std::unique_ptr<DonationEffect> donationEffect;
    std::string currentAsset;
    std::string currentAddress;
    bool showBalance;
    bool showAssetSymbol;
    bool enableEffects;
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
