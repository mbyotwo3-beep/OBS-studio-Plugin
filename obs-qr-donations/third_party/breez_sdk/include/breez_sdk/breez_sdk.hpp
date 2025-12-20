#pragma once

#include <string>
#include <vector>
#include <optional>
#include <cstdint>
#include <memory>

#include "breez_sdk.h" // C header

namespace breez_sdk {

enum class Network {
    BITCOIN = 0,
    TESTNET = 1,
    SIGNET = 2,
    REGTEST = 3,
    LIQUID = 4, // Added for compatibility
};

enum class EnvironmentType {
    PRODUCTION = 0,
    STAGING = 1,
};

struct NodeConfig {
    // We only know variant 1 (Greenlight) consumes 6 bytes (4 tag + 2 data)
    // We'll treat it as opaque data for now or try to reverse engineer further if needed.
    // For now, we assume we just pass through what default_config returns, 
    // or construct a simple one if we need to.
    std::vector<uint8_t> raw_data; 
};

struct SparkConfig {}; // Dummy struct for compatibility

struct Config {
    std::string breezserver;
    std::string chainnotifier_url;
    std::string mempoolspace_url;
    std::string working_dir;
    Network network;
    uint32_t payment_timeout_sec;
    std::optional<std::string> default_lsp_id;
    std::optional<std::string> api_key;
    double max_feerate_percent;
    uint64_t exemptfee_msat;
    NodeConfig node_config;
};

struct LogEntry {
    std::string line;
    std::string level;
};

class LogStream {
public:
    virtual ~LogStream() {}
    virtual void log(const LogEntry& l) = 0;
};

struct SdkEvent {
    // Simplified: we'll treat it as a variant or just pass the raw bytes
    std::vector<uint8_t> raw_data;
};

class EventListener {
public:
    virtual ~EventListener() {}
    virtual void on_event(const SdkEvent& e) = 0;
};

struct InvoicePaid {
    uint64_t amount_msat;
    std::string payment_hash;
    std::string description;
};

struct NodeInfo {
    std::string id;
    uint32_t block_height;
    uint64_t max_payable_msat;
    uint64_t max_receivable_msat;
    std::vector<std::string> connected_peers;
    uint64_t inbound_liquidity_msats;
    uint64_t channels_balance_msat;
    uint64_t onchain_balance_msat;
};

struct ListPaymentsRequest {
    // Add fields if needed
};

enum class PaymentStatus {
    PENDING = 0,
    COMPLETE = 1,
    FAILED = 2,
};

enum class PaymentType {
    SENT = 0,
    RECEIVED = 1,
    CLOSED_CHANNEL = 2,
};

struct Payment {
    std::string id;
    PaymentStatus status;
    PaymentType payment_type;
    uint64_t amount_msat;
    uint64_t fee_msat;
    uint64_t payment_time;
    std::string description;
    // Add other fields as needed
};

struct SendPaymentRequest {
    std::string bolt11;
};

struct SendPaymentResponse {
    std::string payment_id;
    std::string error_message;
    bool success;
};

struct OnChainSendRequest {
    std::string address;
    uint64_t amount_sat;
    Network network; // Added for compatibility
};

struct OnChainSendResponse {
    std::string txid;
    bool success; // Added for compatibility
    std::string error_message; // Added for compatibility
};

struct CreateInvoiceRequest {
    uint64_t amount_msat;
    std::string description;
    uint32_t expiry; // Added for compatibility
};

struct Invoice {
    std::string bolt11;
    std::string payment_hash;
};

class SDK {
public:
    SDK();
    ~SDK();

    // Static helper to get default config
    static Config default_config(EnvironmentType env_type, std::string api_key, NodeConfig node_config);

    // Connect function
    static std::unique_ptr<SDK> connect(const Config& config, const std::vector<uint8_t>& seed, EventListener* listener);

    // Methods
    NodeInfo node_info() const;
    std::vector<Payment> list_payments(const ListPaymentsRequest& req) const;
    SendPaymentResponse send_payment(const SendPaymentRequest& req) const;
    OnChainSendResponse send_on_chain(const OnChainSendRequest& req) const;
    Invoice create_invoice(const CreateInvoiceRequest& req) const;
    
    void set_payment_listener(EventListener* listener);

private:
    // Opaque pointer to internal SDK handle (if we had one)
    // For now, we might need to store the raw pointer or handle from C SDK.
    void* m_handle = nullptr;
};

// Serialization helpers (exposed for implementation)
std::vector<uint8_t> serialize_config(const Config& config);
Config deserialize_config(const std::vector<uint8_t>& bytes);

} // namespace breez_sdk
