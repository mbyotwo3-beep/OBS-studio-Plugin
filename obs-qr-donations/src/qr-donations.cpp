#include "qr-donations.hpp"
#include "qr-widget.hpp"
#include "donation-effect.hpp"
#include "asset-manager.hpp"
#include <QMainWindow>
#include <QMessageBox>
#include <QApplication>
#include <QDesktopWidget>

namespace QRDonations {

// Static initialization
static bool s_initialized = false;

// Source info structure
static struct obs_source_info s_sourceInfo = {
    .id = "qr_donations_source",
    .type = OBS_SOURCE_TYPE_INPUT,
    .output_flags = OBS_SOURCE_VIDEO | OBS_SOURCE_CUSTOM_DRAW | OBS_SOURCE_INTERACTION,
    .get_name = GetSourceName,
    .create = CreateSource,
    .destroy = DestroySource,
    .get_defaults = GetSourceDefaults,
    .get_properties = GetSourceProperties,
    .update = UpdateSource,
    .video_render = RenderSource,
    .get_width = GetSourceWidth,
    .get_height = GetSourceHeight,
    .icon_type = OBS_ICON_TYPE_IMAGE,
};

QRDonationsSource::QRDonationsSource(obs_data_t *settings, obs_source_t *source)
    : QObject()
    , source(source)
    , widget(nullptr)
    , showBalance(true)
    , showAssetSymbol(true)
    , enableEffects(true)
{
    // Get the main window for parenting our widget
    auto mainWindow = static_cast<QMainWindow *>(obs_frontend_get_main_window());
    
    // Create the widget
    widget = new QRDonationsWidget(mainWindow);
    
    // Create donation effect overlay
    donationEffect = std::make_unique<DonationEffect>(mainWindow);
    donationEffect->setWindowFlags(Qt::Tool | Qt::FramelessWindowHint | Qt::WindowStaysOnTopHint);
    donationEffect->setAttribute(Qt::WA_TranslucentBackground);
    donationEffect->setAttribute(Qt::WA_TransparentForMouseEvents);
    
    // Set initial size to cover the whole screen
    QRect screenGeometry = QApplication::desktop()->screenGeometry();
    donationEffect->setGeometry(screenGeometry);
    
    // Load settings
    UpdateSource(this, settings);
}

QRDonationsSource::~QRDonationsSource()
{
    if (widget) {
        widget->hide();
        widget->deleteLater();
    }
    
    if (donationEffect) {
        donationEffect->hide();
    }
}

void QRDonationsSource::update(obs_data_t *settings)
{
    currentAsset = obs_data_get_string(settings, "asset");
    currentAddress = obs_data_get_string(settings, "address");
    showBalance = obs_data_get_bool(settings, "show_balance");
    showAssetSymbol = obs_data_get_bool(settings, "show_asset_symbol");
    enableEffects = obs_data_get_bool(settings, "enable_effects");
    
    if (widget) {
        widget->setAddress(currentAsset, currentAddress);
        widget->setDisplayOptions(showBalance, showAssetSymbol);
    }
    
    // Update effect settings
    if (donationEffect) {
        // Set effect color based on currency
        QColor effectColor(255, 215, 0); // Default gold color
        if (currentAsset == "BTC") effectColor = QColor(247, 147, 26);
        else if (currentAsset == "ETH") effectColor = QColor(78, 93, 109);
        else if (currentAsset == "LTC") effectColor = QColor(191, 191, 191);
        
        donationEffect->setEffectColor(effectColor);
    }
}

void QRDonationsSource::onDonationReceived(double amount, const QString &currency) {
    if (!enableEffects || !donationEffect) {
        return;
    }
    
    // Position the effect over the widget if it's visible
    if (widget && widget->isVisible()) {
        QPoint globalPos = widget->mapToGlobal(QPoint(0, 0));
        donationEffect->setGeometry(widget->width(), widget->height(), 
                                  widget->width(), widget->height());
        donationEffect->move(globalPos);
    }
    
    // Trigger the effect
    donationEffect->triggerEffect(amount, currency);
}

void QRDonationsSource::showProperties()
{
    if (widget) {
        widget->show();
        widget->raise();
        widget->activateWindow();
    }
}

void QRDonationsSource::hideProperties()
{
    if (widget) {
        widget->hide();
    }
}

void QRDonationsSource::render(gs_effect_t *effect)
{
    if (!widget) return;
    
    // Get the widget as a QImage
    QImage image = widget->grab().toImage().convertToFormat(QImage::Format_RGBA8888);
    
    // Set up the texture
    gs_texture_t *texture = gs_texture_create(
        image.width(),
        image.height(),
        GS_RGBA,
        1,
        (const uint8_t **)&image.bits(),
        GS_DYNAMIC
    );
    
    // Draw the texture
    gs_effect_set_texture(gs_effect_get_param_by_name(effect, "image"), texture);
    gs_draw_sprite(texture, 0, 0, 0);
    
    // Clean up
    gs_texture_destroy(texture);
}

uint32_t QRDonationsSource::getWidth() const
{
    return widget ? widget->width() : 0;
}

uint32_t QRDonationsSource::getHeight() const
{
    return widget ? widget->height() : 0;
}

// Source callbacks implementation
const char *GetSourceName(void *)
{
    return "QR Donations";
}

void *CreateSource(obs_data_t *settings, obs_source_t *source)
{
    try {
        return new QRDonationsSource(settings, source);
    } catch (const std::exception &e) {
        blog(LOG_ERROR, "[QR Donations] Failed to create source: %s", e.what());
        return nullptr;
    }
}

void DestroySource(void *data)
{
    delete static_cast<QRDonationsSource *>(data);
}

void GetSourceDefaults(obs_data_t *settings)
{
    obs_data_set_default_string(settings, "asset", "BTC");
    obs_data_set_default_string(settings, "address", "");
    obs_data_set_default_bool(settings, "show_balance", true);
    obs_data_set_default_bool(settings, "show_asset_symbol", true);
    obs_data_set_default_bool(settings, "enable_effects", true);
}

obs_properties_t *GetSourceProperties(void *data)
{
    auto *source = static_cast<QRDonationsSource *>(data);
    auto *props = obs_properties_create();
    
    // Asset selection
    auto *assetList = obs_properties_add_list(
        props,
        "asset",
        "Cryptocurrency",
        OBS_COMBO_TYPE_LIST,
        OBS_COMBO_FORMAT_STRING
    );
    
    // Add supported assets
    auto assets = AssetManager::instance().getSupportedAssets();
    for (const auto &asset : assets) {
        obs_property_list_add_string(assetList, asset.name.c_str(), asset.id.c_str());
    }
    
    // Address input
    obs_properties_add_text(
        props,
        "address",
        "Wallet Address",
        OBS_TEXT_DEFAULT
    );
    
    // Display options
    obs_properties_add_bool(
        props,
        "show_balance",
        "Show Balance"
    );
    
    obs_properties_add_bool(
        props,
        "show_asset_symbol",
        "Show Asset Symbol"
    );
    
    // Effect options
    obs_properties_add_bool(
        props,
        "enable_effects",
        "Enable Visual Effects on Donation"
    );
    
    return props;
}

void UpdateSource(void *data, obs_data_t *settings)
{
    auto *source = static_cast<QRDonationsSource *>(data);
    if (source) {
        source->update(settings);
    }
}

void RenderSource(void *data, gs_effect_t *effect)
{
    auto *source = static_cast<QRDonationsSource *>(data);
    if (source) {
        source->render(effect);
    }
}

uint32_t GetSourceWidth(void *data)
{
    auto *source = static_cast<QRDonationsSource *>(data);
    return source ? source->getWidth() : 0;
}

uint32_t GetSourceHeight(void *data)
{
    auto *source = static_cast<QRDonationsSource *>(data);
    return source ? source->getHeight() : 0;
}

// Plugin initialization
void InitializeSource()
{
    if (s_initialized) return;
    
    // Initialize asset manager
    AssetManager::instance().initialize();
    
    // Register the source
    obs_register_source(&s_sourceInfo);
    
    s_initialized = true;
    blog(LOG_INFO, "[QR Donations] Source initialized");
}

} // namespace QRDonations
