#include <iostream>
#include <vector>
#include <iomanip>
#include <cstring>
#include "breez_sdk.h"

// Helper to print hex dump
void print_hex(const RustBuffer& buf) {
    std::cout << "Buffer len: " << buf.len << ", capacity: " << buf.capacity << std::endl;
    const uint8_t* data = (const uint8_t*)buf.data;
    for (size_t i = 0; i < buf.len; ++i) {
        std::cout << std::hex << std::setw(2) << std::setfill('0') << (int)data[i] << " ";
        if ((i + 1) % 16 == 0) std::cout << std::endl;
    }
    std::cout << std::dec << std::endl;
}

int main() {
    // EnvironmentType::Production = ? (0 failed)
    // Let's try 1.
    uint8_t env_data[] = {0, 0, 0, 1}; 
    ForeignBytes env_bytes = {4, env_data};

    RustCallStatus status = {0};

    RustBuffer env_buf = ffi_breez_sdk_bindings_rustbuffer_from_bytes(env_bytes, &status);
    if (status.code != 0) {
        std::cerr << "Error creating env buffer" << std::endl;
        return 1;
    }

    // Empty API key
    uint8_t empty[] = {};
    ForeignBytes empty_bytes = {0, empty};
    RustBuffer api_key_buf = ffi_breez_sdk_bindings_rustbuffer_from_bytes(empty_bytes, &status);

    // NodeConfig (Greenlight?)
    // Variant 1 consumes 6 bytes total (4 tag + 2 data).
   // Variant 0 (Spark?) - 4 bytes tag + ?
  // Let's try just the tag first
  std::vector<uint8_t> node_config_data = {0, 0, 0, 2};
  RustBuffer node_config = ffi_breez_sdk_bindings_rustbuffer_from_bytes(
      ForeignBytes{(int)node_config_data.size(), node_config_data.data()}, &status);
    
    std::cout << "Calling default_config..." << std::endl;
    RustBuffer config_buf = uniffi_breez_sdk_bindings_fn_func_default_config(env_buf, api_key_buf, node_config, &status);

    // Input buffers are consumed by the function, so we don't free them.


    if (status.code != 0) {
        std::cerr << "Error calling default_config. Code: " << (int)status.code << std::endl;
        if (status.errorBuf.len > 0) {
             std::cerr << "Error message: " << (char*)status.errorBuf.data << std::endl;
        }
        return 1;
    }

    std::cout << "Success! Config buffer:" << std::endl;
    print_hex(config_buf);

    // Clean up
    if (config_buf.data) {
        ffi_breez_sdk_bindings_rustbuffer_free(config_buf, &status);
    }

    return 0;
}
