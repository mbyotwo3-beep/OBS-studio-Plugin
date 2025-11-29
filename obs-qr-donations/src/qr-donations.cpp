#include "qr-donations.hpp"
#include "qr-widget.hpp"
#include "asset-manager.hpp"
#include "breez-service.hpp"
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
    , enableSound(false)
    , soundEffect(new QSoundEffect(this))
{
    // Get the main window for parenting our widget
    auto mainWindow = static_cast<QMainWindow *>(obs_frontend_get_main_window());
    
    // Create the widget
    widget = new QRDonationsWidget(mainWindow);
    
    // Load settings
    UpdateSource(this, settings);
}

QRDonationsSource::~QRDonationsSource()
{
    if (widget) {
        widget->hide();
        widget->deleteLater();
    }
}

void QRDonationsSource::update(obs_data_t *settings)
{
    currentAsset = obs_data_get_string(settings, "asset");
    QString btcAddr = obs_data_get_string(settings, "bitcoin_address");
    QString liquidAddr = obs_data_get_string(settings, "liquid_address");

    // Map selected asset to the corresponding on-chain address
    if (QString::fromStdString(currentAsset).compare("L-BTC", Qt::CaseInsensitive) == 0) {
        currentAddress = liquidAddr.toStdString();
    } else {
        currentAddress = btcAddr.toStdString();
    }
    showBalance = obs_data_get_bool(settings, "show_balance");
    showAssetSymbol = obs_data_get_bool(settings, "show_asset_symbol");
    
    if (widget) {
        widget->setAddress(currentAsset, currentAddress);
        // Set both on-chain addresses (if any) so tabs show appropriate values
        widget->setBitcoinAddress(btcAddr.toStdString());
        widget->setLiquidAddress(liquidAddr.toStdString());
        widget->setDisplayOptions(showBalance, showAssetSymbol);
    }
    
    // Update effect settings (deprecated, kept for backward compatibility)
    
    // Audio settings
    enableSound = obs_data_get_bool(settings, "enable_sound");
    QString newSoundFile = obs_data_get_string(settings, "sound_file");
    
    if (soundFilePath != newSoundFile) {
        soundFilePath = newSoundFile;
        if (!soundFilePath.isEmpty()) {
            soundEffect->setSource(QUrl::fromLocalFile(soundFilePath));
            soundEffect->setVolume(1.0f);
        }
    }
}

void QRDonationsSource::onDonationReceived(double amount, const QString &currency) {
    // Show notification in widget
    if (widget) {
        qint64 amountSats = static_cast<qint64>(amount * 100000000.0);
        widget->onPaymentReceived(amountSats, QString(), currency);
    }
    
    // Play sound if enabled
    if (enableSound && soundEffect && !soundFilePath.isEmpty()) {
        soundEffect->play();
    }
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
    obs_data_set_default_string(settings, "bitcoin_address", "");
    obs_data_set_default_string(settings, "liquid_address", "");
    obs_data_set_default_string(settings, "breez_test_status", "");
    obs_data_set_default_bool(settings, "show_balance", true);
    obs_data_set_default_bool(settings, "show_asset_symbol", true);
    obs_data_set_default_bool(settings, "enable_sound", false);
    obs_data_set_default_string(settings, "sound_file", "");
}

static bool TestBreezConnection(obs_properties_t *props, obs_property_t *property, obs_data_t *settings);

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
        "bitcoin_address",
        "Bitcoin (on-chain) Address",
        OBS_TEXT_DEFAULT
    );

    obs_properties_add_text(
        props,
        "liquid_address",
        "Liquid (on-chain) Address",
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
    
    // Audio options
    obs_properties_add_bool(
        props,
        "enable_sound",
        "Enable Sound Notification"
    );
    
    obs_properties_add_path(
        props,
        "sound_file",
        "Sound File",
        OBS_PATH_FILE,
        "Audio Files (*.wav *.mp3 *.ogg)",
        nullptr
    );

    // Breez / Spark (Lightning) settings
    obs_properties_add_bool(
        props,
        "enable_lightning",
        "Enable Lightning (Breez Spark)"
    );

    obs_properties_add_text(
        props,
        "breez_api_key",
        "Breez API Key",
        OBS_TEXT_DEFAULT
    );

    obs_properties_add_text(
        props,
        "spark_url",
        "Spark Wallet URL",
        OBS_TEXT_DEFAULT
    );

    obs_properties_add_text(
        props,
        "spark_access_key",
        "Spark Access Key",
        OBS_TEXT_DEFAULT
    );

    // Button to test Breez + Spark connection
    obs_properties_add_button(
        props,
        "test_breez_connection",
        "Test Breez Connection",
        TestBreezConnection
    );

    // Read-only field to show the result of the last Breez connection test
    obs_properties_add_text(
        props,
        "breez_test_status",
        "Breez Test Status",
        OBS_TEXT_DEFAULT
    );

    // Read-only status shown after running "Test Breez Connection"
    obs_properties_add_text(
        props,
        "breez_test_status",
        "Breez Test Status",
        OBS_TEXT_DEFAULT
    );
    
    return props;
}

void UpdateSource(void *data, obs_data_t *settings)
{
    auto *source = static_cast<QRDonationsSource *>(data);
    if (source) {
        source->update(settings);
    }

    // Handle Breez initialization when enable_lightning toggled
    bool enableLightning = obs_data_get_bool(settings, "enable_lightning");
    QString apiKey = obs_data_get_string(settings, "breez_api_key");
    if (enableLightning && apiKey.isEmpty()) {
        // Prevent enabling Lightning without an API key; update settings and inform user
        blog(LOG_WARNING, "[QR Donations] Breez API key required to enable Lightning");
        obs_data_set_bool(settings, "enable_lightning", false);
        if (source && source->widget) {
            source->widget->setLightningStatus("Please provide a Breez API key before enabling Lightning.", false);
        }
        // Do not initialize Breez if apiKey is empty
        return;
    }
    if (enableLightning) {
        QString apiKey = obs_data_get_string(settings, "breez_api_key");
        QString sparkUrl = obs_data_get_string(settings, "spark_url");
        QString sparkKey = obs_data_get_string(settings, "spark_access_key");
        QString asset = obs_data_get_string(settings, "asset");
        QString network = (asset.compare("L-BTC", Qt::CaseInsensitive) == 0) ? QString("liquid") : QString("bitcoin");

        // Initialize Breez service (won't do anything if SDK not compiled in)
        bool initialized = BreezService::instance().initialize(apiKey, sparkUrl, sparkKey, network);
        if (initialized) {
            // Connect to payment received UI notification (only once)
                static bool connected = false;
            if (!connected) {
                QObject::connect(&BreezService::instance(), &BreezService::paymentReceived,
                                 static_cast<QRDonationsSource *>(data)->widget,
                                 &QRDonationsWidget::onPaymentReceived,
                                 Qt::QueuedConnection);
                    QObject::connect(&BreezService::instance(), &BreezService::serviceReady,
                                     static_cast<QRDonationsSource *>(data)->widget,
                                     [data](bool ready) {
                                         auto *src = static_cast<QRDonationsSource *>(data);
                                         if (src && src->widget) {
                                             if (ready)
                                                 src->widget->setLightningStatus("Lightning ready", true);
                                             else
                                                 src->widget->setLightningStatus("Lightning not ready", false);
                                         }
                                     },
                                     Qt::QueuedConnection);

                    QObject::connect(&BreezService::instance(), &BreezService::errorOccurred,
                                     static_cast<QRDonationsSource *>(data)->widget,
                                     [data](const QString &msg) {
                                         auto *src = static_cast<QRDonationsSource *>(data);
                                         if (src && src->widget) {
                                             src->widget->setLightningStatus(msg, false);
                                         }
                                     },
                                     Qt::QueuedConnection);
                connected = true;
            }
        }
    }
}

// Test Breez connection callback
static bool TestBreezConnection(obs_properties_t *props, obs_property_t *property, obs_data_t *settings)
{
    Q_UNUSED(props);
    Q_UNUSED(property);

    QString apiKey = obs_data_get_string(settings, "breez_api_key");
    QString sparkUrl = obs_data_get_string(settings, "spark_url");
    QString sparkKey = obs_data_get_string(settings, "spark_access_key");

    if (apiKey.isEmpty()) {
        blog(LOG_WARNING, "[QR Donations] Breez API key is empty; cannot test connection");
        QMainWindow *mw = static_cast<QMainWindow *>(obs_frontend_get_main_window());
        QMessageBox::warning(mw, "Breez Test", "Breez API key is required to test the connection.");
        return false;
    }

    // This will call the initialization flow which will emit serviceReady or errorOccurred
    QString asset = obs_data_get_string(settings, "asset");
    QString network = (asset.compare("L-BTC", Qt::CaseInsensitive) == 0) ? QString("liquid") : QString("bitcoin");
    bool ok = BreezService::instance().initialize(apiKey, sparkUrl, sparkKey, network);
    if (!ok) {
        blog(LOG_WARNING, "[QR Donations] Breez initialization (test) failed");
        QMainWindow *mw = static_cast<QMainWindow *>(obs_frontend_get_main_window());
        QMessageBox::critical(mw, "Breez Test", "Breez initialization failed. Check API key and Spark settings.");
        obs_data_set_string(settings, "breez_test_status", "Failed: invalid credentials or SDK not present");
    } else {
        QMainWindow *mw = static_cast<QMainWindow *>(obs_frontend_get_main_window());
        QMessageBox::information(mw, "Breez Test", "Breez initialized successfully. Lightning should now be available.");
        obs_data_set_string(settings, "breez_test_status", "Success: Breez initialized");
    }

    // UI will be updated by serviceReady / errorOccurred connected earlier
    return ok;
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
