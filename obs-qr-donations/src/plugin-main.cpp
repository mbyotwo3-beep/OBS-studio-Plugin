#include <obs-module.h>
#include "qr-donations.hpp"

OBS_DECLARE_MODULE()
OBS_MODULE_USE_DEFAULT_LOCALE("obs-qr-donations", "en-US")

bool obs_module_load(void)
{
    blog(LOG_INFO, "[QR Donations] Plugin loaded successfully (version %s)", "1.0.0");
    
    // Register our source
    Q_INIT_RESOURCE(obs_qr_donations);
    QRDonations::InitializeSource();
    
    return true;
}

void obs_module_unload()
{
    blog(LOG_INFO, "[QR Donations] Plugin unloaded");
    Q_CLEANUP_RESOURCE(obs_qr_donations);
}

const char *obs_module_name()
{
    return "QR Donations";
}

const char *obs_module_description()
{
    return "Displays QR codes for receiving Bitcoin donations (on-chain and Lightning via Breez)";
}
