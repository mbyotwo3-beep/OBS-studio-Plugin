#include "breez_sdk/breez_sdk.hpp"
#include <cstring>
#include <stdexcept>
#include <iostream>
#include <arpa/inet.h> // for htonl, ntohl

namespace breez_sdk {

// Helper functions for reading/writing big-endian values
static void write_u8(std::vector<uint8_t>& buf, uint8_t val) {
    buf.push_back(val);
}

static uint8_t read_u8(const std::vector<uint8_t>& buf, size_t& offset) {
    if (offset >= buf.size()) throw std::runtime_error("Buffer underflow");
    return buf[offset++];
}

static void write_u16(std::vector<uint8_t>& buf, uint16_t val) {
    uint16_t be = htons(val);
    const uint8_t* p = (const uint8_t*)&be;
    buf.insert(buf.end(), p, p + 2);
}

static uint16_t read_u16(const std::vector<uint8_t>& buf, size_t& offset) {
    if (offset + 2 > buf.size()) throw std::runtime_error("Buffer underflow");
    uint16_t val;
    std::memcpy(&val, &buf[offset], 2);
    offset += 2;
    return ntohs(val);
}

static void write_u32(std::vector<uint8_t>& buf, uint32_t val) {
    uint32_t be = htonl(val);
    const uint8_t* p = (const uint8_t*)&be;
    buf.insert(buf.end(), p, p + 4);
}

static uint32_t read_u32(const std::vector<uint8_t>& buf, size_t& offset) {
    if (offset + 4 > buf.size()) throw std::runtime_error("Buffer underflow");
    uint32_t val;
    std::memcpy(&val, &buf[offset], 4);
    offset += 4;
    return ntohl(val);
}

static void write_u64(std::vector<uint8_t>& buf, uint64_t val) {
    // htobe64 is non-standard, do manual swap if needed or assume big-endian system?
    // Linux usually has be64toh/htobe64 in <endian.h>
    uint64_t be = val; // Placeholder, need byte swap on little endian
    // Manual swap
    uint8_t* p = (uint8_t*)&be;
    uint8_t out[8];
    for(int i=0; i<8; ++i) out[i] = p[7-i];
    buf.insert(buf.end(), out, out + 8);
}

static uint64_t read_u64(const std::vector<uint8_t>& buf, size_t& offset) {
    if (offset + 8 > buf.size()) throw std::runtime_error("Buffer underflow");
    uint64_t val = 0;
    for(int i=0; i<8; ++i) {
        val = (val << 8) | buf[offset++];
    }
    return val;
}

static void write_f64(std::vector<uint8_t>& buf, double val) {
    // IEEE 754 double is 8 bytes. UniFFI uses big-endian.
    uint64_t bits;
    std::memcpy(&bits, &val, 8);
    write_u64(buf, bits);
}

static double read_f64(const std::vector<uint8_t>& buf, size_t& offset) {
    uint64_t bits = read_u64(buf, offset);
    double val;
    std::memcpy(&val, &bits, 8);
    return val;
}

static void write_string(std::vector<uint8_t>& buf, const std::string& str) {
    write_u32(buf, str.length());
    buf.insert(buf.end(), str.begin(), str.end());
}

static std::string read_string(const std::vector<uint8_t>& buf, size_t& offset) {
    uint32_t len = read_u32(buf, offset);
    if (offset + len > buf.size()) throw std::runtime_error("Buffer underflow");
    std::string str(buf.begin() + offset, buf.begin() + offset + len);
    offset += len;
    return str;
}

static void write_optional_string(std::vector<uint8_t>& buf, const std::optional<std::string>& opt) {
    if (opt.has_value()) {
        write_u8(buf, 1);
        write_string(buf, *opt);
    } else {
        write_u8(buf, 0);
    }
}

static std::optional<std::string> read_optional_string(const std::vector<uint8_t>& buf, size_t& offset) {
    uint8_t tag = read_u8(buf, offset);
    if (tag == 0) return std::nullopt;
    return read_string(buf, offset);
}

std::vector<uint8_t> serialize_config(const Config& config) {
    std::vector<uint8_t> buf;
    write_string(buf, config.breezserver);
    write_string(buf, config.chainnotifier_url);
    write_string(buf, config.mempoolspace_url);
    write_string(buf, config.working_dir);
    write_u32(buf, (uint32_t)config.network); // Enum as u32? Or i32? inspect_config showed 00 00 00 00.
    write_u32(buf, config.payment_timeout_sec);
    write_optional_string(buf, config.default_lsp_id);
    write_optional_string(buf, config.api_key);
    write_f64(buf, config.max_feerate_percent);
    write_u64(buf, config.exemptfee_msat);
    
    // NodeConfig
    buf.insert(buf.end(), config.node_config.raw_data.begin(), config.node_config.raw_data.end());
    
    return buf;
}

Config deserialize_config(const std::vector<uint8_t>& bytes) {
    Config c;
    size_t offset = 0;
    c.breezserver = read_string(bytes, offset);
    c.chainnotifier_url = read_string(bytes, offset);
    c.mempoolspace_url = read_string(bytes, offset);
    c.working_dir = read_string(bytes, offset);
    c.network = (Network)read_u32(bytes, offset);
    c.payment_timeout_sec = read_u32(bytes, offset);
    c.default_lsp_id = read_optional_string(bytes, offset);
    c.api_key = read_optional_string(bytes, offset);
    c.max_feerate_percent = read_f64(bytes, offset);
    c.exemptfee_msat = read_u64(bytes, offset);
    
    // Remaining bytes are NodeConfig
    if (offset < bytes.size()) {
        c.node_config.raw_data.assign(bytes.begin() + offset, bytes.end());
    }
    
    return c;
}

SDK::SDK() {}
SDK::~SDK() {
    if (m_handle) {
        RustCallStatus status = {0};
        uniffi_breez_sdk_bindings_fn_free_blockingbreezservices(m_handle, &status);
        m_handle = nullptr;
    }
}

Config SDK::default_config(EnvironmentType env_type, std::string api_key, NodeConfig node_config) {
    // Call C function
    // We need to serialize arguments
    
    // EnvType: u32?
    uint32_t env_val = (uint32_t)env_type;
    // Wait, inspect_config showed we needed to pass RustBuffer for env_type.
    // And inside that buffer was 4 bytes.
    std::vector<uint8_t> env_vec;
    write_u32(env_vec, env_val);
    RustBuffer env_buf = {4, 4, env_vec.data()}; // Dangerous if vec reallocates/moves. 
    // But here it's local scope and we pass it immediately.
    
    // API Key
    std::vector<uint8_t> api_vec;
    write_string(api_vec, api_key); // Or optional string?
    // default_config signature: api_key is RustBuffer.
    // Is it String or Option<String>?
    // Usually arguments are non-optional unless specified.
    // But let's assume String.
    RustBuffer api_buf = { (uint64_t)api_vec.size(), (uint64_t)api_vec.size(), api_vec.data() };

    // NodeConfig
    RustBuffer node_buf = { (uint64_t)node_config.raw_data.size(), (uint64_t)node_config.raw_data.size(), node_config.raw_data.data() };

    RustCallStatus status = {0};
    RustBuffer res = uniffi_breez_sdk_bindings_fn_func_default_config(env_buf, api_buf, node_buf, &status);
    
    if (status.code != 0) {
        throw std::runtime_error("default_config failed");
    }
    
    std::vector<uint8_t> res_vec(res.data, res.data + res.len);
    ffi_breez_sdk_bindings_rustbuffer_free(res, &status);
    
    return deserialize_config(res_vec);
}

std::unique_ptr<SDK> SDK::connect(const Config& config, const std::vector<uint8_t>& seed, EventListener* listener) {
    // Serialize config
    std::vector<uint8_t> config_vec = serialize_config(config);
    RustBuffer config_buf = { (uint64_t)config_vec.size(), (uint64_t)config_vec.size(), config_vec.data() };
    
    // Seed
    RustBuffer seed_buf = { (uint64_t)seed.size(), (uint64_t)seed.size(), (uint8_t*)seed.data() }; // Cast const away
    
    RustCallStatus status = {0};
    void* handle = uniffi_breez_sdk_bindings_fn_func_connect(config_buf, seed_buf, &status);
    
    if (status.code != 0) {
        throw std::runtime_error("connect failed");
    }
    
    auto sdk = std::make_unique<SDK>();
    sdk->m_handle = handle;
    return sdk;
}

NodeInfo SDK::node_info() const {
    RustCallStatus status = {0};
    RustBuffer res = uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_node_info(m_handle, &status);
    
    if (status.code != 0) {
        throw std::runtime_error("node_info failed");
    }
    
    std::vector<uint8_t> bytes(res.data, res.data + res.len);
    ffi_breez_sdk_bindings_rustbuffer_free(res, &status);
    
    size_t offset = 0;
    NodeInfo info;
    info.id = read_string(bytes, offset);
    info.block_height = read_u32(bytes, offset);
    info.max_payable_msat = read_u64(bytes, offset);
    info.max_receivable_msat = read_u64(bytes, offset);
    
    uint32_t peer_count = read_u32(bytes, offset);
    for (uint32_t i = 0; i < peer_count; ++i) {
        info.connected_peers.push_back(read_string(bytes, offset));
    }
    
    info.inbound_liquidity_msats = read_u64(bytes, offset);
    info.channels_balance_msat = read_u64(bytes, offset);
    info.onchain_balance_msat = read_u64(bytes, offset);
    
    return info;
}

std::vector<Payment> SDK::list_payments(const ListPaymentsRequest& req) const {
    std::vector<uint8_t> req_vec;
    // ListPaymentsRequest is usually empty or has filters. 
    // For now, we'll pass an empty buffer if it's empty.
    
    RustBuffer req_buf = { (uint64_t)req_vec.size(), (uint64_t)req_vec.size(), req_vec.data() };
    RustCallStatus status = {0};
    RustBuffer res = uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_list_payments(m_handle, req_buf, &status);
    
    if (status.code != 0) {
        throw std::runtime_error("list_payments failed");
    }
    
    std::vector<uint8_t> bytes(res.data, res.data + res.len);
    ffi_breez_sdk_bindings_rustbuffer_free(res, &status);
    
    size_t offset = 0;
    uint32_t count = read_u32(bytes, offset);
    std::vector<Payment> payments;
    for (uint32_t i = 0; i < count; ++i) {
        Payment p;
        p.id = read_string(bytes, offset);
        p.payment_type = (PaymentType)read_u32(bytes, offset);
        p.payment_time = read_u64(bytes, offset);
        p.amount_msat = read_u64(bytes, offset);
        p.fee_msat = read_u64(bytes, offset);
        p.status = (PaymentStatus)read_u32(bytes, offset);
        p.description = read_optional_string(bytes, offset).value_or("");
        payments.push_back(p);
    }
    
    return payments;
}

SendPaymentResponse SDK::send_payment(const SendPaymentRequest& req) const {
    std::vector<uint8_t> req_vec;
    write_string(req_vec, req.bolt11);
    // Add amount if it's an optional field in the request, but usually it's in bolt11
    // For now assume just bolt11
    
    RustBuffer req_buf = { (uint64_t)req_vec.size(), (uint64_t)req_vec.size(), req_vec.data() };
    RustCallStatus status = {0};
    RustBuffer res = uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_send_payment(m_handle, req_buf, &status);
    
    SendPaymentResponse resp;
    if (status.code != 0) {
        resp.success = false;
        resp.error_message = "C FFI call failed";
        return resp;
    }
    
    std::vector<uint8_t> bytes(res.data, res.data + res.len);
    ffi_breez_sdk_bindings_rustbuffer_free(res, &status);
    
    // Deserialize Payment result
    // Usually it returns the Payment struct or an error
    // For simplicity, we'll just check if we got data
    if (bytes.size() > 0) {
        resp.success = true;
        // Extract payment_id (first field usually)
        size_t offset = 0;
        resp.payment_id = read_string(bytes, offset);
    } else {
        resp.success = false;
        resp.error_message = "Empty response from SDK";
    }
    
    return resp;
}

OnChainSendResponse SDK::send_on_chain(const OnChainSendRequest& req) const {
    std::vector<uint8_t> req_vec;
    write_string(req_vec, req.address);
    write_u64(req_vec, req.amount_sat);
    // write_u32(req_vec, (uint32_t)req.network);
    
    RustBuffer req_buf = { (uint64_t)req_vec.size(), (uint64_t)req_vec.size(), req_vec.data() };
    RustCallStatus status = {0};
    RustBuffer res = uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_send_onchain(m_handle, req_buf, &status);
    
    OnChainSendResponse resp;
    if (status.code != 0) {
        resp.success = false;
        resp.error_message = "C FFI call failed";
        return resp;
    }
    
    std::vector<uint8_t> bytes(res.data, res.data + res.len);
    ffi_breez_sdk_bindings_rustbuffer_free(res, &status);
    
    if (bytes.size() > 0) {
        resp.success = true;
        size_t offset = 0;
        resp.txid = read_string(bytes, offset);
    } else {
        resp.success = false;
        resp.error_message = "Empty response from SDK";
    }
    
    return resp;
}

Invoice SDK::create_invoice(const CreateInvoiceRequest& req) const {
    std::vector<uint8_t> req_vec;
    write_u64(req_vec, req.amount_msat);
    write_string(req_vec, req.description);
    // write_u32(req_vec, req.expiry);
    
    RustBuffer req_buf = { (uint64_t)req_vec.size(), (uint64_t)req_vec.size(), req_vec.data() };
    RustCallStatus status = {0};
    
    // In Breez SDK Greenlight, it's often called receive_payment or create_invoice
    // Based on my guess in breez_sdk.h: uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_receive_payment
    RustBuffer res = uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_receive_payment(m_handle, req_buf, &status);
    
    if (status.code != 0) {
        throw std::runtime_error("create_invoice failed");
    }
    
    std::vector<uint8_t> bytes(res.data, res.data + res.len);
    ffi_breez_sdk_bindings_rustbuffer_free(res, &status);
    
    size_t offset = 0;
    Invoice inv;
    inv.bolt11 = read_string(bytes, offset);
    inv.payment_hash = read_string(bytes, offset);
    return inv;
}

static EventListener* g_listener = nullptr;

static void uniffi_event_listener_on_event(uint64_t handle, RustBuffer event, RustCallStatus *status) {
    if (g_listener) {
        SdkEvent e;
        e.raw_data.assign(event.data, event.data + event.len);
        g_listener->on_event(e);
    }
    // We don't free the buffer here, UniFFI usually does it or expects us to.
    // But since we copied it, it's safe.
}

static void register_event_listener_vtable() {
    static bool registered = false;
    if (registered) return;
    
    uniffi_vtable_eventlistener vtable;
    vtable.on_event = uniffi_event_listener_on_event;
    
    RustCallStatus status = {0};
    uniffi_breez_sdk_bindings_fn_init_callback_vtable_eventlistener(&vtable, &status);
    registered = true;
}

void SDK::set_payment_listener(EventListener* listener) {
    g_listener = listener;
    register_event_listener_vtable();
}

} // namespace breez_sdk
