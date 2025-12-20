#ifndef BREEZ_SDK_H
#define BREEZ_SDK_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct RustBuffer {
    uint64_t capacity;
    uint64_t len;
    uint8_t *data;
} RustBuffer;

typedef struct RustCallStatus {
    int8_t code;
    RustBuffer errorBuf;
} RustCallStatus;

typedef struct ForeignBytes {
    int len;
    const uint8_t *data;
} ForeignBytes;

// Helper functions (exported by the library)
RustBuffer ffi_breez_sdk_bindings_rustbuffer_alloc(int size, RustCallStatus *status);
RustBuffer ffi_breez_sdk_bindings_rustbuffer_from_bytes(struct ForeignBytes bytes, RustCallStatus *status);
void ffi_breez_sdk_bindings_rustbuffer_free(RustBuffer buf, RustCallStatus *status);

// Greenlight SDK functions
RustBuffer uniffi_breez_sdk_bindings_fn_func_default_config(RustBuffer env_type, RustBuffer api_key, RustBuffer node_config, RustCallStatus *status);
void* uniffi_breez_sdk_bindings_fn_func_connect(RustBuffer req, RustBuffer seed, RustCallStatus *status);

// Methods on BlockingBreezServices (handle is void*)
void uniffi_breez_sdk_bindings_fn_free_blockingbreezservices(void* ptr, RustCallStatus *status);
RustBuffer uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_node_info(void* ptr, RustCallStatus *status);
RustBuffer uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_list_payments(void* ptr, RustBuffer req, RustCallStatus *status);
RustBuffer uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_send_payment(void* ptr, RustBuffer req, RustCallStatus *status);
RustBuffer uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_send_onchain(void* ptr, RustBuffer req, RustCallStatus *status);
RustBuffer uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_receive_payment(void* ptr, RustBuffer req, RustCallStatus *status);
RustBuffer uniffi_breez_sdk_bindings_fn_method_blockingbreezservices_create_invoice(void* ptr, RustBuffer req, RustCallStatus *status); // Guessing name

// EventListener callback
typedef void (*uniffi_callback_eventlistener_on_event)(uint64_t handle, RustBuffer event, RustCallStatus *status);

typedef struct uniffi_vtable_eventlistener {
    uniffi_callback_eventlistener_on_event on_event;
} uniffi_vtable_eventlistener;

void uniffi_breez_sdk_bindings_fn_init_callback_vtable_eventlistener(const struct uniffi_vtable_eventlistener *vtable, RustCallStatus *status);

#ifdef __cplusplus
}
#endif

#endif // BREEZ_SDK_H
