package breez_sdk_spark

// #include <breez_sdk_spark.h>
import "C"

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"runtime"
	"runtime/cgo"
	"sync"
	"sync/atomic"
	"unsafe"
)

// This is needed, because as of go 1.24
// type RustBuffer C.RustBuffer cannot have methods,
// RustBuffer is treated as non-local type
type GoRustBuffer struct {
	inner C.RustBuffer
}

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() uint64
	Capacity() uint64
}

func RustBufferFromExternal(b RustBufferI) GoRustBuffer {
	return GoRustBuffer{
		inner: C.RustBuffer{
			capacity: C.uint64_t(b.Capacity()),
			len:      C.uint64_t(b.Len()),
			data:     (*C.uchar)(b.Data()),
		},
	}
}

func (cb GoRustBuffer) Capacity() uint64 {
	return uint64(cb.inner.capacity)
}

func (cb GoRustBuffer) Len() uint64 {
	return uint64(cb.inner.len)
}

func (cb GoRustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.inner.data)
}

func (cb GoRustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.inner.data), C.uint64_t(cb.inner.len))
	return bytes.NewReader(b)
}

func (cb GoRustBuffer) Free() {
	rustCall(func(status *C.RustCallStatus) bool {
		C.ffi_breez_sdk_spark_rustbuffer_free(cb.inner, status)
		return false
	})
}

func (cb GoRustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.inner.data), C.int(cb.inner.len))
}

func stringToRustBuffer(str string) C.RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) C.RustBuffer {
	if len(b) == 0 {
		return C.RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes{
		len:  C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}

	return rustCall(func(status *C.RustCallStatus) C.RustBuffer {
		return C.ffi_breez_sdk_spark_rustbuffer_from_bytes(foreign, status)
	})
}

type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) C.RustBuffer
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) C.RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}

func rustCallWithError[E any, U any](converter BufReader[*E], callback func(*C.RustCallStatus) U) (U, *E) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)
	return returnValue, err
}

func checkCallStatus[E any](converter BufReader[*E], status C.RustCallStatus) *E {
	switch status.code {
	case 0:
		return nil
	case 1:
		return LiftFromRustBuffer(converter, GoRustBuffer{inner: status.errorBuf})
	case 2:
		// when the rust code sees a panic, it tries to construct a rustBuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(GoRustBuffer{inner: status.errorBuf})))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		panic(fmt.Errorf("unknown status code: %d", status.code))
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a C.RustBuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(GoRustBuffer{
				inner: status.errorBuf,
			})))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError[error](nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}

type NativeError interface {
	AsError() error
}

func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {

	FfiConverterBitcoinChainServiceINSTANCE.register()
	FfiConverterFiatServiceINSTANCE.register()
	FfiConverterPaymentObserverINSTANCE.register()
	FfiConverterRestClientINSTANCE.register()
	FfiConverterStorageINSTANCE.register()
	FfiConverterSyncStorageINSTANCE.register()
	FfiConverterCallbackInterfaceEventListenerINSTANCE.register()
	FfiConverterCallbackInterfaceLoggerINSTANCE.register()
	uniffiCheckChecksums()
}

func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 26
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_breez_sdk_spark_uniffi_contract_version()
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("breez_sdk_spark: UniFFI contract version mismatch")
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_func_connect()
		})
		if checksum != 40345 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_func_connect: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_func_default_config()
		})
		if checksum != 62194 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_func_default_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_func_init_logging()
		})
		if checksum != 8518 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_func_init_logging: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_get_address_utxos()
		})
		if checksum != 20959 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_get_address_utxos: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_get_transaction_status()
		})
		if checksum != 23018 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_get_transaction_status: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_get_transaction_hex()
		})
		if checksum != 59376 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_get_transaction_hex: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_broadcast_transaction()
		})
		if checksum != 65179 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_broadcast_transaction: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_recommended_fees()
		})
		if checksum != 43230 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_bitcoinchainservice_recommended_fees: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_add_event_listener()
		})
		if checksum != 37737 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_add_event_listener: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_check_lightning_address_available()
		})
		if checksum != 31624 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_check_lightning_address_available: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_check_message()
		})
		if checksum != 4385 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_check_message: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_claim_deposit()
		})
		if checksum != 43529 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_claim_deposit: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_claim_htlc_payment()
		})
		if checksum != 57587 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_claim_htlc_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_delete_lightning_address()
		})
		if checksum != 44132 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_delete_lightning_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_disconnect()
		})
		if checksum != 330 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_disconnect: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_get_info()
		})
		if checksum != 6771 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_get_info: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_get_lightning_address()
		})
		if checksum != 36552 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_get_lightning_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_get_payment()
		})
		if checksum != 11540 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_get_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_get_token_issuer()
		})
		if checksum != 26649 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_get_token_issuer: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_get_tokens_metadata()
		})
		if checksum != 40125 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_get_tokens_metadata: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_get_user_settings()
		})
		if checksum != 38537 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_get_user_settings: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_list_fiat_currencies()
		})
		if checksum != 63366 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_list_fiat_currencies: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_list_fiat_rates()
		})
		if checksum != 5904 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_list_fiat_rates: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_list_payments()
		})
		if checksum != 16156 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_list_payments: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_list_unclaimed_deposits()
		})
		if checksum != 22486 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_list_unclaimed_deposits: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_lnurl_pay()
		})
		if checksum != 10147 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_lnurl_pay: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_lnurl_withdraw()
		})
		if checksum != 45652 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_lnurl_withdraw: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_parse()
		})
		if checksum != 14285 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_parse: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_prepare_lnurl_pay()
		})
		if checksum != 37691 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_prepare_lnurl_pay: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_prepare_send_payment()
		})
		if checksum != 34185 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_prepare_send_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_receive_payment()
		})
		if checksum != 36984 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_receive_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_recommended_fees()
		})
		if checksum != 16947 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_recommended_fees: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_refund_deposit()
		})
		if checksum != 33646 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_refund_deposit: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_register_lightning_address()
		})
		if checksum != 530 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_register_lightning_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_remove_event_listener()
		})
		if checksum != 41066 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_remove_event_listener: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_send_payment()
		})
		if checksum != 54349 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_send_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_sign_message()
		})
		if checksum != 57563 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_sign_message: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_sync_wallet()
		})
		if checksum != 30368 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_sync_wallet: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_breezsdk_update_user_settings()
		})
		if checksum != 1721 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_breezsdk_update_user_settings: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_fiatservice_fetch_fiat_currencies()
		})
		if checksum != 19092 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_fiatservice_fetch_fiat_currencies: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_fiatservice_fetch_fiat_rates()
		})
		if checksum != 11512 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_fiatservice_fetch_fiat_rates: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_paymentobserver_before_send()
		})
		if checksum != 30686 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_paymentobserver_before_send: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_restclient_get_request()
		})
		if checksum != 8260 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_restclient_get_request: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_restclient_post_request()
		})
		if checksum != 24889 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_restclient_post_request: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_restclient_delete_request()
		})
		if checksum != 51072 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_restclient_delete_request: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_build()
		})
		if checksum != 8126 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_build: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_chain_service()
		})
		if checksum != 2848 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_chain_service: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_default_storage()
		})
		if checksum != 14543 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_default_storage: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_fiat_service()
		})
		if checksum != 37854 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_fiat_service: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_key_set()
		})
		if checksum != 42926 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_key_set: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_lnurl_client()
		})
		if checksum != 51060 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_lnurl_client: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_payment_observer()
		})
		if checksum != 21617 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_payment_observer: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_real_time_sync_storage()
		})
		if checksum != 20579 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_real_time_sync_storage: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_rest_chain_service()
		})
		if checksum != 63155 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_rest_chain_service: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_storage()
		})
		if checksum != 59400 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_sdkbuilder_with_storage: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_delete_cached_item()
		})
		if checksum != 6883 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_delete_cached_item: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_get_cached_item()
		})
		if checksum != 30248 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_get_cached_item: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_set_cached_item()
		})
		if checksum != 7970 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_set_cached_item: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_list_payments()
		})
		if checksum != 19728 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_list_payments: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_insert_payment()
		})
		if checksum != 28075 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_insert_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_set_payment_metadata()
		})
		if checksum != 45500 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_set_payment_metadata: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_get_payment_by_id()
		})
		if checksum != 35394 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_get_payment_by_id: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_get_payment_by_invoice()
		})
		if checksum != 57075 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_get_payment_by_invoice: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_add_deposit()
		})
		if checksum != 60240 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_add_deposit: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_delete_deposit()
		})
		if checksum != 60586 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_delete_deposit: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_list_deposits()
		})
		if checksum != 54118 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_list_deposits: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_update_deposit()
		})
		if checksum != 39803 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_update_deposit: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_storage_set_lnurl_metadata()
		})
		if checksum != 7460 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_storage_set_lnurl_metadata: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_add_outgoing_change()
		})
		if checksum != 19087 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_add_outgoing_change: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_complete_outgoing_sync()
		})
		if checksum != 20071 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_complete_outgoing_sync: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_get_pending_outgoing_changes()
		})
		if checksum != 23473 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_get_pending_outgoing_changes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_get_last_revision()
		})
		if checksum != 36887 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_get_last_revision: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_insert_incoming_records()
		})
		if checksum != 41782 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_insert_incoming_records: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_delete_incoming_record()
		})
		if checksum != 23002 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_delete_incoming_record: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_rebase_pending_outgoing_records()
		})
		if checksum != 61508 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_rebase_pending_outgoing_records: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_get_incoming_records()
		})
		if checksum != 53552 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_get_incoming_records: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_get_latest_outgoing_change()
		})
		if checksum != 16326 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_get_latest_outgoing_change: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_syncstorage_update_record_from_incoming()
		})
		if checksum != 9986 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_syncstorage_update_record_from_incoming: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_tokenissuer_burn_issuer_token()
		})
		if checksum != 56056 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_tokenissuer_burn_issuer_token: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_tokenissuer_create_issuer_token()
		})
		if checksum != 33277 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_tokenissuer_create_issuer_token: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_tokenissuer_freeze_issuer_token()
		})
		if checksum != 32344 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_tokenissuer_freeze_issuer_token: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_tokenissuer_get_issuer_token_balance()
		})
		if checksum != 9758 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_tokenissuer_get_issuer_token_balance: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_tokenissuer_get_issuer_token_metadata()
		})
		if checksum != 57707 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_tokenissuer_get_issuer_token_metadata: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_tokenissuer_mint_issuer_token()
		})
		if checksum != 36459 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_tokenissuer_mint_issuer_token: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_tokenissuer_unfreeze_issuer_token()
		})
		if checksum != 65025 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_tokenissuer_unfreeze_issuer_token: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_constructor_sdkbuilder_new()
		})
		if checksum != 65435 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_constructor_sdkbuilder_new: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_eventlistener_on_event()
		})
		if checksum != 24807 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_eventlistener_on_event: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_breez_sdk_spark_checksum_method_logger_log()
		})
		if checksum != 11839 {
			// If this happens try cleaning and rebuilding your project
			panic("breez_sdk_spark: uniffi_breez_sdk_spark_checksum_method_logger_log: UniFFI API checksum mismatch")
		}
	}
}

type FfiConverterUint16 struct{}

var FfiConverterUint16INSTANCE = FfiConverterUint16{}

func (FfiConverterUint16) Lower(value uint16) C.uint16_t {
	return C.uint16_t(value)
}

func (FfiConverterUint16) Write(writer io.Writer, value uint16) {
	writeUint16(writer, value)
}

func (FfiConverterUint16) Lift(value C.uint16_t) uint16 {
	return uint16(value)
}

func (FfiConverterUint16) Read(reader io.Reader) uint16 {
	return readUint16(reader)
}

type FfiDestroyerUint16 struct{}

func (FfiDestroyerUint16) Destroy(_ uint16) {}

type FfiConverterUint32 struct{}

var FfiConverterUint32INSTANCE = FfiConverterUint32{}

func (FfiConverterUint32) Lower(value uint32) C.uint32_t {
	return C.uint32_t(value)
}

func (FfiConverterUint32) Write(writer io.Writer, value uint32) {
	writeUint32(writer, value)
}

func (FfiConverterUint32) Lift(value C.uint32_t) uint32 {
	return uint32(value)
}

func (FfiConverterUint32) Read(reader io.Reader) uint32 {
	return readUint32(reader)
}

type FfiDestroyerUint32 struct{}

func (FfiDestroyerUint32) Destroy(_ uint32) {}

type FfiConverterUint64 struct{}

var FfiConverterUint64INSTANCE = FfiConverterUint64{}

func (FfiConverterUint64) Lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func (FfiConverterUint64) Write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func (FfiConverterUint64) Lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func (FfiConverterUint64) Read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type FfiDestroyerUint64 struct{}

func (FfiDestroyerUint64) Destroy(_ uint64) {}

type FfiConverterFloat64 struct{}

var FfiConverterFloat64INSTANCE = FfiConverterFloat64{}

func (FfiConverterFloat64) Lower(value float64) C.double {
	return C.double(value)
}

func (FfiConverterFloat64) Write(writer io.Writer, value float64) {
	writeFloat64(writer, value)
}

func (FfiConverterFloat64) Lift(value C.double) float64 {
	return float64(value)
}

func (FfiConverterFloat64) Read(reader io.Reader) float64 {
	return readFloat64(reader)
}

type FfiDestroyerFloat64 struct{}

func (FfiDestroyerFloat64) Destroy(_ float64) {}

type FfiConverterBool struct{}

var FfiConverterBoolINSTANCE = FfiConverterBool{}

func (FfiConverterBool) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func (FfiConverterBool) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func (FfiConverterBool) Lift(value C.int8_t) bool {
	return value != 0
}

func (FfiConverterBool) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type FfiDestroyerBool struct{}

func (FfiDestroyerBool) Destroy(_ bool) {}

type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) C.RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct{}

func (FfiDestroyerString) Destroy(_ string) {}

type FfiConverterBytes struct{}

var FfiConverterBytesINSTANCE = FfiConverterBytes{}

func (c FfiConverterBytes) Lower(value []byte) C.RustBuffer {
	return LowerIntoRustBuffer[[]byte](c, value)
}

func (c FfiConverterBytes) Write(writer io.Writer, value []byte) {
	if len(value) > math.MaxInt32 {
		panic("[]byte is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := writer.Write(value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing []byte, expected %d, written %d", len(value), write_length))
	}
}

func (c FfiConverterBytes) Lift(rb RustBufferI) []byte {
	return LiftFromRustBuffer[[]byte](c, rb)
}

func (c FfiConverterBytes) Read(reader io.Reader) []byte {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading []byte, expected %d, read %d", length, read_length))
	}
	return buffer
}

type FfiDestroyerBytes struct{}

func (FfiDestroyerBytes) Destroy(_ []byte) {}

// Below is an implementation of synchronization requirements outlined in the link.
// https://github.com/mozilla/uniffi-rs/blob/0dc031132d9493ca812c3af6e7dd60ad2ea95bf0/uniffi_bindgen/src/bindings/kotlin/templates/ObjectRuntime.kt#L31

type FfiObject struct {
	pointer       unsafe.Pointer
	callCounter   atomic.Int64
	cloneFunction func(unsafe.Pointer, *C.RustCallStatus) unsafe.Pointer
	freeFunction  func(unsafe.Pointer, *C.RustCallStatus)
	destroyed     atomic.Bool
}

func newFfiObject(
	pointer unsafe.Pointer,
	cloneFunction func(unsafe.Pointer, *C.RustCallStatus) unsafe.Pointer,
	freeFunction func(unsafe.Pointer, *C.RustCallStatus),
) FfiObject {
	return FfiObject{
		pointer:       pointer,
		cloneFunction: cloneFunction,
		freeFunction:  freeFunction,
	}
}

func (ffiObject *FfiObject) incrementPointer(debugName string) unsafe.Pointer {
	for {
		counter := ffiObject.callCounter.Load()
		if counter <= -1 {
			panic(fmt.Errorf("%v object has already been destroyed", debugName))
		}
		if counter == math.MaxInt64 {
			panic(fmt.Errorf("%v object call counter would overflow", debugName))
		}
		if ffiObject.callCounter.CompareAndSwap(counter, counter+1) {
			break
		}
	}

	return rustCall(func(status *C.RustCallStatus) unsafe.Pointer {
		return ffiObject.cloneFunction(ffiObject.pointer, status)
	})
}

func (ffiObject *FfiObject) decrementPointer() {
	if ffiObject.callCounter.Add(-1) == -1 {
		ffiObject.freeRustArcPtr()
	}
}

func (ffiObject *FfiObject) destroy() {
	if ffiObject.destroyed.CompareAndSwap(false, true) {
		if ffiObject.callCounter.Add(-1) == -1 {
			ffiObject.freeRustArcPtr()
		}
	}
}

func (ffiObject *FfiObject) freeRustArcPtr() {
	rustCall(func(status *C.RustCallStatus) int32 {
		ffiObject.freeFunction(ffiObject.pointer, status)
		return 0
	})
}

type BitcoinChainService interface {
	GetAddressUtxos(address string) ([]Utxo, error)
	GetTransactionStatus(txid string) (TxStatus, error)
	GetTransactionHex(txid string) (string, error)
	BroadcastTransaction(tx string) error
	RecommendedFees() (RecommendedFees, error)
}
type BitcoinChainServiceImpl struct {
	ffiObject FfiObject
}

func (_self *BitcoinChainServiceImpl) GetAddressUtxos(address string) ([]Utxo, error) {
	_pointer := _self.ffiObject.incrementPointer("BitcoinChainService")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ChainServiceError](
		FfiConverterChainServiceErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) []Utxo {
			return FfiConverterSequenceUtxoINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_bitcoinchainservice_get_address_utxos(
			_pointer, FfiConverterStringINSTANCE.Lower(address)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BitcoinChainServiceImpl) GetTransactionStatus(txid string) (TxStatus, error) {
	_pointer := _self.ffiObject.incrementPointer("BitcoinChainService")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ChainServiceError](
		FfiConverterChainServiceErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) TxStatus {
			return FfiConverterTxStatusINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_bitcoinchainservice_get_transaction_status(
			_pointer, FfiConverterStringINSTANCE.Lower(txid)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BitcoinChainServiceImpl) GetTransactionHex(txid string) (string, error) {
	_pointer := _self.ffiObject.incrementPointer("BitcoinChainService")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ChainServiceError](
		FfiConverterChainServiceErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) string {
			return FfiConverterStringINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_bitcoinchainservice_get_transaction_hex(
			_pointer, FfiConverterStringINSTANCE.Lower(txid)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BitcoinChainServiceImpl) BroadcastTransaction(tx string) error {
	_pointer := _self.ffiObject.incrementPointer("BitcoinChainService")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[ChainServiceError](
		FfiConverterChainServiceErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_bitcoinchainservice_broadcast_transaction(
			_pointer, FfiConverterStringINSTANCE.Lower(tx)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

func (_self *BitcoinChainServiceImpl) RecommendedFees() (RecommendedFees, error) {
	_pointer := _self.ffiObject.incrementPointer("BitcoinChainService")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ChainServiceError](
		FfiConverterChainServiceErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) RecommendedFees {
			return FfiConverterRecommendedFeesINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_bitcoinchainservice_recommended_fees(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}
func (object *BitcoinChainServiceImpl) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterBitcoinChainService struct {
	handleMap *concurrentHandleMap[BitcoinChainService]
}

var FfiConverterBitcoinChainServiceINSTANCE = FfiConverterBitcoinChainService{
	handleMap: newConcurrentHandleMap[BitcoinChainService](),
}

func (c FfiConverterBitcoinChainService) Lift(pointer unsafe.Pointer) BitcoinChainService {
	result := &BitcoinChainServiceImpl{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_bitcoinchainservice(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_bitcoinchainservice(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*BitcoinChainServiceImpl).Destroy)
	return result
}

func (c FfiConverterBitcoinChainService) Read(reader io.Reader) BitcoinChainService {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterBitcoinChainService) Lower(value BitcoinChainService) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := unsafe.Pointer(uintptr(c.handleMap.insert(value)))
	return pointer

}

func (c FfiConverterBitcoinChainService) Write(writer io.Writer, value BitcoinChainService) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerBitcoinChainService struct{}

func (_ FfiDestroyerBitcoinChainService) Destroy(value BitcoinChainService) {
	if val, ok := value.(*BitcoinChainServiceImpl); ok {
		val.Destroy()
	} else {
		panic("Expected *BitcoinChainServiceImpl")
	}
}

type uniffiCallbackResult C.int8_t

const (
	uniffiIdxCallbackFree               uniffiCallbackResult = 0
	uniffiCallbackResultSuccess         uniffiCallbackResult = 0
	uniffiCallbackResultError           uniffiCallbackResult = 1
	uniffiCallbackUnexpectedResultError uniffiCallbackResult = 2
	uniffiCallbackCancelled             uniffiCallbackResult = 3
)

type concurrentHandleMap[T any] struct {
	handles       map[uint64]T
	currentHandle uint64
	lock          sync.RWMutex
}

func newConcurrentHandleMap[T any]() *concurrentHandleMap[T] {
	return &concurrentHandleMap[T]{
		handles: map[uint64]T{},
	}
}

func (cm *concurrentHandleMap[T]) insert(obj T) uint64 {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.currentHandle = cm.currentHandle + 1
	cm.handles[cm.currentHandle] = obj
	return cm.currentHandle
}

func (cm *concurrentHandleMap[T]) remove(handle uint64) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	delete(cm.handles, handle)
}

func (cm *concurrentHandleMap[T]) tryGet(handle uint64) (T, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.handles[handle]
	return val, ok
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod0
func breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod0(uniffiHandle C.uint64_t, address C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterBitcoinChainServiceINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetAddressUtxos(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: address,
				}),
			)

		if err != nil {
			var actualError *ChainServiceError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterChainServiceErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterSequenceUtxoINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod1
func breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod1(uniffiHandle C.uint64_t, txid C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterBitcoinChainServiceINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetTransactionStatus(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: txid,
				}),
			)

		if err != nil {
			var actualError *ChainServiceError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterChainServiceErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterTxStatusINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod2
func breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod2(uniffiHandle C.uint64_t, txid C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterBitcoinChainServiceINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetTransactionHex(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: txid,
				}),
			)

		if err != nil {
			var actualError *ChainServiceError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterChainServiceErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterStringINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod3
func breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod3(uniffiHandle C.uint64_t, tx C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterBitcoinChainServiceINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.BroadcastTransaction(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: tx,
				}),
			)

		if err != nil {
			var actualError *ChainServiceError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterChainServiceErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod4
func breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod4(uniffiHandle C.uint64_t, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterBitcoinChainServiceINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.RecommendedFees()

		if err != nil {
			var actualError *ChainServiceError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterChainServiceErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterRecommendedFeesINSTANCE.Lower(res)
	}()
}

var UniffiVTableCallbackInterfaceBitcoinChainServiceINSTANCE = C.UniffiVTableCallbackInterfaceBitcoinChainService{
	getAddressUtxos:      (C.UniffiCallbackInterfaceBitcoinChainServiceMethod0)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod0),
	getTransactionStatus: (C.UniffiCallbackInterfaceBitcoinChainServiceMethod1)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod1),
	getTransactionHex:    (C.UniffiCallbackInterfaceBitcoinChainServiceMethod2)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod2),
	broadcastTransaction: (C.UniffiCallbackInterfaceBitcoinChainServiceMethod3)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod3),
	recommendedFees:      (C.UniffiCallbackInterfaceBitcoinChainServiceMethod4)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceMethod4),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceFree),
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceFree
func breez_sdk_spark_cgo_dispatchCallbackInterfaceBitcoinChainServiceFree(handle C.uint64_t) {
	FfiConverterBitcoinChainServiceINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterBitcoinChainService) register() {
	C.uniffi_breez_sdk_spark_fn_init_callback_vtable_bitcoinchainservice(&UniffiVTableCallbackInterfaceBitcoinChainServiceINSTANCE)
}

// `BreezSDK` is a wrapper around `SparkSDK` that provides a more structured API
// with request/response objects and comprehensive error handling.
type BreezSdkInterface interface {
	// Registers a listener to receive SDK events
	//
	// # Arguments
	//
	// * `listener` - An implementation of the `EventListener` trait
	//
	// # Returns
	//
	// A unique identifier for the listener, which can be used to remove it later
	AddEventListener(listener EventListener) string
	CheckLightningAddressAvailable(req CheckLightningAddressRequest) (bool, error)
	// Verifies a message signature against the provided public key. The message
	// is SHA256 hashed before verification. The signature can be hex encoded
	// in either DER or compact format.
	CheckMessage(request CheckMessageRequest) (CheckMessageResponse, error)
	ClaimDeposit(request ClaimDepositRequest) (ClaimDepositResponse, error)
	ClaimHtlcPayment(request ClaimHtlcPaymentRequest) (ClaimHtlcPaymentResponse, error)
	DeleteLightningAddress() error
	// Stops the SDK's background tasks
	//
	// This method stops the background tasks started by the `start()` method.
	// It should be called before your application terminates to ensure proper cleanup.
	//
	// # Returns
	//
	// Result containing either success or an `SdkError` if the background task couldn't be stopped
	Disconnect() error
	// Returns the balance of the wallet in satoshis
	GetInfo(request GetInfoRequest) (GetInfoResponse, error)
	GetLightningAddress() (*LightningAddressInfo, error)
	GetPayment(request GetPaymentRequest) (GetPaymentResponse, error)
	// Returns an instance of the [`TokenIssuer`] for managing token issuance.
	GetTokenIssuer() *TokenIssuer
	// Returns the metadata for the given token identifiers.
	//
	// Results are not guaranteed to be in the same order as the input token identifiers.
	//
	// If the metadata is not found locally in cache, it will be queried from
	// the Spark network and then cached.
	GetTokensMetadata(request GetTokensMetadataRequest) (GetTokensMetadataResponse, error)
	// Returns the user settings for the wallet.
	//
	// Some settings are fetched from the Spark network so network requests are performed.
	GetUserSettings() (UserSettings, error)
	// List fiat currencies for which there is a known exchange rate,
	// sorted by the canonical name of the currency.
	ListFiatCurrencies() (ListFiatCurrenciesResponse, error)
	// List the latest rates of fiat currencies, sorted by name.
	ListFiatRates() (ListFiatRatesResponse, error)
	// Lists payments from the storage with pagination
	//
	// This method provides direct access to the payment history stored in the database.
	// It returns payments in reverse chronological order (newest first).
	//
	// # Arguments
	//
	// * `request` - Contains pagination parameters (offset and limit)
	//
	// # Returns
	//
	// * `Ok(ListPaymentsResponse)` - Contains the list of payments if successful
	// * `Err(SdkError)` - If there was an error accessing the storage

	ListPayments(request ListPaymentsRequest) (ListPaymentsResponse, error)
	ListUnclaimedDeposits(request ListUnclaimedDepositsRequest) (ListUnclaimedDepositsResponse, error)
	LnurlPay(request LnurlPayRequest) (LnurlPayResponse, error)
	// Performs an LNURL withdraw operation for the amount of satoshis to
	// withdraw and the LNURL withdraw request details. The LNURL withdraw request
	// details can be obtained from calling [`BreezSdk::parse`].
	//
	// The method generates a Lightning invoice for the withdraw amount, stores
	// the LNURL withdraw metadata, and performs the LNURL withdraw using  the generated
	// invoice.
	//
	// If the `completion_timeout_secs` parameter is provided and greater than 0, the
	// method will wait for the payment to be completed within that period. If the
	// withdraw is completed within the timeout, the `payment` field in the response
	// will be set with the payment details. If the `completion_timeout_secs`
	// parameter is not provided or set to 0, the method will not wait for the payment
	// to be completed. If the withdraw is not completed within the
	// timeout, the `payment` field will be empty.
	//
	// # Arguments
	//
	// * `request` - The LNURL withdraw request
	//
	// # Returns
	//
	// Result containing either:
	// * `LnurlWithdrawResponse` - The payment details if the withdraw request was successful
	// * `SdkError` - If there was an error during the withdraw process
	LnurlWithdraw(request LnurlWithdrawRequest) (LnurlWithdrawResponse, error)
	Parse(input string) (InputType, error)
	PrepareLnurlPay(request PrepareLnurlPayRequest) (PrepareLnurlPayResponse, error)
	PrepareSendPayment(request PrepareSendPaymentRequest) (PrepareSendPaymentResponse, error)
	ReceivePayment(request ReceivePaymentRequest) (ReceivePaymentResponse, error)
	// Get the recommended BTC fees based on the configured chain service.
	RecommendedFees() (RecommendedFees, error)
	RefundDeposit(request RefundDepositRequest) (RefundDepositResponse, error)
	RegisterLightningAddress(request RegisterLightningAddressRequest) (LightningAddressInfo, error)
	// Removes a previously registered event listener
	//
	// # Arguments
	//
	// * `id` - The listener ID returned from `add_event_listener`
	//
	// # Returns
	//
	// `true` if the listener was found and removed, `false` otherwise
	RemoveEventListener(id string) bool
	SendPayment(request SendPaymentRequest) (SendPaymentResponse, error)
	// Signs a message with the wallet's identity key. The message is SHA256
	// hashed before signing. The returned signature will be hex encoded in
	// DER format by default, or compact format if specified.
	SignMessage(request SignMessageRequest) (SignMessageResponse, error)
	// Synchronizes the wallet with the Spark network
	SyncWallet(request SyncWalletRequest) (SyncWalletResponse, error)
	// Updates the user settings for the wallet.
	//
	// Some settings are updated on the Spark network so network requests may be performed.
	UpdateUserSettings(request UpdateUserSettingsRequest) error
}

// `BreezSDK` is a wrapper around `SparkSDK` that provides a more structured API
// with request/response objects and comprehensive error handling.
type BreezSdk struct {
	ffiObject FfiObject
}

// Registers a listener to receive SDK events
//
// # Arguments
//
// * `listener` - An implementation of the `EventListener` trait
//
// # Returns
//
// A unique identifier for the listener, which can be used to remove it later
func (_self *BreezSdk) AddEventListener(listener EventListener) string {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, _ := uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) string {
			return FfiConverterStringINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_add_event_listener(
			_pointer, FfiConverterCallbackInterfaceEventListenerINSTANCE.Lower(listener)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res
}

func (_self *BreezSdk) CheckLightningAddressAvailable(req CheckLightningAddressRequest) (bool, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) C.int8_t {
			res := C.ffi_breez_sdk_spark_rust_future_complete_i8(handle, status)
			return res
		},
		// liftFn
		func(ffi C.int8_t) bool {
			return FfiConverterBoolINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_check_lightning_address_available(
			_pointer, FfiConverterCheckLightningAddressRequestINSTANCE.Lower(req)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_i8(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_i8(handle)
		},
	)

	return res, err
}

// Verifies a message signature against the provided public key. The message
// is SHA256 hashed before verification. The signature can be hex encoded
// in either DER or compact format.
func (_self *BreezSdk) CheckMessage(request CheckMessageRequest) (CheckMessageResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) CheckMessageResponse {
			return FfiConverterCheckMessageResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_check_message(
			_pointer, FfiConverterCheckMessageRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) ClaimDeposit(request ClaimDepositRequest) (ClaimDepositResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) ClaimDepositResponse {
			return FfiConverterClaimDepositResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_claim_deposit(
			_pointer, FfiConverterClaimDepositRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) ClaimHtlcPayment(request ClaimHtlcPaymentRequest) (ClaimHtlcPaymentResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) ClaimHtlcPaymentResponse {
			return FfiConverterClaimHtlcPaymentResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_claim_htlc_payment(
			_pointer, FfiConverterClaimHtlcPaymentRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) DeleteLightningAddress() error {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_delete_lightning_address(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Stops the SDK's background tasks
//
// This method stops the background tasks started by the `start()` method.
// It should be called before your application terminates to ensure proper cleanup.
//
// # Returns
//
// Result containing either success or an `SdkError` if the background task couldn't be stopped
func (_self *BreezSdk) Disconnect() error {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_disconnect(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Returns the balance of the wallet in satoshis
func (_self *BreezSdk) GetInfo(request GetInfoRequest) (GetInfoResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) GetInfoResponse {
			return FfiConverterGetInfoResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_get_info(
			_pointer, FfiConverterGetInfoRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) GetLightningAddress() (*LightningAddressInfo, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) *LightningAddressInfo {
			return FfiConverterOptionalLightningAddressInfoINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_get_lightning_address(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) GetPayment(request GetPaymentRequest) (GetPaymentResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) GetPaymentResponse {
			return FfiConverterGetPaymentResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_get_payment(
			_pointer, FfiConverterGetPaymentRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Returns an instance of the [`TokenIssuer`] for managing token issuance.
func (_self *BreezSdk) GetTokenIssuer() *TokenIssuer {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTokenIssuerINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_breez_sdk_spark_fn_method_breezsdk_get_token_issuer(
			_pointer, _uniffiStatus)
	}))
}

// Returns the metadata for the given token identifiers.
//
// Results are not guaranteed to be in the same order as the input token identifiers.
//
// If the metadata is not found locally in cache, it will be queried from
// the Spark network and then cached.
func (_self *BreezSdk) GetTokensMetadata(request GetTokensMetadataRequest) (GetTokensMetadataResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) GetTokensMetadataResponse {
			return FfiConverterGetTokensMetadataResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_get_tokens_metadata(
			_pointer, FfiConverterGetTokensMetadataRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Returns the user settings for the wallet.
//
// Some settings are fetched from the Spark network so network requests are performed.
func (_self *BreezSdk) GetUserSettings() (UserSettings, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) UserSettings {
			return FfiConverterUserSettingsINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_get_user_settings(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// List fiat currencies for which there is a known exchange rate,
// sorted by the canonical name of the currency.
func (_self *BreezSdk) ListFiatCurrencies() (ListFiatCurrenciesResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) ListFiatCurrenciesResponse {
			return FfiConverterListFiatCurrenciesResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_list_fiat_currencies(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// List the latest rates of fiat currencies, sorted by name.
func (_self *BreezSdk) ListFiatRates() (ListFiatRatesResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) ListFiatRatesResponse {
			return FfiConverterListFiatRatesResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_list_fiat_rates(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Lists payments from the storage with pagination
//
// This method provides direct access to the payment history stored in the database.
// It returns payments in reverse chronological order (newest first).
//
// # Arguments
//
// * `request` - Contains pagination parameters (offset and limit)
//
// # Returns
//
// * `Ok(ListPaymentsResponse)` - Contains the list of payments if successful
// * `Err(SdkError)` - If there was an error accessing the storage

func (_self *BreezSdk) ListPayments(request ListPaymentsRequest) (ListPaymentsResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) ListPaymentsResponse {
			return FfiConverterListPaymentsResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_list_payments(
			_pointer, FfiConverterListPaymentsRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) ListUnclaimedDeposits(request ListUnclaimedDepositsRequest) (ListUnclaimedDepositsResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) ListUnclaimedDepositsResponse {
			return FfiConverterListUnclaimedDepositsResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_list_unclaimed_deposits(
			_pointer, FfiConverterListUnclaimedDepositsRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) LnurlPay(request LnurlPayRequest) (LnurlPayResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) LnurlPayResponse {
			return FfiConverterLnurlPayResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_lnurl_pay(
			_pointer, FfiConverterLnurlPayRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Performs an LNURL withdraw operation for the amount of satoshis to
// withdraw and the LNURL withdraw request details. The LNURL withdraw request
// details can be obtained from calling [`BreezSdk::parse`].
//
// The method generates a Lightning invoice for the withdraw amount, stores
// the LNURL withdraw metadata, and performs the LNURL withdraw using  the generated
// invoice.
//
// If the `completion_timeout_secs` parameter is provided and greater than 0, the
// method will wait for the payment to be completed within that period. If the
// withdraw is completed within the timeout, the `payment` field in the response
// will be set with the payment details. If the `completion_timeout_secs`
// parameter is not provided or set to 0, the method will not wait for the payment
// to be completed. If the withdraw is not completed within the
// timeout, the `payment` field will be empty.
//
// # Arguments
//
// * `request` - The LNURL withdraw request
//
// # Returns
//
// Result containing either:
// * `LnurlWithdrawResponse` - The payment details if the withdraw request was successful
// * `SdkError` - If there was an error during the withdraw process
func (_self *BreezSdk) LnurlWithdraw(request LnurlWithdrawRequest) (LnurlWithdrawResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) LnurlWithdrawResponse {
			return FfiConverterLnurlWithdrawResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_lnurl_withdraw(
			_pointer, FfiConverterLnurlWithdrawRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) Parse(input string) (InputType, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) InputType {
			return FfiConverterInputTypeINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_parse(
			_pointer, FfiConverterStringINSTANCE.Lower(input)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) PrepareLnurlPay(request PrepareLnurlPayRequest) (PrepareLnurlPayResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) PrepareLnurlPayResponse {
			return FfiConverterPrepareLnurlPayResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_prepare_lnurl_pay(
			_pointer, FfiConverterPrepareLnurlPayRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) PrepareSendPayment(request PrepareSendPaymentRequest) (PrepareSendPaymentResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) PrepareSendPaymentResponse {
			return FfiConverterPrepareSendPaymentResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_prepare_send_payment(
			_pointer, FfiConverterPrepareSendPaymentRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) ReceivePayment(request ReceivePaymentRequest) (ReceivePaymentResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) ReceivePaymentResponse {
			return FfiConverterReceivePaymentResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_receive_payment(
			_pointer, FfiConverterReceivePaymentRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Get the recommended BTC fees based on the configured chain service.
func (_self *BreezSdk) RecommendedFees() (RecommendedFees, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) RecommendedFees {
			return FfiConverterRecommendedFeesINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_recommended_fees(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) RefundDeposit(request RefundDepositRequest) (RefundDepositResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) RefundDepositResponse {
			return FfiConverterRefundDepositResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_refund_deposit(
			_pointer, FfiConverterRefundDepositRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *BreezSdk) RegisterLightningAddress(request RegisterLightningAddressRequest) (LightningAddressInfo, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) LightningAddressInfo {
			return FfiConverterLightningAddressInfoINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_register_lightning_address(
			_pointer, FfiConverterRegisterLightningAddressRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Removes a previously registered event listener
//
// # Arguments
//
// * `id` - The listener ID returned from `add_event_listener`
//
// # Returns
//
// `true` if the listener was found and removed, `false` otherwise
func (_self *BreezSdk) RemoveEventListener(id string) bool {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, _ := uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) C.int8_t {
			res := C.ffi_breez_sdk_spark_rust_future_complete_i8(handle, status)
			return res
		},
		// liftFn
		func(ffi C.int8_t) bool {
			return FfiConverterBoolINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_remove_event_listener(
			_pointer, FfiConverterStringINSTANCE.Lower(id)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_i8(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_i8(handle)
		},
	)

	return res
}

func (_self *BreezSdk) SendPayment(request SendPaymentRequest) (SendPaymentResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) SendPaymentResponse {
			return FfiConverterSendPaymentResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_send_payment(
			_pointer, FfiConverterSendPaymentRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Signs a message with the wallet's identity key. The message is SHA256
// hashed before signing. The returned signature will be hex encoded in
// DER format by default, or compact format if specified.
func (_self *BreezSdk) SignMessage(request SignMessageRequest) (SignMessageResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) SignMessageResponse {
			return FfiConverterSignMessageResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_sign_message(
			_pointer, FfiConverterSignMessageRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Synchronizes the wallet with the Spark network
func (_self *BreezSdk) SyncWallet(request SyncWalletRequest) (SyncWalletResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) SyncWalletResponse {
			return FfiConverterSyncWalletResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_sync_wallet(
			_pointer, FfiConverterSyncWalletRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Updates the user settings for the wallet.
//
// Some settings are updated on the Spark network so network requests may be performed.
func (_self *BreezSdk) UpdateUserSettings(request UpdateUserSettingsRequest) error {
	_pointer := _self.ffiObject.incrementPointer("*BreezSdk")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_breezsdk_update_user_settings(
			_pointer, FfiConverterUpdateUserSettingsRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}
func (object *BreezSdk) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterBreezSdk struct{}

var FfiConverterBreezSdkINSTANCE = FfiConverterBreezSdk{}

func (c FfiConverterBreezSdk) Lift(pointer unsafe.Pointer) *BreezSdk {
	result := &BreezSdk{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_breezsdk(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_breezsdk(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*BreezSdk).Destroy)
	return result
}

func (c FfiConverterBreezSdk) Read(reader io.Reader) *BreezSdk {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterBreezSdk) Lower(value *BreezSdk) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*BreezSdk")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterBreezSdk) Write(writer io.Writer, value *BreezSdk) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerBreezSdk struct{}

func (_ FfiDestroyerBreezSdk) Destroy(value *BreezSdk) {
	value.Destroy()
}

// Trait covering fiat-related functionality
type FiatService interface {
	// List all supported fiat currencies for which there is a known exchange rate.
	FetchFiatCurrencies() ([]FiatCurrency, error)
	// Get the live rates from the server.
	FetchFiatRates() ([]Rate, error)
}

// Trait covering fiat-related functionality
type FiatServiceImpl struct {
	ffiObject FfiObject
}

// List all supported fiat currencies for which there is a known exchange rate.
func (_self *FiatServiceImpl) FetchFiatCurrencies() ([]FiatCurrency, error) {
	_pointer := _self.ffiObject.incrementPointer("FiatService")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ServiceConnectivityError](
		FfiConverterServiceConnectivityErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) []FiatCurrency {
			return FfiConverterSequenceFiatCurrencyINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_fiatservice_fetch_fiat_currencies(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Get the live rates from the server.
func (_self *FiatServiceImpl) FetchFiatRates() ([]Rate, error) {
	_pointer := _self.ffiObject.incrementPointer("FiatService")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ServiceConnectivityError](
		FfiConverterServiceConnectivityErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) []Rate {
			return FfiConverterSequenceRateINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_fiatservice_fetch_fiat_rates(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}
func (object *FiatServiceImpl) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterFiatService struct {
	handleMap *concurrentHandleMap[FiatService]
}

var FfiConverterFiatServiceINSTANCE = FfiConverterFiatService{
	handleMap: newConcurrentHandleMap[FiatService](),
}

func (c FfiConverterFiatService) Lift(pointer unsafe.Pointer) FiatService {
	result := &FiatServiceImpl{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_fiatservice(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_fiatservice(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*FiatServiceImpl).Destroy)
	return result
}

func (c FfiConverterFiatService) Read(reader io.Reader) FiatService {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterFiatService) Lower(value FiatService) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := unsafe.Pointer(uintptr(c.handleMap.insert(value)))
	return pointer

}

func (c FfiConverterFiatService) Write(writer io.Writer, value FiatService) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerFiatService struct{}

func (_ FfiDestroyerFiatService) Destroy(value FiatService) {
	if val, ok := value.(*FiatServiceImpl); ok {
		val.Destroy()
	} else {
		panic("Expected *FiatServiceImpl")
	}
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceMethod0
func breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceMethod0(uniffiHandle C.uint64_t, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterFiatServiceINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.FetchFiatCurrencies()

		if err != nil {
			var actualError *ServiceConnectivityError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterServiceConnectivityErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterSequenceFiatCurrencyINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceMethod1
func breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceMethod1(uniffiHandle C.uint64_t, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterFiatServiceINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.FetchFiatRates()

		if err != nil {
			var actualError *ServiceConnectivityError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterServiceConnectivityErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterSequenceRateINSTANCE.Lower(res)
	}()
}

var UniffiVTableCallbackInterfaceFiatServiceINSTANCE = C.UniffiVTableCallbackInterfaceFiatService{
	fetchFiatCurrencies: (C.UniffiCallbackInterfaceFiatServiceMethod0)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceMethod0),
	fetchFiatRates:      (C.UniffiCallbackInterfaceFiatServiceMethod1)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceMethod1),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceFree),
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceFree
func breez_sdk_spark_cgo_dispatchCallbackInterfaceFiatServiceFree(handle C.uint64_t) {
	FfiConverterFiatServiceINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterFiatService) register() {
	C.uniffi_breez_sdk_spark_fn_init_callback_vtable_fiatservice(&UniffiVTableCallbackInterfaceFiatServiceINSTANCE)
}

// This interface is used to observe outgoing payments before Lightning, Spark and onchain Bitcoin payments.
// If the implementation returns an error, the payment is cancelled.
type PaymentObserver interface {
	// Called before Lightning, Spark or onchain Bitcoin payments are made
	BeforeSend(payments []ProvisionalPayment) error
}

// This interface is used to observe outgoing payments before Lightning, Spark and onchain Bitcoin payments.
// If the implementation returns an error, the payment is cancelled.
type PaymentObserverImpl struct {
	ffiObject FfiObject
}

// Called before Lightning, Spark or onchain Bitcoin payments are made
func (_self *PaymentObserverImpl) BeforeSend(payments []ProvisionalPayment) error {
	_pointer := _self.ffiObject.incrementPointer("PaymentObserver")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[PaymentObserverError](
		FfiConverterPaymentObserverErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_paymentobserver_before_send(
			_pointer, FfiConverterSequenceProvisionalPaymentINSTANCE.Lower(payments)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}
func (object *PaymentObserverImpl) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterPaymentObserver struct {
	handleMap *concurrentHandleMap[PaymentObserver]
}

var FfiConverterPaymentObserverINSTANCE = FfiConverterPaymentObserver{
	handleMap: newConcurrentHandleMap[PaymentObserver](),
}

func (c FfiConverterPaymentObserver) Lift(pointer unsafe.Pointer) PaymentObserver {
	result := &PaymentObserverImpl{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_paymentobserver(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_paymentobserver(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*PaymentObserverImpl).Destroy)
	return result
}

func (c FfiConverterPaymentObserver) Read(reader io.Reader) PaymentObserver {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterPaymentObserver) Lower(value PaymentObserver) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := unsafe.Pointer(uintptr(c.handleMap.insert(value)))
	return pointer

}

func (c FfiConverterPaymentObserver) Write(writer io.Writer, value PaymentObserver) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerPaymentObserver struct{}

func (_ FfiDestroyerPaymentObserver) Destroy(value PaymentObserver) {
	if val, ok := value.(*PaymentObserverImpl); ok {
		val.Destroy()
	} else {
		panic("Expected *PaymentObserverImpl")
	}
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfacePaymentObserverMethod0
func breez_sdk_spark_cgo_dispatchCallbackInterfacePaymentObserverMethod0(uniffiHandle C.uint64_t, payments C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterPaymentObserverINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.BeforeSend(
				FfiConverterSequenceProvisionalPaymentINSTANCE.Lift(GoRustBuffer{
					inner: payments,
				}),
			)

		if err != nil {
			var actualError *PaymentObserverError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterPaymentObserverErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

var UniffiVTableCallbackInterfacePaymentObserverINSTANCE = C.UniffiVTableCallbackInterfacePaymentObserver{
	beforeSend: (C.UniffiCallbackInterfacePaymentObserverMethod0)(C.breez_sdk_spark_cgo_dispatchCallbackInterfacePaymentObserverMethod0),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.breez_sdk_spark_cgo_dispatchCallbackInterfacePaymentObserverFree),
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfacePaymentObserverFree
func breez_sdk_spark_cgo_dispatchCallbackInterfacePaymentObserverFree(handle C.uint64_t) {
	FfiConverterPaymentObserverINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterPaymentObserver) register() {
	C.uniffi_breez_sdk_spark_fn_init_callback_vtable_paymentobserver(&UniffiVTableCallbackInterfacePaymentObserverINSTANCE)
}

type RestClient interface {
	// Makes a GET request and logs on DEBUG.
	// ### Arguments
	// - `url`: the URL on which GET will be called
	// - `headers`: optional headers that will be set on the request
	GetRequest(url string, headers *map[string]string) (RestResponse, error)
	// Makes a POST request, and logs on DEBUG.
	// ### Arguments
	// - `url`: the URL on which POST will be called
	// - `headers`: the optional POST headers
	// - `body`: the optional POST body
	PostRequest(url string, headers *map[string]string, body *string) (RestResponse, error)
	// Makes a DELETE request, and logs on DEBUG.
	// ### Arguments
	// - `url`: the URL on which DELETE will be called
	// - `headers`: the optional DELETE headers
	// - `body`: the optional DELETE body
	DeleteRequest(url string, headers *map[string]string, body *string) (RestResponse, error)
}
type RestClientImpl struct {
	ffiObject FfiObject
}

// Makes a GET request and logs on DEBUG.
// ### Arguments
// - `url`: the URL on which GET will be called
// - `headers`: optional headers that will be set on the request
func (_self *RestClientImpl) GetRequest(url string, headers *map[string]string) (RestResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("RestClient")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ServiceConnectivityError](
		FfiConverterServiceConnectivityErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) RestResponse {
			return FfiConverterRestResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_restclient_get_request(
			_pointer, FfiConverterStringINSTANCE.Lower(url), FfiConverterOptionalMapStringStringINSTANCE.Lower(headers)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Makes a POST request, and logs on DEBUG.
// ### Arguments
// - `url`: the URL on which POST will be called
// - `headers`: the optional POST headers
// - `body`: the optional POST body
func (_self *RestClientImpl) PostRequest(url string, headers *map[string]string, body *string) (RestResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("RestClient")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ServiceConnectivityError](
		FfiConverterServiceConnectivityErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) RestResponse {
			return FfiConverterRestResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_restclient_post_request(
			_pointer, FfiConverterStringINSTANCE.Lower(url), FfiConverterOptionalMapStringStringINSTANCE.Lower(headers), FfiConverterOptionalStringINSTANCE.Lower(body)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Makes a DELETE request, and logs on DEBUG.
// ### Arguments
// - `url`: the URL on which DELETE will be called
// - `headers`: the optional DELETE headers
// - `body`: the optional DELETE body
func (_self *RestClientImpl) DeleteRequest(url string, headers *map[string]string, body *string) (RestResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("RestClient")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[ServiceConnectivityError](
		FfiConverterServiceConnectivityErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) RestResponse {
			return FfiConverterRestResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_restclient_delete_request(
			_pointer, FfiConverterStringINSTANCE.Lower(url), FfiConverterOptionalMapStringStringINSTANCE.Lower(headers), FfiConverterOptionalStringINSTANCE.Lower(body)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}
func (object *RestClientImpl) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterRestClient struct {
	handleMap *concurrentHandleMap[RestClient]
}

var FfiConverterRestClientINSTANCE = FfiConverterRestClient{
	handleMap: newConcurrentHandleMap[RestClient](),
}

func (c FfiConverterRestClient) Lift(pointer unsafe.Pointer) RestClient {
	result := &RestClientImpl{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_restclient(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_restclient(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*RestClientImpl).Destroy)
	return result
}

func (c FfiConverterRestClient) Read(reader io.Reader) RestClient {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterRestClient) Lower(value RestClient) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := unsafe.Pointer(uintptr(c.handleMap.insert(value)))
	return pointer

}

func (c FfiConverterRestClient) Write(writer io.Writer, value RestClient) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerRestClient struct{}

func (_ FfiDestroyerRestClient) Destroy(value RestClient) {
	if val, ok := value.(*RestClientImpl); ok {
		val.Destroy()
	} else {
		panic("Expected *RestClientImpl")
	}
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod0
func breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod0(uniffiHandle C.uint64_t, url C.RustBuffer, headers C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterRestClientINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetRequest(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: url,
				}),
				FfiConverterOptionalMapStringStringINSTANCE.Lift(GoRustBuffer{
					inner: headers,
				}),
			)

		if err != nil {
			var actualError *ServiceConnectivityError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterServiceConnectivityErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterRestResponseINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod1
func breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod1(uniffiHandle C.uint64_t, url C.RustBuffer, headers C.RustBuffer, body C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterRestClientINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.PostRequest(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: url,
				}),
				FfiConverterOptionalMapStringStringINSTANCE.Lift(GoRustBuffer{
					inner: headers,
				}),
				FfiConverterOptionalStringINSTANCE.Lift(GoRustBuffer{
					inner: body,
				}),
			)

		if err != nil {
			var actualError *ServiceConnectivityError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterServiceConnectivityErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterRestResponseINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod2
func breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod2(uniffiHandle C.uint64_t, url C.RustBuffer, headers C.RustBuffer, body C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterRestClientINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.DeleteRequest(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: url,
				}),
				FfiConverterOptionalMapStringStringINSTANCE.Lift(GoRustBuffer{
					inner: headers,
				}),
				FfiConverterOptionalStringINSTANCE.Lift(GoRustBuffer{
					inner: body,
				}),
			)

		if err != nil {
			var actualError *ServiceConnectivityError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterServiceConnectivityErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterRestResponseINSTANCE.Lower(res)
	}()
}

var UniffiVTableCallbackInterfaceRestClientINSTANCE = C.UniffiVTableCallbackInterfaceRestClient{
	getRequest:    (C.UniffiCallbackInterfaceRestClientMethod0)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod0),
	postRequest:   (C.UniffiCallbackInterfaceRestClientMethod1)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod1),
	deleteRequest: (C.UniffiCallbackInterfaceRestClientMethod2)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientMethod2),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientFree),
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientFree
func breez_sdk_spark_cgo_dispatchCallbackInterfaceRestClientFree(handle C.uint64_t) {
	FfiConverterRestClientINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterRestClient) register() {
	C.uniffi_breez_sdk_spark_fn_init_callback_vtable_restclient(&UniffiVTableCallbackInterfaceRestClientINSTANCE)
}

// Builder for creating `BreezSdk` instances with customizable components.
type SdkBuilderInterface interface {
	// Builds the `BreezSdk` instance with the configured components.
	Build() (*BreezSdk, error)
	// Sets the chain service to be used by the SDK.
	// Arguments:
	// - `chain_service`: The chain service to be used.
	WithChainService(chainService BitcoinChainService)
	// Sets the root storage directory to initialize the default storage with.
	// This initializes both storage and real-time sync storage with the
	// default implementations.
	// Arguments:
	// - `storage_dir`: The data directory for storage.
	WithDefaultStorage(storageDir string)
	// Sets the fiat service to be used by the SDK.
	// Arguments:
	// - `fiat_service`: The fiat service to be used.
	WithFiatService(fiatService FiatService)
	// Sets the key set type to be used by the SDK.
	// Arguments:
	// - `key_set_type`: The key set type which determines the derivation path.
	// - `use_address_index`: Controls the structure of the BIP derivation path.
	WithKeySet(keySetType KeySetType, useAddressIndex bool, accountNumber *uint32)
	WithLnurlClient(lnurlClient RestClient)
	// Sets the payment observer to be used by the SDK.
	// Arguments:
	// - `payment_observer`: The payment observer to be used.
	WithPaymentObserver(paymentObserver PaymentObserver)
	// Sets the real-time sync storage implementation to be used by the SDK.
	// Arguments:
	// - `storage`: The sync storage implementation to be used.
	WithRealTimeSyncStorage(storage SyncStorage)
	// Sets the REST chain service to be used by the SDK.
	// Arguments:
	// - `url`: The base URL of the REST API.
	// - `api_type`: The API type to be used.
	// - `credentials`: Optional credentials for basic authentication.
	WithRestChainService(url string, apiType ChainApiType, credentials *Credentials)
	// Sets the storage implementation to be used by the SDK.
	// Arguments:
	// - `storage`: The storage implementation to be used.
	WithStorage(storage Storage)
}

// Builder for creating `BreezSdk` instances with customizable components.
type SdkBuilder struct {
	ffiObject FfiObject
}

// Creates a new `SdkBuilder` with the provided configuration.
// Arguments:
// - `config`: The configuration to be used.
// - `seed`: The seed for wallet generation.
func NewSdkBuilder(config Config, seed Seed) *SdkBuilder {
	return FfiConverterSdkBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_breez_sdk_spark_fn_constructor_sdkbuilder_new(FfiConverterConfigINSTANCE.Lower(config), FfiConverterSeedINSTANCE.Lower(seed), _uniffiStatus)
	}))
}

// Builds the `BreezSdk` instance with the configured components.
func (_self *SdkBuilder) Build() (*BreezSdk, error) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) unsafe.Pointer {
			res := C.ffi_breez_sdk_spark_rust_future_complete_pointer(handle, status)
			return res
		},
		// liftFn
		func(ffi unsafe.Pointer) *BreezSdk {
			return FfiConverterBreezSdkINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_build(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_pointer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_pointer(handle)
		},
	)

	return res, err
}

// Sets the chain service to be used by the SDK.
// Arguments:
// - `chain_service`: The chain service to be used.
func (_self *SdkBuilder) WithChainService(chainService BitcoinChainService) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_chain_service(
			_pointer, FfiConverterBitcoinChainServiceINSTANCE.Lower(chainService)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}

// Sets the root storage directory to initialize the default storage with.
// This initializes both storage and real-time sync storage with the
// default implementations.
// Arguments:
// - `storage_dir`: The data directory for storage.
func (_self *SdkBuilder) WithDefaultStorage(storageDir string) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_default_storage(
			_pointer, FfiConverterStringINSTANCE.Lower(storageDir)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}

// Sets the fiat service to be used by the SDK.
// Arguments:
// - `fiat_service`: The fiat service to be used.
func (_self *SdkBuilder) WithFiatService(fiatService FiatService) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_fiat_service(
			_pointer, FfiConverterFiatServiceINSTANCE.Lower(fiatService)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}

// Sets the key set type to be used by the SDK.
// Arguments:
// - `key_set_type`: The key set type which determines the derivation path.
// - `use_address_index`: Controls the structure of the BIP derivation path.
func (_self *SdkBuilder) WithKeySet(keySetType KeySetType, useAddressIndex bool, accountNumber *uint32) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_key_set(
			_pointer, FfiConverterKeySetTypeINSTANCE.Lower(keySetType), FfiConverterBoolINSTANCE.Lower(useAddressIndex), FfiConverterOptionalUint32INSTANCE.Lower(accountNumber)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}

func (_self *SdkBuilder) WithLnurlClient(lnurlClient RestClient) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_lnurl_client(
			_pointer, FfiConverterRestClientINSTANCE.Lower(lnurlClient)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}

// Sets the payment observer to be used by the SDK.
// Arguments:
// - `payment_observer`: The payment observer to be used.
func (_self *SdkBuilder) WithPaymentObserver(paymentObserver PaymentObserver) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_payment_observer(
			_pointer, FfiConverterPaymentObserverINSTANCE.Lower(paymentObserver)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}

// Sets the real-time sync storage implementation to be used by the SDK.
// Arguments:
// - `storage`: The sync storage implementation to be used.
func (_self *SdkBuilder) WithRealTimeSyncStorage(storage SyncStorage) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_real_time_sync_storage(
			_pointer, FfiConverterSyncStorageINSTANCE.Lower(storage)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}

// Sets the REST chain service to be used by the SDK.
// Arguments:
// - `url`: The base URL of the REST API.
// - `api_type`: The API type to be used.
// - `credentials`: Optional credentials for basic authentication.
func (_self *SdkBuilder) WithRestChainService(url string, apiType ChainApiType, credentials *Credentials) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_rest_chain_service(
			_pointer, FfiConverterStringINSTANCE.Lower(url), FfiConverterChainApiTypeINSTANCE.Lower(apiType), FfiConverterOptionalCredentialsINSTANCE.Lower(credentials)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}

// Sets the storage implementation to be used by the SDK.
// Arguments:
// - `storage`: The storage implementation to be used.
func (_self *SdkBuilder) WithStorage(storage Storage) {
	_pointer := _self.ffiObject.incrementPointer("*SdkBuilder")
	defer _self.ffiObject.decrementPointer()
	uniffiRustCallAsync[error](
		nil,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_sdkbuilder_with_storage(
			_pointer, FfiConverterStorageINSTANCE.Lower(storage)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

}
func (object *SdkBuilder) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterSdkBuilder struct{}

var FfiConverterSdkBuilderINSTANCE = FfiConverterSdkBuilder{}

func (c FfiConverterSdkBuilder) Lift(pointer unsafe.Pointer) *SdkBuilder {
	result := &SdkBuilder{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_sdkbuilder(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_sdkbuilder(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*SdkBuilder).Destroy)
	return result
}

func (c FfiConverterSdkBuilder) Read(reader io.Reader) *SdkBuilder {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterSdkBuilder) Lower(value *SdkBuilder) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*SdkBuilder")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterSdkBuilder) Write(writer io.Writer, value *SdkBuilder) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerSdkBuilder struct{}

func (_ FfiDestroyerSdkBuilder) Destroy(value *SdkBuilder) {
	value.Destroy()
}

// Trait for persistent storage
type Storage interface {
	DeleteCachedItem(key string) error
	GetCachedItem(key string) (*string, error)
	SetCachedItem(key string, value string) error
	// Lists payments with optional filters and pagination
	//
	// # Arguments
	//
	// * `list_payments_request` - The request to list payments
	//
	// # Returns
	//
	// A vector of payments or a `StorageError`
	ListPayments(request ListPaymentsRequest) ([]Payment, error)
	// Inserts a payment into storage
	//
	// # Arguments
	//
	// * `payment` - The payment to insert
	//
	// # Returns
	//
	// Success or a `StorageError`
	InsertPayment(payment Payment) error
	// Inserts payment metadata into storage
	//
	// # Arguments
	//
	// * `payment_id` - The ID of the payment
	// * `metadata` - The metadata to insert
	//
	// # Returns
	//
	// Success or a `StorageError`
	SetPaymentMetadata(paymentId string, metadata PaymentMetadata) error
	// Gets a payment by its ID
	// # Arguments
	//
	// * `id` - The ID of the payment to retrieve
	//
	// # Returns
	//
	// The payment if found or None if not found
	GetPaymentById(id string) (Payment, error)
	// Gets a payment by its invoice
	// # Arguments
	//
	// * `invoice` - The invoice of the payment to retrieve
	// # Returns
	//
	// The payment if found or None if not found
	GetPaymentByInvoice(invoice string) (*Payment, error)
	// Add a deposit to storage
	// # Arguments
	//
	// * `txid` - The transaction ID of the deposit
	// * `vout` - The output index of the deposit
	// * `amount_sats` - The amount of the deposit in sats
	//
	// # Returns
	//
	// Success or a `StorageError`
	AddDeposit(txid string, vout uint32, amountSats uint64) error
	// Removes an unclaimed deposit from storage
	// # Arguments
	//
	// * `txid` - The transaction ID of the deposit
	// * `vout` - The output index of the deposit
	//
	// # Returns
	//
	// Success or a `StorageError`
	DeleteDeposit(txid string, vout uint32) error
	// Lists all unclaimed deposits from storage
	// # Returns
	//
	// A vector of `DepositInfo` or a `StorageError`
	ListDeposits() ([]DepositInfo, error)
	// Updates or inserts unclaimed deposit details
	// # Arguments
	//
	// * `txid` - The transaction ID of the deposit
	// * `vout` - The output index of the deposit
	// * `payload` - The payload for the update
	//
	// # Returns
	//
	// Success or a `StorageError`
	UpdateDeposit(txid string, vout uint32, payload UpdateDepositPayload) error
	SetLnurlMetadata(metadata []SetLnurlMetadataItem) error
}

// Trait for persistent storage
type StorageImpl struct {
	ffiObject FfiObject
}

func (_self *StorageImpl) DeleteCachedItem(key string) error {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_storage_delete_cached_item(
			_pointer, FfiConverterStringINSTANCE.Lower(key)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

func (_self *StorageImpl) GetCachedItem(key string) (*string, error) {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) *string {
			return FfiConverterOptionalStringINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_storage_get_cached_item(
			_pointer, FfiConverterStringINSTANCE.Lower(key)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

func (_self *StorageImpl) SetCachedItem(key string, value string) error {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_storage_set_cached_item(
			_pointer, FfiConverterStringINSTANCE.Lower(key), FfiConverterStringINSTANCE.Lower(value)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Lists payments with optional filters and pagination
//
// # Arguments
//
// * `list_payments_request` - The request to list payments
//
// # Returns
//
// A vector of payments or a `StorageError`
func (_self *StorageImpl) ListPayments(request ListPaymentsRequest) ([]Payment, error) {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) []Payment {
			return FfiConverterSequencePaymentINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_storage_list_payments(
			_pointer, FfiConverterListPaymentsRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Inserts a payment into storage
//
// # Arguments
//
// * `payment` - The payment to insert
//
// # Returns
//
// Success or a `StorageError`
func (_self *StorageImpl) InsertPayment(payment Payment) error {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_storage_insert_payment(
			_pointer, FfiConverterPaymentINSTANCE.Lower(payment)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Inserts payment metadata into storage
//
// # Arguments
//
// * `payment_id` - The ID of the payment
// * `metadata` - The metadata to insert
//
// # Returns
//
// Success or a `StorageError`
func (_self *StorageImpl) SetPaymentMetadata(paymentId string, metadata PaymentMetadata) error {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_storage_set_payment_metadata(
			_pointer, FfiConverterStringINSTANCE.Lower(paymentId), FfiConverterPaymentMetadataINSTANCE.Lower(metadata)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Gets a payment by its ID
// # Arguments
//
// * `id` - The ID of the payment to retrieve
//
// # Returns
//
// The payment if found or None if not found
func (_self *StorageImpl) GetPaymentById(id string) (Payment, error) {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) Payment {
			return FfiConverterPaymentINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_storage_get_payment_by_id(
			_pointer, FfiConverterStringINSTANCE.Lower(id)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Gets a payment by its invoice
// # Arguments
//
// * `invoice` - The invoice of the payment to retrieve
// # Returns
//
// The payment if found or None if not found
func (_self *StorageImpl) GetPaymentByInvoice(invoice string) (*Payment, error) {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) *Payment {
			return FfiConverterOptionalPaymentINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_storage_get_payment_by_invoice(
			_pointer, FfiConverterStringINSTANCE.Lower(invoice)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Add a deposit to storage
// # Arguments
//
// * `txid` - The transaction ID of the deposit
// * `vout` - The output index of the deposit
// * `amount_sats` - The amount of the deposit in sats
//
// # Returns
//
// Success or a `StorageError`
func (_self *StorageImpl) AddDeposit(txid string, vout uint32, amountSats uint64) error {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_storage_add_deposit(
			_pointer, FfiConverterStringINSTANCE.Lower(txid), FfiConverterUint32INSTANCE.Lower(vout), FfiConverterUint64INSTANCE.Lower(amountSats)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Removes an unclaimed deposit from storage
// # Arguments
//
// * `txid` - The transaction ID of the deposit
// * `vout` - The output index of the deposit
//
// # Returns
//
// Success or a `StorageError`
func (_self *StorageImpl) DeleteDeposit(txid string, vout uint32) error {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_storage_delete_deposit(
			_pointer, FfiConverterStringINSTANCE.Lower(txid), FfiConverterUint32INSTANCE.Lower(vout)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Lists all unclaimed deposits from storage
// # Returns
//
// A vector of `DepositInfo` or a `StorageError`
func (_self *StorageImpl) ListDeposits() ([]DepositInfo, error) {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) []DepositInfo {
			return FfiConverterSequenceDepositInfoINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_storage_list_deposits(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Updates or inserts unclaimed deposit details
// # Arguments
//
// * `txid` - The transaction ID of the deposit
// * `vout` - The output index of the deposit
// * `payload` - The payload for the update
//
// # Returns
//
// Success or a `StorageError`
func (_self *StorageImpl) UpdateDeposit(txid string, vout uint32, payload UpdateDepositPayload) error {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_storage_update_deposit(
			_pointer, FfiConverterStringINSTANCE.Lower(txid), FfiConverterUint32INSTANCE.Lower(vout), FfiConverterUpdateDepositPayloadINSTANCE.Lower(payload)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

func (_self *StorageImpl) SetLnurlMetadata(metadata []SetLnurlMetadataItem) error {
	_pointer := _self.ffiObject.incrementPointer("Storage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[StorageError](
		FfiConverterStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_storage_set_lnurl_metadata(
			_pointer, FfiConverterSequenceSetLnurlMetadataItemINSTANCE.Lower(metadata)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}
func (object *StorageImpl) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterStorage struct {
	handleMap *concurrentHandleMap[Storage]
}

var FfiConverterStorageINSTANCE = FfiConverterStorage{
	handleMap: newConcurrentHandleMap[Storage](),
}

func (c FfiConverterStorage) Lift(pointer unsafe.Pointer) Storage {
	result := &StorageImpl{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_storage(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_storage(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*StorageImpl).Destroy)
	return result
}

func (c FfiConverterStorage) Read(reader io.Reader) Storage {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterStorage) Lower(value Storage) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := unsafe.Pointer(uintptr(c.handleMap.insert(value)))
	return pointer

}

func (c FfiConverterStorage) Write(writer io.Writer, value Storage) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerStorage struct{}

func (_ FfiDestroyerStorage) Destroy(value Storage) {
	if val, ok := value.(*StorageImpl); ok {
		val.Destroy()
	} else {
		panic("Expected *StorageImpl")
	}
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod0
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod0(uniffiHandle C.uint64_t, key C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.DeleteCachedItem(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: key,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod1
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod1(uniffiHandle C.uint64_t, key C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetCachedItem(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: key,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterOptionalStringINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod2
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod2(uniffiHandle C.uint64_t, key C.RustBuffer, value C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.SetCachedItem(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: key,
				}),
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: value,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod3
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod3(uniffiHandle C.uint64_t, request C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.ListPayments(
				FfiConverterListPaymentsRequestINSTANCE.Lift(GoRustBuffer{
					inner: request,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterSequencePaymentINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod4
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod4(uniffiHandle C.uint64_t, payment C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.InsertPayment(
				FfiConverterPaymentINSTANCE.Lift(GoRustBuffer{
					inner: payment,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod5
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod5(uniffiHandle C.uint64_t, paymentId C.RustBuffer, metadata C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.SetPaymentMetadata(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: paymentId,
				}),
				FfiConverterPaymentMetadataINSTANCE.Lift(GoRustBuffer{
					inner: metadata,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod6
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod6(uniffiHandle C.uint64_t, id C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetPaymentById(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: id,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterPaymentINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod7
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod7(uniffiHandle C.uint64_t, invoice C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetPaymentByInvoice(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: invoice,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterOptionalPaymentINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod8
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod8(uniffiHandle C.uint64_t, txid C.RustBuffer, vout C.uint32_t, amountSats C.uint64_t, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.AddDeposit(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: txid,
				}),
				FfiConverterUint32INSTANCE.Lift(vout),
				FfiConverterUint64INSTANCE.Lift(amountSats),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod9
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod9(uniffiHandle C.uint64_t, txid C.RustBuffer, vout C.uint32_t, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.DeleteDeposit(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: txid,
				}),
				FfiConverterUint32INSTANCE.Lift(vout),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod10
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod10(uniffiHandle C.uint64_t, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.ListDeposits()

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterSequenceDepositInfoINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod11
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod11(uniffiHandle C.uint64_t, txid C.RustBuffer, vout C.uint32_t, payload C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.UpdateDeposit(
				FfiConverterStringINSTANCE.Lift(GoRustBuffer{
					inner: txid,
				}),
				FfiConverterUint32INSTANCE.Lift(vout),
				FfiConverterUpdateDepositPayloadINSTANCE.Lift(GoRustBuffer{
					inner: payload,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod12
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod12(uniffiHandle C.uint64_t, metadata C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.SetLnurlMetadata(
				FfiConverterSequenceSetLnurlMetadataItemINSTANCE.Lift(GoRustBuffer{
					inner: metadata,
				}),
			)

		if err != nil {
			var actualError *StorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

var UniffiVTableCallbackInterfaceStorageINSTANCE = C.UniffiVTableCallbackInterfaceStorage{
	deleteCachedItem:    (C.UniffiCallbackInterfaceStorageMethod0)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod0),
	getCachedItem:       (C.UniffiCallbackInterfaceStorageMethod1)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod1),
	setCachedItem:       (C.UniffiCallbackInterfaceStorageMethod2)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod2),
	listPayments:        (C.UniffiCallbackInterfaceStorageMethod3)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod3),
	insertPayment:       (C.UniffiCallbackInterfaceStorageMethod4)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod4),
	setPaymentMetadata:  (C.UniffiCallbackInterfaceStorageMethod5)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod5),
	getPaymentById:      (C.UniffiCallbackInterfaceStorageMethod6)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod6),
	getPaymentByInvoice: (C.UniffiCallbackInterfaceStorageMethod7)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod7),
	addDeposit:          (C.UniffiCallbackInterfaceStorageMethod8)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod8),
	deleteDeposit:       (C.UniffiCallbackInterfaceStorageMethod9)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod9),
	listDeposits:        (C.UniffiCallbackInterfaceStorageMethod10)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod10),
	updateDeposit:       (C.UniffiCallbackInterfaceStorageMethod11)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod11),
	setLnurlMetadata:    (C.UniffiCallbackInterfaceStorageMethod12)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageMethod12),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageFree),
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageFree
func breez_sdk_spark_cgo_dispatchCallbackInterfaceStorageFree(handle C.uint64_t) {
	FfiConverterStorageINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterStorage) register() {
	C.uniffi_breez_sdk_spark_fn_init_callback_vtable_storage(&UniffiVTableCallbackInterfaceStorageINSTANCE)
}

type SyncStorage interface {
	AddOutgoingChange(record UnversionedRecordChange) (uint64, error)
	CompleteOutgoingSync(record Record) error
	GetPendingOutgoingChanges(limit uint32) ([]OutgoingChange, error)
	// Get the revision number of the last synchronized record
	GetLastRevision() (uint64, error)
	// Insert incoming records from remote sync
	InsertIncomingRecords(records []Record) error
	// Delete an incoming record after it has been processed
	DeleteIncomingRecord(record Record) error
	// Update revision numbers of pending outgoing records to be higher than the given revision
	RebasePendingOutgoingRecords(revision uint64) error
	// Get incoming records that need to be processed, up to the specified limit
	GetIncomingRecords(limit uint32) ([]IncomingChange, error)
	// Get the latest outgoing record if any exists
	GetLatestOutgoingChange() (*OutgoingChange, error)
	// Update the sync state record from an incoming record
	UpdateRecordFromIncoming(record Record) error
}
type SyncStorageImpl struct {
	ffiObject FfiObject
}

func (_self *SyncStorageImpl) AddOutgoingChange(record UnversionedRecordChange) (uint64, error) {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) C.uint64_t {
			res := C.ffi_breez_sdk_spark_rust_future_complete_u64(handle, status)
			return res
		},
		// liftFn
		func(ffi C.uint64_t) uint64 {
			return FfiConverterUint64INSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_add_outgoing_change(
			_pointer, FfiConverterUnversionedRecordChangeINSTANCE.Lower(record)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_u64(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_u64(handle)
		},
	)

	return res, err
}

func (_self *SyncStorageImpl) CompleteOutgoingSync(record Record) error {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_complete_outgoing_sync(
			_pointer, FfiConverterRecordINSTANCE.Lower(record)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

func (_self *SyncStorageImpl) GetPendingOutgoingChanges(limit uint32) ([]OutgoingChange, error) {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) []OutgoingChange {
			return FfiConverterSequenceOutgoingChangeINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_get_pending_outgoing_changes(
			_pointer, FfiConverterUint32INSTANCE.Lower(limit)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Get the revision number of the last synchronized record
func (_self *SyncStorageImpl) GetLastRevision() (uint64, error) {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) C.uint64_t {
			res := C.ffi_breez_sdk_spark_rust_future_complete_u64(handle, status)
			return res
		},
		// liftFn
		func(ffi C.uint64_t) uint64 {
			return FfiConverterUint64INSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_get_last_revision(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_u64(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_u64(handle)
		},
	)

	return res, err
}

// Insert incoming records from remote sync
func (_self *SyncStorageImpl) InsertIncomingRecords(records []Record) error {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_insert_incoming_records(
			_pointer, FfiConverterSequenceRecordINSTANCE.Lower(records)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Delete an incoming record after it has been processed
func (_self *SyncStorageImpl) DeleteIncomingRecord(record Record) error {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_delete_incoming_record(
			_pointer, FfiConverterRecordINSTANCE.Lower(record)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Update revision numbers of pending outgoing records to be higher than the given revision
func (_self *SyncStorageImpl) RebasePendingOutgoingRecords(revision uint64) error {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_rebase_pending_outgoing_records(
			_pointer, FfiConverterUint64INSTANCE.Lower(revision)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}

// Get incoming records that need to be processed, up to the specified limit
func (_self *SyncStorageImpl) GetIncomingRecords(limit uint32) ([]IncomingChange, error) {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) []IncomingChange {
			return FfiConverterSequenceIncomingChangeINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_get_incoming_records(
			_pointer, FfiConverterUint32INSTANCE.Lower(limit)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Get the latest outgoing record if any exists
func (_self *SyncStorageImpl) GetLatestOutgoingChange() (*OutgoingChange, error) {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) *OutgoingChange {
			return FfiConverterOptionalOutgoingChangeINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_get_latest_outgoing_change(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Update the sync state record from an incoming record
func (_self *SyncStorageImpl) UpdateRecordFromIncoming(record Record) error {
	_pointer := _self.ffiObject.incrementPointer("SyncStorage")
	defer _self.ffiObject.decrementPointer()
	_, err := uniffiRustCallAsync[SyncStorageError](
		FfiConverterSyncStorageErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) struct{} {
			C.ffi_breez_sdk_spark_rust_future_complete_void(handle, status)
			return struct{}{}
		},
		// liftFn
		func(_ struct{}) struct{} { return struct{}{} },
		C.uniffi_breez_sdk_spark_fn_method_syncstorage_update_record_from_incoming(
			_pointer, FfiConverterRecordINSTANCE.Lower(record)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_void(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_void(handle)
		},
	)

	return err
}
func (object *SyncStorageImpl) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterSyncStorage struct {
	handleMap *concurrentHandleMap[SyncStorage]
}

var FfiConverterSyncStorageINSTANCE = FfiConverterSyncStorage{
	handleMap: newConcurrentHandleMap[SyncStorage](),
}

func (c FfiConverterSyncStorage) Lift(pointer unsafe.Pointer) SyncStorage {
	result := &SyncStorageImpl{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_syncstorage(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_syncstorage(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*SyncStorageImpl).Destroy)
	return result
}

func (c FfiConverterSyncStorage) Read(reader io.Reader) SyncStorage {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterSyncStorage) Lower(value SyncStorage) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := unsafe.Pointer(uintptr(c.handleMap.insert(value)))
	return pointer

}

func (c FfiConverterSyncStorage) Write(writer io.Writer, value SyncStorage) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerSyncStorage struct{}

func (_ FfiDestroyerSyncStorage) Destroy(value SyncStorage) {
	if val, ok := value.(*SyncStorageImpl); ok {
		val.Destroy()
	} else {
		panic("Expected *SyncStorageImpl")
	}
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod0
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod0(uniffiHandle C.uint64_t, record C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteU64, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructU64, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteU64(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructU64{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.AddOutgoingChange(
				FfiConverterUnversionedRecordChangeINSTANCE.Lift(GoRustBuffer{
					inner: record,
				}),
			)

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterUint64INSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod1
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod1(uniffiHandle C.uint64_t, record C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.CompleteOutgoingSync(
				FfiConverterRecordINSTANCE.Lift(GoRustBuffer{
					inner: record,
				}),
			)

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod2
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod2(uniffiHandle C.uint64_t, limit C.uint32_t, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetPendingOutgoingChanges(
				FfiConverterUint32INSTANCE.Lift(limit),
			)

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterSequenceOutgoingChangeINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod3
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod3(uniffiHandle C.uint64_t, uniffiFutureCallback C.UniffiForeignFutureCompleteU64, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructU64, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteU64(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructU64{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetLastRevision()

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterUint64INSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod4
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod4(uniffiHandle C.uint64_t, records C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.InsertIncomingRecords(
				FfiConverterSequenceRecordINSTANCE.Lift(GoRustBuffer{
					inner: records,
				}),
			)

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod5
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod5(uniffiHandle C.uint64_t, record C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.DeleteIncomingRecord(
				FfiConverterRecordINSTANCE.Lift(GoRustBuffer{
					inner: record,
				}),
			)

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod6
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod6(uniffiHandle C.uint64_t, revision C.uint64_t, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.RebasePendingOutgoingRecords(
				FfiConverterUint64INSTANCE.Lift(revision),
			)

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod7
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod7(uniffiHandle C.uint64_t, limit C.uint32_t, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetIncomingRecords(
				FfiConverterUint32INSTANCE.Lift(limit),
			)

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterSequenceIncomingChangeINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod8
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod8(uniffiHandle C.uint64_t, uniffiFutureCallback C.UniffiForeignFutureCompleteRustBuffer, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructRustBuffer, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteRustBuffer(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructRustBuffer{}
		uniffiOutReturn := &asyncResult.returnValue
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		res, err :=
			uniffiObj.GetLatestOutgoingChange()

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

		*uniffiOutReturn = FfiConverterOptionalOutgoingChangeINSTANCE.Lower(res)
	}()
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod9
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod9(uniffiHandle C.uint64_t, record C.RustBuffer, uniffiFutureCallback C.UniffiForeignFutureCompleteVoid, uniffiCallbackData C.uint64_t, uniffiOutReturn *C.UniffiForeignFuture) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterSyncStorageINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}

	result := make(chan C.UniffiForeignFutureStructVoid, 1)
	cancel := make(chan struct{}, 1)
	guardHandle := cgo.NewHandle(cancel)
	*uniffiOutReturn = C.UniffiForeignFuture{
		handle: C.uint64_t(guardHandle),
		free:   C.UniffiForeignFutureFree(C.breez_sdk_spark_uniffiFreeGorutine),
	}

	// Wait for compleation or cancel
	go func() {
		select {
		case <-cancel:
		case res := <-result:
			C.call_UniffiForeignFutureCompleteVoid(uniffiFutureCallback, uniffiCallbackData, res)
		}
	}()

	// Eval callback asynchroniously
	go func() {
		asyncResult := &C.UniffiForeignFutureStructVoid{}
		callStatus := &asyncResult.callStatus
		defer func() {
			result <- *asyncResult
		}()

		err :=
			uniffiObj.UpdateRecordFromIncoming(
				FfiConverterRecordINSTANCE.Lift(GoRustBuffer{
					inner: record,
				}),
			)

		if err != nil {
			var actualError *SyncStorageError
			if errors.As(err, &actualError) {
				if actualError != nil {
					*callStatus = C.RustCallStatus{
						code:     C.int8_t(uniffiCallbackResultError),
						errorBuf: FfiConverterSyncStorageErrorINSTANCE.Lower(actualError),
					}
					return
				}
			} else {
				*callStatus = C.RustCallStatus{
					code: C.int8_t(uniffiCallbackUnexpectedResultError),
				}
				return
			}
		}

	}()
}

var UniffiVTableCallbackInterfaceSyncStorageINSTANCE = C.UniffiVTableCallbackInterfaceSyncStorage{
	addOutgoingChange:            (C.UniffiCallbackInterfaceSyncStorageMethod0)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod0),
	completeOutgoingSync:         (C.UniffiCallbackInterfaceSyncStorageMethod1)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod1),
	getPendingOutgoingChanges:    (C.UniffiCallbackInterfaceSyncStorageMethod2)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod2),
	getLastRevision:              (C.UniffiCallbackInterfaceSyncStorageMethod3)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod3),
	insertIncomingRecords:        (C.UniffiCallbackInterfaceSyncStorageMethod4)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod4),
	deleteIncomingRecord:         (C.UniffiCallbackInterfaceSyncStorageMethod5)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod5),
	rebasePendingOutgoingRecords: (C.UniffiCallbackInterfaceSyncStorageMethod6)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod6),
	getIncomingRecords:           (C.UniffiCallbackInterfaceSyncStorageMethod7)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod7),
	getLatestOutgoingChange:      (C.UniffiCallbackInterfaceSyncStorageMethod8)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod8),
	updateRecordFromIncoming:     (C.UniffiCallbackInterfaceSyncStorageMethod9)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageMethod9),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageFree),
}

//export breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageFree
func breez_sdk_spark_cgo_dispatchCallbackInterfaceSyncStorageFree(handle C.uint64_t) {
	FfiConverterSyncStorageINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterSyncStorage) register() {
	C.uniffi_breez_sdk_spark_fn_init_callback_vtable_syncstorage(&UniffiVTableCallbackInterfaceSyncStorageINSTANCE)
}

type TokenIssuerInterface interface {
	// Burns supply of the issuer token
	//
	// # Arguments
	//
	// * `request`: The request containing the amount of the supply to burn
	//
	// # Returns
	//
	// Result containing either:
	// * `Payment` - The payment representing the burn transaction
	// * `SdkError` - If there was an error during the burn process
	BurnIssuerToken(request BurnIssuerTokenRequest) (Payment, error)
	// Creates a new issuer token
	//
	// # Arguments
	//
	// * `request`: The request containing the token parameters
	//
	// # Returns
	//
	// Result containing either:
	// * `TokenMetadata` - The metadata of the created token
	// * `SdkError` - If there was an error during the token creation
	CreateIssuerToken(request CreateIssuerTokenRequest) (TokenMetadata, error)
	// Freezes tokens held at the specified address
	//
	// # Arguments
	//
	// * `request`: The request containing the spark address where the tokens to be frozen are held
	//
	// # Returns
	//
	// Result containing either:
	// * `FreezeIssuerTokenResponse` - The response containing details of the freeze operation
	// * `SdkError` - If there was an error during the freeze process
	FreezeIssuerToken(request FreezeIssuerTokenRequest) (FreezeIssuerTokenResponse, error)
	// Gets the issuer token balance
	//
	// # Returns
	//
	// Result containing either:
	// * `TokenBalance` - The balance of the issuer token
	// * `SdkError` - If there was an error during the retrieval or no issuer token exists
	GetIssuerTokenBalance() (TokenBalance, error)
	// Gets the issuer token metadata
	//
	// # Returns
	//
	// Result containing either:
	// * `TokenMetadata` - The metadata of the issuer token
	// * `SdkError` - If there was an error during the retrieval or no issuer token exists
	GetIssuerTokenMetadata() (TokenMetadata, error)
	// Mints supply for the issuer token
	//
	// # Arguments
	//
	// * `request`: The request contiaining the amount of the supply to mint
	//
	// # Returns
	//
	// Result containing either:
	// * `Payment` - The payment representing the minting transaction
	// * `SdkError` - If there was an error during the minting process
	MintIssuerToken(request MintIssuerTokenRequest) (Payment, error)
	// Unfreezes tokens held at the specified address
	//
	// # Arguments
	//
	// * `request`: The request containing the spark address where the tokens to be unfrozen are held
	//
	// # Returns
	//
	// Result containing either:
	// * `UnfreezeIssuerTokenResponse` - The response containing details of the unfreeze operation
	// * `SdkError` - If there was an error during the unfreeze process
	UnfreezeIssuerToken(request UnfreezeIssuerTokenRequest) (UnfreezeIssuerTokenResponse, error)
}
type TokenIssuer struct {
	ffiObject FfiObject
}

// Burns supply of the issuer token
//
// # Arguments
//
// * `request`: The request containing the amount of the supply to burn
//
// # Returns
//
// Result containing either:
// * `Payment` - The payment representing the burn transaction
// * `SdkError` - If there was an error during the burn process
func (_self *TokenIssuer) BurnIssuerToken(request BurnIssuerTokenRequest) (Payment, error) {
	_pointer := _self.ffiObject.incrementPointer("*TokenIssuer")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) Payment {
			return FfiConverterPaymentINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_tokenissuer_burn_issuer_token(
			_pointer, FfiConverterBurnIssuerTokenRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Creates a new issuer token
//
// # Arguments
//
// * `request`: The request containing the token parameters
//
// # Returns
//
// Result containing either:
// * `TokenMetadata` - The metadata of the created token
// * `SdkError` - If there was an error during the token creation
func (_self *TokenIssuer) CreateIssuerToken(request CreateIssuerTokenRequest) (TokenMetadata, error) {
	_pointer := _self.ffiObject.incrementPointer("*TokenIssuer")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) TokenMetadata {
			return FfiConverterTokenMetadataINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_tokenissuer_create_issuer_token(
			_pointer, FfiConverterCreateIssuerTokenRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Freezes tokens held at the specified address
//
// # Arguments
//
// * `request`: The request containing the spark address where the tokens to be frozen are held
//
// # Returns
//
// Result containing either:
// * `FreezeIssuerTokenResponse` - The response containing details of the freeze operation
// * `SdkError` - If there was an error during the freeze process
func (_self *TokenIssuer) FreezeIssuerToken(request FreezeIssuerTokenRequest) (FreezeIssuerTokenResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*TokenIssuer")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) FreezeIssuerTokenResponse {
			return FfiConverterFreezeIssuerTokenResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_tokenissuer_freeze_issuer_token(
			_pointer, FfiConverterFreezeIssuerTokenRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Gets the issuer token balance
//
// # Returns
//
// Result containing either:
// * `TokenBalance` - The balance of the issuer token
// * `SdkError` - If there was an error during the retrieval or no issuer token exists
func (_self *TokenIssuer) GetIssuerTokenBalance() (TokenBalance, error) {
	_pointer := _self.ffiObject.incrementPointer("*TokenIssuer")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) TokenBalance {
			return FfiConverterTokenBalanceINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_tokenissuer_get_issuer_token_balance(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Gets the issuer token metadata
//
// # Returns
//
// Result containing either:
// * `TokenMetadata` - The metadata of the issuer token
// * `SdkError` - If there was an error during the retrieval or no issuer token exists
func (_self *TokenIssuer) GetIssuerTokenMetadata() (TokenMetadata, error) {
	_pointer := _self.ffiObject.incrementPointer("*TokenIssuer")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) TokenMetadata {
			return FfiConverterTokenMetadataINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_tokenissuer_get_issuer_token_metadata(
			_pointer),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Mints supply for the issuer token
//
// # Arguments
//
// * `request`: The request contiaining the amount of the supply to mint
//
// # Returns
//
// Result containing either:
// * `Payment` - The payment representing the minting transaction
// * `SdkError` - If there was an error during the minting process
func (_self *TokenIssuer) MintIssuerToken(request MintIssuerTokenRequest) (Payment, error) {
	_pointer := _self.ffiObject.incrementPointer("*TokenIssuer")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) Payment {
			return FfiConverterPaymentINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_tokenissuer_mint_issuer_token(
			_pointer, FfiConverterMintIssuerTokenRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}

// Unfreezes tokens held at the specified address
//
// # Arguments
//
// * `request`: The request containing the spark address where the tokens to be unfrozen are held
//
// # Returns
//
// Result containing either:
// * `UnfreezeIssuerTokenResponse` - The response containing details of the unfreeze operation
// * `SdkError` - If there was an error during the unfreeze process
func (_self *TokenIssuer) UnfreezeIssuerToken(request UnfreezeIssuerTokenRequest) (UnfreezeIssuerTokenResponse, error) {
	_pointer := _self.ffiObject.incrementPointer("*TokenIssuer")
	defer _self.ffiObject.decrementPointer()
	res, err := uniffiRustCallAsync[SdkError](
		FfiConverterSdkErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_breez_sdk_spark_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) UnfreezeIssuerTokenResponse {
			return FfiConverterUnfreezeIssuerTokenResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_breez_sdk_spark_fn_method_tokenissuer_unfreeze_issuer_token(
			_pointer, FfiConverterUnfreezeIssuerTokenRequestINSTANCE.Lower(request)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_breez_sdk_spark_rust_future_free_rust_buffer(handle)
		},
	)

	return res, err
}
func (object *TokenIssuer) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTokenIssuer struct{}

var FfiConverterTokenIssuerINSTANCE = FfiConverterTokenIssuer{}

func (c FfiConverterTokenIssuer) Lift(pointer unsafe.Pointer) *TokenIssuer {
	result := &TokenIssuer{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_breez_sdk_spark_fn_clone_tokenissuer(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_breez_sdk_spark_fn_free_tokenissuer(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*TokenIssuer).Destroy)
	return result
}

func (c FfiConverterTokenIssuer) Read(reader io.Reader) *TokenIssuer {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterTokenIssuer) Lower(value *TokenIssuer) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*TokenIssuer")
	defer value.ffiObject.decrementPointer()
	return pointer

}

func (c FfiConverterTokenIssuer) Write(writer io.Writer, value *TokenIssuer) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerTokenIssuer struct{}

func (_ FfiDestroyerTokenIssuer) Destroy(value *TokenIssuer) {
	value.Destroy()
}

// Payload of the AES success action, as received from the LNURL endpoint
//
// See [`AesSuccessActionDataDecrypted`] for a similar wrapper containing the decrypted payload
type AesSuccessActionData struct {
	// Contents description, up to 144 characters
	Description string
	// Base64, AES-encrypted data where encryption key is payment preimage, up to 4kb of characters
	Ciphertext string
	// Base64, initialization vector, exactly 24 characters
	Iv string
}

func (r *AesSuccessActionData) Destroy() {
	FfiDestroyerString{}.Destroy(r.Description)
	FfiDestroyerString{}.Destroy(r.Ciphertext)
	FfiDestroyerString{}.Destroy(r.Iv)
}

type FfiConverterAesSuccessActionData struct{}

var FfiConverterAesSuccessActionDataINSTANCE = FfiConverterAesSuccessActionData{}

func (c FfiConverterAesSuccessActionData) Lift(rb RustBufferI) AesSuccessActionData {
	return LiftFromRustBuffer[AesSuccessActionData](c, rb)
}

func (c FfiConverterAesSuccessActionData) Read(reader io.Reader) AesSuccessActionData {
	return AesSuccessActionData{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterAesSuccessActionData) Lower(value AesSuccessActionData) C.RustBuffer {
	return LowerIntoRustBuffer[AesSuccessActionData](c, value)
}

func (c FfiConverterAesSuccessActionData) Write(writer io.Writer, value AesSuccessActionData) {
	FfiConverterStringINSTANCE.Write(writer, value.Description)
	FfiConverterStringINSTANCE.Write(writer, value.Ciphertext)
	FfiConverterStringINSTANCE.Write(writer, value.Iv)
}

type FfiDestroyerAesSuccessActionData struct{}

func (_ FfiDestroyerAesSuccessActionData) Destroy(value AesSuccessActionData) {
	value.Destroy()
}

// Wrapper for the decrypted [`AesSuccessActionData`] payload
type AesSuccessActionDataDecrypted struct {
	// Contents description, up to 144 characters
	Description string
	// Decrypted content
	Plaintext string
}

func (r *AesSuccessActionDataDecrypted) Destroy() {
	FfiDestroyerString{}.Destroy(r.Description)
	FfiDestroyerString{}.Destroy(r.Plaintext)
}

type FfiConverterAesSuccessActionDataDecrypted struct{}

var FfiConverterAesSuccessActionDataDecryptedINSTANCE = FfiConverterAesSuccessActionDataDecrypted{}

func (c FfiConverterAesSuccessActionDataDecrypted) Lift(rb RustBufferI) AesSuccessActionDataDecrypted {
	return LiftFromRustBuffer[AesSuccessActionDataDecrypted](c, rb)
}

func (c FfiConverterAesSuccessActionDataDecrypted) Read(reader io.Reader) AesSuccessActionDataDecrypted {
	return AesSuccessActionDataDecrypted{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterAesSuccessActionDataDecrypted) Lower(value AesSuccessActionDataDecrypted) C.RustBuffer {
	return LowerIntoRustBuffer[AesSuccessActionDataDecrypted](c, value)
}

func (c FfiConverterAesSuccessActionDataDecrypted) Write(writer io.Writer, value AesSuccessActionDataDecrypted) {
	FfiConverterStringINSTANCE.Write(writer, value.Description)
	FfiConverterStringINSTANCE.Write(writer, value.Plaintext)
}

type FfiDestroyerAesSuccessActionDataDecrypted struct{}

func (_ FfiDestroyerAesSuccessActionDataDecrypted) Destroy(value AesSuccessActionDataDecrypted) {
	value.Destroy()
}

type Bip21Details struct {
	AmountSat      *uint64
	AssetId        *string
	Uri            string
	Extras         []Bip21Extra
	Label          *string
	Message        *string
	PaymentMethods []InputType
}

func (r *Bip21Details) Destroy() {
	FfiDestroyerOptionalUint64{}.Destroy(r.AmountSat)
	FfiDestroyerOptionalString{}.Destroy(r.AssetId)
	FfiDestroyerString{}.Destroy(r.Uri)
	FfiDestroyerSequenceBip21Extra{}.Destroy(r.Extras)
	FfiDestroyerOptionalString{}.Destroy(r.Label)
	FfiDestroyerOptionalString{}.Destroy(r.Message)
	FfiDestroyerSequenceInputType{}.Destroy(r.PaymentMethods)
}

type FfiConverterBip21Details struct{}

var FfiConverterBip21DetailsINSTANCE = FfiConverterBip21Details{}

func (c FfiConverterBip21Details) Lift(rb RustBufferI) Bip21Details {
	return LiftFromRustBuffer[Bip21Details](c, rb)
}

func (c FfiConverterBip21Details) Read(reader io.Reader) Bip21Details {
	return Bip21Details{
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterSequenceBip21ExtraINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterSequenceInputTypeINSTANCE.Read(reader),
	}
}

func (c FfiConverterBip21Details) Lower(value Bip21Details) C.RustBuffer {
	return LowerIntoRustBuffer[Bip21Details](c, value)
}

func (c FfiConverterBip21Details) Write(writer io.Writer, value Bip21Details) {
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.AmountSat)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.AssetId)
	FfiConverterStringINSTANCE.Write(writer, value.Uri)
	FfiConverterSequenceBip21ExtraINSTANCE.Write(writer, value.Extras)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Label)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Message)
	FfiConverterSequenceInputTypeINSTANCE.Write(writer, value.PaymentMethods)
}

type FfiDestroyerBip21Details struct{}

func (_ FfiDestroyerBip21Details) Destroy(value Bip21Details) {
	value.Destroy()
}

type Bip21Extra struct {
	Key   string
	Value string
}

func (r *Bip21Extra) Destroy() {
	FfiDestroyerString{}.Destroy(r.Key)
	FfiDestroyerString{}.Destroy(r.Value)
}

type FfiConverterBip21Extra struct{}

var FfiConverterBip21ExtraINSTANCE = FfiConverterBip21Extra{}

func (c FfiConverterBip21Extra) Lift(rb RustBufferI) Bip21Extra {
	return LiftFromRustBuffer[Bip21Extra](c, rb)
}

func (c FfiConverterBip21Extra) Read(reader io.Reader) Bip21Extra {
	return Bip21Extra{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterBip21Extra) Lower(value Bip21Extra) C.RustBuffer {
	return LowerIntoRustBuffer[Bip21Extra](c, value)
}

func (c FfiConverterBip21Extra) Write(writer io.Writer, value Bip21Extra) {
	FfiConverterStringINSTANCE.Write(writer, value.Key)
	FfiConverterStringINSTANCE.Write(writer, value.Value)
}

type FfiDestroyerBip21Extra struct{}

func (_ FfiDestroyerBip21Extra) Destroy(value Bip21Extra) {
	value.Destroy()
}

type BitcoinAddressDetails struct {
	Address string
	Network BitcoinNetwork
	Source  PaymentRequestSource
}

func (r *BitcoinAddressDetails) Destroy() {
	FfiDestroyerString{}.Destroy(r.Address)
	FfiDestroyerBitcoinNetwork{}.Destroy(r.Network)
	FfiDestroyerPaymentRequestSource{}.Destroy(r.Source)
}

type FfiConverterBitcoinAddressDetails struct{}

var FfiConverterBitcoinAddressDetailsINSTANCE = FfiConverterBitcoinAddressDetails{}

func (c FfiConverterBitcoinAddressDetails) Lift(rb RustBufferI) BitcoinAddressDetails {
	return LiftFromRustBuffer[BitcoinAddressDetails](c, rb)
}

func (c FfiConverterBitcoinAddressDetails) Read(reader io.Reader) BitcoinAddressDetails {
	return BitcoinAddressDetails{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterBitcoinNetworkINSTANCE.Read(reader),
		FfiConverterPaymentRequestSourceINSTANCE.Read(reader),
	}
}

func (c FfiConverterBitcoinAddressDetails) Lower(value BitcoinAddressDetails) C.RustBuffer {
	return LowerIntoRustBuffer[BitcoinAddressDetails](c, value)
}

func (c FfiConverterBitcoinAddressDetails) Write(writer io.Writer, value BitcoinAddressDetails) {
	FfiConverterStringINSTANCE.Write(writer, value.Address)
	FfiConverterBitcoinNetworkINSTANCE.Write(writer, value.Network)
	FfiConverterPaymentRequestSourceINSTANCE.Write(writer, value.Source)
}

type FfiDestroyerBitcoinAddressDetails struct{}

func (_ FfiDestroyerBitcoinAddressDetails) Destroy(value BitcoinAddressDetails) {
	value.Destroy()
}

type Bolt11Invoice struct {
	Bolt11 string
	Source PaymentRequestSource
}

func (r *Bolt11Invoice) Destroy() {
	FfiDestroyerString{}.Destroy(r.Bolt11)
	FfiDestroyerPaymentRequestSource{}.Destroy(r.Source)
}

type FfiConverterBolt11Invoice struct{}

var FfiConverterBolt11InvoiceINSTANCE = FfiConverterBolt11Invoice{}

func (c FfiConverterBolt11Invoice) Lift(rb RustBufferI) Bolt11Invoice {
	return LiftFromRustBuffer[Bolt11Invoice](c, rb)
}

func (c FfiConverterBolt11Invoice) Read(reader io.Reader) Bolt11Invoice {
	return Bolt11Invoice{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterPaymentRequestSourceINSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt11Invoice) Lower(value Bolt11Invoice) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt11Invoice](c, value)
}

func (c FfiConverterBolt11Invoice) Write(writer io.Writer, value Bolt11Invoice) {
	FfiConverterStringINSTANCE.Write(writer, value.Bolt11)
	FfiConverterPaymentRequestSourceINSTANCE.Write(writer, value.Source)
}

type FfiDestroyerBolt11Invoice struct{}

func (_ FfiDestroyerBolt11Invoice) Destroy(value Bolt11Invoice) {
	value.Destroy()
}

type Bolt11InvoiceDetails struct {
	AmountMsat              *uint64
	Description             *string
	DescriptionHash         *string
	Expiry                  uint64
	Invoice                 Bolt11Invoice
	MinFinalCltvExpiryDelta uint64
	Network                 BitcoinNetwork
	PayeePubkey             string
	PaymentHash             string
	PaymentSecret           string
	RoutingHints            []Bolt11RouteHint
	Timestamp               uint64
}

func (r *Bolt11InvoiceDetails) Destroy() {
	FfiDestroyerOptionalUint64{}.Destroy(r.AmountMsat)
	FfiDestroyerOptionalString{}.Destroy(r.Description)
	FfiDestroyerOptionalString{}.Destroy(r.DescriptionHash)
	FfiDestroyerUint64{}.Destroy(r.Expiry)
	FfiDestroyerBolt11Invoice{}.Destroy(r.Invoice)
	FfiDestroyerUint64{}.Destroy(r.MinFinalCltvExpiryDelta)
	FfiDestroyerBitcoinNetwork{}.Destroy(r.Network)
	FfiDestroyerString{}.Destroy(r.PayeePubkey)
	FfiDestroyerString{}.Destroy(r.PaymentHash)
	FfiDestroyerString{}.Destroy(r.PaymentSecret)
	FfiDestroyerSequenceBolt11RouteHint{}.Destroy(r.RoutingHints)
	FfiDestroyerUint64{}.Destroy(r.Timestamp)
}

type FfiConverterBolt11InvoiceDetails struct{}

var FfiConverterBolt11InvoiceDetailsINSTANCE = FfiConverterBolt11InvoiceDetails{}

func (c FfiConverterBolt11InvoiceDetails) Lift(rb RustBufferI) Bolt11InvoiceDetails {
	return LiftFromRustBuffer[Bolt11InvoiceDetails](c, rb)
}

func (c FfiConverterBolt11InvoiceDetails) Read(reader io.Reader) Bolt11InvoiceDetails {
	return Bolt11InvoiceDetails{
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterBolt11InvoiceINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterBitcoinNetworkINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterSequenceBolt11RouteHintINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt11InvoiceDetails) Lower(value Bolt11InvoiceDetails) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt11InvoiceDetails](c, value)
}

func (c FfiConverterBolt11InvoiceDetails) Write(writer io.Writer, value Bolt11InvoiceDetails) {
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.AmountMsat)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Description)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.DescriptionHash)
	FfiConverterUint64INSTANCE.Write(writer, value.Expiry)
	FfiConverterBolt11InvoiceINSTANCE.Write(writer, value.Invoice)
	FfiConverterUint64INSTANCE.Write(writer, value.MinFinalCltvExpiryDelta)
	FfiConverterBitcoinNetworkINSTANCE.Write(writer, value.Network)
	FfiConverterStringINSTANCE.Write(writer, value.PayeePubkey)
	FfiConverterStringINSTANCE.Write(writer, value.PaymentHash)
	FfiConverterStringINSTANCE.Write(writer, value.PaymentSecret)
	FfiConverterSequenceBolt11RouteHintINSTANCE.Write(writer, value.RoutingHints)
	FfiConverterUint64INSTANCE.Write(writer, value.Timestamp)
}

type FfiDestroyerBolt11InvoiceDetails struct{}

func (_ FfiDestroyerBolt11InvoiceDetails) Destroy(value Bolt11InvoiceDetails) {
	value.Destroy()
}

type Bolt11RouteHint struct {
	Hops []Bolt11RouteHintHop
}

func (r *Bolt11RouteHint) Destroy() {
	FfiDestroyerSequenceBolt11RouteHintHop{}.Destroy(r.Hops)
}

type FfiConverterBolt11RouteHint struct{}

var FfiConverterBolt11RouteHintINSTANCE = FfiConverterBolt11RouteHint{}

func (c FfiConverterBolt11RouteHint) Lift(rb RustBufferI) Bolt11RouteHint {
	return LiftFromRustBuffer[Bolt11RouteHint](c, rb)
}

func (c FfiConverterBolt11RouteHint) Read(reader io.Reader) Bolt11RouteHint {
	return Bolt11RouteHint{
		FfiConverterSequenceBolt11RouteHintHopINSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt11RouteHint) Lower(value Bolt11RouteHint) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt11RouteHint](c, value)
}

func (c FfiConverterBolt11RouteHint) Write(writer io.Writer, value Bolt11RouteHint) {
	FfiConverterSequenceBolt11RouteHintHopINSTANCE.Write(writer, value.Hops)
}

type FfiDestroyerBolt11RouteHint struct{}

func (_ FfiDestroyerBolt11RouteHint) Destroy(value Bolt11RouteHint) {
	value.Destroy()
}

type Bolt11RouteHintHop struct {
	// The `node_id` of the non-target end of the route
	SrcNodeId string
	// The `short_channel_id` of this channel
	ShortChannelId string
	// The fees which must be paid to use this channel
	FeesBaseMsat               uint32
	FeesProportionalMillionths uint32
	// The difference in CLTV values between this node and the next node.
	CltvExpiryDelta uint16
	// The minimum value, in msat, which must be relayed to the next hop.
	HtlcMinimumMsat *uint64
	// The maximum value in msat available for routing with a single HTLC.
	HtlcMaximumMsat *uint64
}

func (r *Bolt11RouteHintHop) Destroy() {
	FfiDestroyerString{}.Destroy(r.SrcNodeId)
	FfiDestroyerString{}.Destroy(r.ShortChannelId)
	FfiDestroyerUint32{}.Destroy(r.FeesBaseMsat)
	FfiDestroyerUint32{}.Destroy(r.FeesProportionalMillionths)
	FfiDestroyerUint16{}.Destroy(r.CltvExpiryDelta)
	FfiDestroyerOptionalUint64{}.Destroy(r.HtlcMinimumMsat)
	FfiDestroyerOptionalUint64{}.Destroy(r.HtlcMaximumMsat)
}

type FfiConverterBolt11RouteHintHop struct{}

var FfiConverterBolt11RouteHintHopINSTANCE = FfiConverterBolt11RouteHintHop{}

func (c FfiConverterBolt11RouteHintHop) Lift(rb RustBufferI) Bolt11RouteHintHop {
	return LiftFromRustBuffer[Bolt11RouteHintHop](c, rb)
}

func (c FfiConverterBolt11RouteHintHop) Read(reader io.Reader) Bolt11RouteHintHop {
	return Bolt11RouteHintHop{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint16INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt11RouteHintHop) Lower(value Bolt11RouteHintHop) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt11RouteHintHop](c, value)
}

func (c FfiConverterBolt11RouteHintHop) Write(writer io.Writer, value Bolt11RouteHintHop) {
	FfiConverterStringINSTANCE.Write(writer, value.SrcNodeId)
	FfiConverterStringINSTANCE.Write(writer, value.ShortChannelId)
	FfiConverterUint32INSTANCE.Write(writer, value.FeesBaseMsat)
	FfiConverterUint32INSTANCE.Write(writer, value.FeesProportionalMillionths)
	FfiConverterUint16INSTANCE.Write(writer, value.CltvExpiryDelta)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.HtlcMinimumMsat)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.HtlcMaximumMsat)
}

type FfiDestroyerBolt11RouteHintHop struct{}

func (_ FfiDestroyerBolt11RouteHintHop) Destroy(value Bolt11RouteHintHop) {
	value.Destroy()
}

type Bolt12Invoice struct {
	Invoice string
	Source  PaymentRequestSource
}

func (r *Bolt12Invoice) Destroy() {
	FfiDestroyerString{}.Destroy(r.Invoice)
	FfiDestroyerPaymentRequestSource{}.Destroy(r.Source)
}

type FfiConverterBolt12Invoice struct{}

var FfiConverterBolt12InvoiceINSTANCE = FfiConverterBolt12Invoice{}

func (c FfiConverterBolt12Invoice) Lift(rb RustBufferI) Bolt12Invoice {
	return LiftFromRustBuffer[Bolt12Invoice](c, rb)
}

func (c FfiConverterBolt12Invoice) Read(reader io.Reader) Bolt12Invoice {
	return Bolt12Invoice{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterPaymentRequestSourceINSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt12Invoice) Lower(value Bolt12Invoice) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt12Invoice](c, value)
}

func (c FfiConverterBolt12Invoice) Write(writer io.Writer, value Bolt12Invoice) {
	FfiConverterStringINSTANCE.Write(writer, value.Invoice)
	FfiConverterPaymentRequestSourceINSTANCE.Write(writer, value.Source)
}

type FfiDestroyerBolt12Invoice struct{}

func (_ FfiDestroyerBolt12Invoice) Destroy(value Bolt12Invoice) {
	value.Destroy()
}

type Bolt12InvoiceDetails struct {
	AmountMsat uint64
	Invoice    Bolt12Invoice
}

func (r *Bolt12InvoiceDetails) Destroy() {
	FfiDestroyerUint64{}.Destroy(r.AmountMsat)
	FfiDestroyerBolt12Invoice{}.Destroy(r.Invoice)
}

type FfiConverterBolt12InvoiceDetails struct{}

var FfiConverterBolt12InvoiceDetailsINSTANCE = FfiConverterBolt12InvoiceDetails{}

func (c FfiConverterBolt12InvoiceDetails) Lift(rb RustBufferI) Bolt12InvoiceDetails {
	return LiftFromRustBuffer[Bolt12InvoiceDetails](c, rb)
}

func (c FfiConverterBolt12InvoiceDetails) Read(reader io.Reader) Bolt12InvoiceDetails {
	return Bolt12InvoiceDetails{
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterBolt12InvoiceINSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt12InvoiceDetails) Lower(value Bolt12InvoiceDetails) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt12InvoiceDetails](c, value)
}

func (c FfiConverterBolt12InvoiceDetails) Write(writer io.Writer, value Bolt12InvoiceDetails) {
	FfiConverterUint64INSTANCE.Write(writer, value.AmountMsat)
	FfiConverterBolt12InvoiceINSTANCE.Write(writer, value.Invoice)
}

type FfiDestroyerBolt12InvoiceDetails struct{}

func (_ FfiDestroyerBolt12InvoiceDetails) Destroy(value Bolt12InvoiceDetails) {
	value.Destroy()
}

type Bolt12InvoiceRequestDetails struct {
}

func (r *Bolt12InvoiceRequestDetails) Destroy() {
}

type FfiConverterBolt12InvoiceRequestDetails struct{}

var FfiConverterBolt12InvoiceRequestDetailsINSTANCE = FfiConverterBolt12InvoiceRequestDetails{}

func (c FfiConverterBolt12InvoiceRequestDetails) Lift(rb RustBufferI) Bolt12InvoiceRequestDetails {
	return LiftFromRustBuffer[Bolt12InvoiceRequestDetails](c, rb)
}

func (c FfiConverterBolt12InvoiceRequestDetails) Read(reader io.Reader) Bolt12InvoiceRequestDetails {
	return Bolt12InvoiceRequestDetails{}
}

func (c FfiConverterBolt12InvoiceRequestDetails) Lower(value Bolt12InvoiceRequestDetails) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt12InvoiceRequestDetails](c, value)
}

func (c FfiConverterBolt12InvoiceRequestDetails) Write(writer io.Writer, value Bolt12InvoiceRequestDetails) {
}

type FfiDestroyerBolt12InvoiceRequestDetails struct{}

func (_ FfiDestroyerBolt12InvoiceRequestDetails) Destroy(value Bolt12InvoiceRequestDetails) {
	value.Destroy()
}

type Bolt12Offer struct {
	Offer  string
	Source PaymentRequestSource
}

func (r *Bolt12Offer) Destroy() {
	FfiDestroyerString{}.Destroy(r.Offer)
	FfiDestroyerPaymentRequestSource{}.Destroy(r.Source)
}

type FfiConverterBolt12Offer struct{}

var FfiConverterBolt12OfferINSTANCE = FfiConverterBolt12Offer{}

func (c FfiConverterBolt12Offer) Lift(rb RustBufferI) Bolt12Offer {
	return LiftFromRustBuffer[Bolt12Offer](c, rb)
}

func (c FfiConverterBolt12Offer) Read(reader io.Reader) Bolt12Offer {
	return Bolt12Offer{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterPaymentRequestSourceINSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt12Offer) Lower(value Bolt12Offer) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt12Offer](c, value)
}

func (c FfiConverterBolt12Offer) Write(writer io.Writer, value Bolt12Offer) {
	FfiConverterStringINSTANCE.Write(writer, value.Offer)
	FfiConverterPaymentRequestSourceINSTANCE.Write(writer, value.Source)
}

type FfiDestroyerBolt12Offer struct{}

func (_ FfiDestroyerBolt12Offer) Destroy(value Bolt12Offer) {
	value.Destroy()
}

type Bolt12OfferBlindedPath struct {
	BlindedHops []string
}

func (r *Bolt12OfferBlindedPath) Destroy() {
	FfiDestroyerSequenceString{}.Destroy(r.BlindedHops)
}

type FfiConverterBolt12OfferBlindedPath struct{}

var FfiConverterBolt12OfferBlindedPathINSTANCE = FfiConverterBolt12OfferBlindedPath{}

func (c FfiConverterBolt12OfferBlindedPath) Lift(rb RustBufferI) Bolt12OfferBlindedPath {
	return LiftFromRustBuffer[Bolt12OfferBlindedPath](c, rb)
}

func (c FfiConverterBolt12OfferBlindedPath) Read(reader io.Reader) Bolt12OfferBlindedPath {
	return Bolt12OfferBlindedPath{
		FfiConverterSequenceStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt12OfferBlindedPath) Lower(value Bolt12OfferBlindedPath) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt12OfferBlindedPath](c, value)
}

func (c FfiConverterBolt12OfferBlindedPath) Write(writer io.Writer, value Bolt12OfferBlindedPath) {
	FfiConverterSequenceStringINSTANCE.Write(writer, value.BlindedHops)
}

type FfiDestroyerBolt12OfferBlindedPath struct{}

func (_ FfiDestroyerBolt12OfferBlindedPath) Destroy(value Bolt12OfferBlindedPath) {
	value.Destroy()
}

type Bolt12OfferDetails struct {
	AbsoluteExpiry *uint64
	Chains         []string
	Description    *string
	Issuer         *string
	MinAmount      *Amount
	Offer          Bolt12Offer
	Paths          []Bolt12OfferBlindedPath
	SigningPubkey  *string
}

func (r *Bolt12OfferDetails) Destroy() {
	FfiDestroyerOptionalUint64{}.Destroy(r.AbsoluteExpiry)
	FfiDestroyerSequenceString{}.Destroy(r.Chains)
	FfiDestroyerOptionalString{}.Destroy(r.Description)
	FfiDestroyerOptionalString{}.Destroy(r.Issuer)
	FfiDestroyerOptionalAmount{}.Destroy(r.MinAmount)
	FfiDestroyerBolt12Offer{}.Destroy(r.Offer)
	FfiDestroyerSequenceBolt12OfferBlindedPath{}.Destroy(r.Paths)
	FfiDestroyerOptionalString{}.Destroy(r.SigningPubkey)
}

type FfiConverterBolt12OfferDetails struct{}

var FfiConverterBolt12OfferDetailsINSTANCE = FfiConverterBolt12OfferDetails{}

func (c FfiConverterBolt12OfferDetails) Lift(rb RustBufferI) Bolt12OfferDetails {
	return LiftFromRustBuffer[Bolt12OfferDetails](c, rb)
}

func (c FfiConverterBolt12OfferDetails) Read(reader io.Reader) Bolt12OfferDetails {
	return Bolt12OfferDetails{
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterSequenceStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalAmountINSTANCE.Read(reader),
		FfiConverterBolt12OfferINSTANCE.Read(reader),
		FfiConverterSequenceBolt12OfferBlindedPathINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterBolt12OfferDetails) Lower(value Bolt12OfferDetails) C.RustBuffer {
	return LowerIntoRustBuffer[Bolt12OfferDetails](c, value)
}

func (c FfiConverterBolt12OfferDetails) Write(writer io.Writer, value Bolt12OfferDetails) {
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.AbsoluteExpiry)
	FfiConverterSequenceStringINSTANCE.Write(writer, value.Chains)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Description)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Issuer)
	FfiConverterOptionalAmountINSTANCE.Write(writer, value.MinAmount)
	FfiConverterBolt12OfferINSTANCE.Write(writer, value.Offer)
	FfiConverterSequenceBolt12OfferBlindedPathINSTANCE.Write(writer, value.Paths)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.SigningPubkey)
}

type FfiDestroyerBolt12OfferDetails struct{}

func (_ FfiDestroyerBolt12OfferDetails) Destroy(value Bolt12OfferDetails) {
	value.Destroy()
}

type BurnIssuerTokenRequest struct {
	Amount u128
}

func (r *BurnIssuerTokenRequest) Destroy() {
	FfiDestroyerTypeu128{}.Destroy(r.Amount)
}

type FfiConverterBurnIssuerTokenRequest struct{}

var FfiConverterBurnIssuerTokenRequestINSTANCE = FfiConverterBurnIssuerTokenRequest{}

func (c FfiConverterBurnIssuerTokenRequest) Lift(rb RustBufferI) BurnIssuerTokenRequest {
	return LiftFromRustBuffer[BurnIssuerTokenRequest](c, rb)
}

func (c FfiConverterBurnIssuerTokenRequest) Read(reader io.Reader) BurnIssuerTokenRequest {
	return BurnIssuerTokenRequest{
		FfiConverterTypeu128INSTANCE.Read(reader),
	}
}

func (c FfiConverterBurnIssuerTokenRequest) Lower(value BurnIssuerTokenRequest) C.RustBuffer {
	return LowerIntoRustBuffer[BurnIssuerTokenRequest](c, value)
}

func (c FfiConverterBurnIssuerTokenRequest) Write(writer io.Writer, value BurnIssuerTokenRequest) {
	FfiConverterTypeu128INSTANCE.Write(writer, value.Amount)
}

type FfiDestroyerBurnIssuerTokenRequest struct{}

func (_ FfiDestroyerBurnIssuerTokenRequest) Destroy(value BurnIssuerTokenRequest) {
	value.Destroy()
}

type CheckLightningAddressRequest struct {
	Username string
}

func (r *CheckLightningAddressRequest) Destroy() {
	FfiDestroyerString{}.Destroy(r.Username)
}

type FfiConverterCheckLightningAddressRequest struct{}

var FfiConverterCheckLightningAddressRequestINSTANCE = FfiConverterCheckLightningAddressRequest{}

func (c FfiConverterCheckLightningAddressRequest) Lift(rb RustBufferI) CheckLightningAddressRequest {
	return LiftFromRustBuffer[CheckLightningAddressRequest](c, rb)
}

func (c FfiConverterCheckLightningAddressRequest) Read(reader io.Reader) CheckLightningAddressRequest {
	return CheckLightningAddressRequest{
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterCheckLightningAddressRequest) Lower(value CheckLightningAddressRequest) C.RustBuffer {
	return LowerIntoRustBuffer[CheckLightningAddressRequest](c, value)
}

func (c FfiConverterCheckLightningAddressRequest) Write(writer io.Writer, value CheckLightningAddressRequest) {
	FfiConverterStringINSTANCE.Write(writer, value.Username)
}

type FfiDestroyerCheckLightningAddressRequest struct{}

func (_ FfiDestroyerCheckLightningAddressRequest) Destroy(value CheckLightningAddressRequest) {
	value.Destroy()
}

type CheckMessageRequest struct {
	// The message that was signed
	Message string
	// The public key that signed the message
	Pubkey string
	// The DER or compact hex encoded signature
	Signature string
}

func (r *CheckMessageRequest) Destroy() {
	FfiDestroyerString{}.Destroy(r.Message)
	FfiDestroyerString{}.Destroy(r.Pubkey)
	FfiDestroyerString{}.Destroy(r.Signature)
}

type FfiConverterCheckMessageRequest struct{}

var FfiConverterCheckMessageRequestINSTANCE = FfiConverterCheckMessageRequest{}

func (c FfiConverterCheckMessageRequest) Lift(rb RustBufferI) CheckMessageRequest {
	return LiftFromRustBuffer[CheckMessageRequest](c, rb)
}

func (c FfiConverterCheckMessageRequest) Read(reader io.Reader) CheckMessageRequest {
	return CheckMessageRequest{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterCheckMessageRequest) Lower(value CheckMessageRequest) C.RustBuffer {
	return LowerIntoRustBuffer[CheckMessageRequest](c, value)
}

func (c FfiConverterCheckMessageRequest) Write(writer io.Writer, value CheckMessageRequest) {
	FfiConverterStringINSTANCE.Write(writer, value.Message)
	FfiConverterStringINSTANCE.Write(writer, value.Pubkey)
	FfiConverterStringINSTANCE.Write(writer, value.Signature)
}

type FfiDestroyerCheckMessageRequest struct{}

func (_ FfiDestroyerCheckMessageRequest) Destroy(value CheckMessageRequest) {
	value.Destroy()
}

type CheckMessageResponse struct {
	IsValid bool
}

func (r *CheckMessageResponse) Destroy() {
	FfiDestroyerBool{}.Destroy(r.IsValid)
}

type FfiConverterCheckMessageResponse struct{}

var FfiConverterCheckMessageResponseINSTANCE = FfiConverterCheckMessageResponse{}

func (c FfiConverterCheckMessageResponse) Lift(rb RustBufferI) CheckMessageResponse {
	return LiftFromRustBuffer[CheckMessageResponse](c, rb)
}

func (c FfiConverterCheckMessageResponse) Read(reader io.Reader) CheckMessageResponse {
	return CheckMessageResponse{
		FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterCheckMessageResponse) Lower(value CheckMessageResponse) C.RustBuffer {
	return LowerIntoRustBuffer[CheckMessageResponse](c, value)
}

func (c FfiConverterCheckMessageResponse) Write(writer io.Writer, value CheckMessageResponse) {
	FfiConverterBoolINSTANCE.Write(writer, value.IsValid)
}

type FfiDestroyerCheckMessageResponse struct{}

func (_ FfiDestroyerCheckMessageResponse) Destroy(value CheckMessageResponse) {
	value.Destroy()
}

type ClaimDepositRequest struct {
	Txid   string
	Vout   uint32
	MaxFee *MaxFee
}

func (r *ClaimDepositRequest) Destroy() {
	FfiDestroyerString{}.Destroy(r.Txid)
	FfiDestroyerUint32{}.Destroy(r.Vout)
	FfiDestroyerOptionalMaxFee{}.Destroy(r.MaxFee)
}

type FfiConverterClaimDepositRequest struct{}

var FfiConverterClaimDepositRequestINSTANCE = FfiConverterClaimDepositRequest{}

func (c FfiConverterClaimDepositRequest) Lift(rb RustBufferI) ClaimDepositRequest {
	return LiftFromRustBuffer[ClaimDepositRequest](c, rb)
}

func (c FfiConverterClaimDepositRequest) Read(reader io.Reader) ClaimDepositRequest {
	return ClaimDepositRequest{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterOptionalMaxFeeINSTANCE.Read(reader),
	}
}

func (c FfiConverterClaimDepositRequest) Lower(value ClaimDepositRequest) C.RustBuffer {
	return LowerIntoRustBuffer[ClaimDepositRequest](c, value)
}

func (c FfiConverterClaimDepositRequest) Write(writer io.Writer, value ClaimDepositRequest) {
	FfiConverterStringINSTANCE.Write(writer, value.Txid)
	FfiConverterUint32INSTANCE.Write(writer, value.Vout)
	FfiConverterOptionalMaxFeeINSTANCE.Write(writer, value.MaxFee)
}

type FfiDestroyerClaimDepositRequest struct{}

func (_ FfiDestroyerClaimDepositRequest) Destroy(value ClaimDepositRequest) {
	value.Destroy()
}

type ClaimDepositResponse struct {
	Payment Payment
}

func (r *ClaimDepositResponse) Destroy() {
	FfiDestroyerPayment{}.Destroy(r.Payment)
}

type FfiConverterClaimDepositResponse struct{}

var FfiConverterClaimDepositResponseINSTANCE = FfiConverterClaimDepositResponse{}

func (c FfiConverterClaimDepositResponse) Lift(rb RustBufferI) ClaimDepositResponse {
	return LiftFromRustBuffer[ClaimDepositResponse](c, rb)
}

func (c FfiConverterClaimDepositResponse) Read(reader io.Reader) ClaimDepositResponse {
	return ClaimDepositResponse{
		FfiConverterPaymentINSTANCE.Read(reader),
	}
}

func (c FfiConverterClaimDepositResponse) Lower(value ClaimDepositResponse) C.RustBuffer {
	return LowerIntoRustBuffer[ClaimDepositResponse](c, value)
}

func (c FfiConverterClaimDepositResponse) Write(writer io.Writer, value ClaimDepositResponse) {
	FfiConverterPaymentINSTANCE.Write(writer, value.Payment)
}

type FfiDestroyerClaimDepositResponse struct{}

func (_ FfiDestroyerClaimDepositResponse) Destroy(value ClaimDepositResponse) {
	value.Destroy()
}

type ClaimHtlcPaymentRequest struct {
	Preimage string
}

func (r *ClaimHtlcPaymentRequest) Destroy() {
	FfiDestroyerString{}.Destroy(r.Preimage)
}

type FfiConverterClaimHtlcPaymentRequest struct{}

var FfiConverterClaimHtlcPaymentRequestINSTANCE = FfiConverterClaimHtlcPaymentRequest{}

func (c FfiConverterClaimHtlcPaymentRequest) Lift(rb RustBufferI) ClaimHtlcPaymentRequest {
	return LiftFromRustBuffer[ClaimHtlcPaymentRequest](c, rb)
}

func (c FfiConverterClaimHtlcPaymentRequest) Read(reader io.Reader) ClaimHtlcPaymentRequest {
	return ClaimHtlcPaymentRequest{
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterClaimHtlcPaymentRequest) Lower(value ClaimHtlcPaymentRequest) C.RustBuffer {
	return LowerIntoRustBuffer[ClaimHtlcPaymentRequest](c, value)
}

func (c FfiConverterClaimHtlcPaymentRequest) Write(writer io.Writer, value ClaimHtlcPaymentRequest) {
	FfiConverterStringINSTANCE.Write(writer, value.Preimage)
}

type FfiDestroyerClaimHtlcPaymentRequest struct{}

func (_ FfiDestroyerClaimHtlcPaymentRequest) Destroy(value ClaimHtlcPaymentRequest) {
	value.Destroy()
}

type ClaimHtlcPaymentResponse struct {
	Payment Payment
}

func (r *ClaimHtlcPaymentResponse) Destroy() {
	FfiDestroyerPayment{}.Destroy(r.Payment)
}

type FfiConverterClaimHtlcPaymentResponse struct{}

var FfiConverterClaimHtlcPaymentResponseINSTANCE = FfiConverterClaimHtlcPaymentResponse{}

func (c FfiConverterClaimHtlcPaymentResponse) Lift(rb RustBufferI) ClaimHtlcPaymentResponse {
	return LiftFromRustBuffer[ClaimHtlcPaymentResponse](c, rb)
}

func (c FfiConverterClaimHtlcPaymentResponse) Read(reader io.Reader) ClaimHtlcPaymentResponse {
	return ClaimHtlcPaymentResponse{
		FfiConverterPaymentINSTANCE.Read(reader),
	}
}

func (c FfiConverterClaimHtlcPaymentResponse) Lower(value ClaimHtlcPaymentResponse) C.RustBuffer {
	return LowerIntoRustBuffer[ClaimHtlcPaymentResponse](c, value)
}

func (c FfiConverterClaimHtlcPaymentResponse) Write(writer io.Writer, value ClaimHtlcPaymentResponse) {
	FfiConverterPaymentINSTANCE.Write(writer, value.Payment)
}

type FfiDestroyerClaimHtlcPaymentResponse struct{}

func (_ FfiDestroyerClaimHtlcPaymentResponse) Destroy(value ClaimHtlcPaymentResponse) {
	value.Destroy()
}

type Config struct {
	ApiKey             *string
	Network            Network
	SyncIntervalSecs   uint32
	MaxDepositClaimFee *MaxFee
	// The domain used for receiving through lnurl-pay and lightning address.
	LnurlDomain *string
	// When this is set to `true` we will prefer to use spark payments over
	// lightning when sending and receiving. This has the benefit of lower fees
	// but is at the cost of privacy.
	PreferSparkOverLightning bool
	// A set of external input parsers that are used by [`BreezSdk::parse`](crate::sdk::BreezSdk::parse) when the input
	// is not recognized. See [`ExternalInputParser`] for more details on how to configure
	// external parsing.
	ExternalInputParsers *[]ExternalInputParser
	// The SDK includes some default external input parsers
	// ([`DEFAULT_EXTERNAL_INPUT_PARSERS`]).
	// Set this to false in order to prevent their use.
	UseDefaultExternalInputParsers bool
	// Url to use for the real-time sync server. Defaults to the Breez real-time sync server.
	RealTimeSyncServerUrl *string
	// Whether the Spark private mode is enabled by default.
	//
	// If set to true, the Spark private mode will be enabled on the first initialization of the SDK.
	// If set to false, no changes will be made to the Spark private mode.
	PrivateEnabledDefault bool
}

func (r *Config) Destroy() {
	FfiDestroyerOptionalString{}.Destroy(r.ApiKey)
	FfiDestroyerNetwork{}.Destroy(r.Network)
	FfiDestroyerUint32{}.Destroy(r.SyncIntervalSecs)
	FfiDestroyerOptionalMaxFee{}.Destroy(r.MaxDepositClaimFee)
	FfiDestroyerOptionalString{}.Destroy(r.LnurlDomain)
	FfiDestroyerBool{}.Destroy(r.PreferSparkOverLightning)
	FfiDestroyerOptionalSequenceExternalInputParser{}.Destroy(r.ExternalInputParsers)
	FfiDestroyerBool{}.Destroy(r.UseDefaultExternalInputParsers)
	FfiDestroyerOptionalString{}.Destroy(r.RealTimeSyncServerUrl)
	FfiDestroyerBool{}.Destroy(r.PrivateEnabledDefault)
}

type FfiConverterConfig struct{}

var FfiConverterConfigINSTANCE = FfiConverterConfig{}

func (c FfiConverterConfig) Lift(rb RustBufferI) Config {
	return LiftFromRustBuffer[Config](c, rb)
}

func (c FfiConverterConfig) Read(reader io.Reader) Config {
	return Config{
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterNetworkINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterOptionalMaxFeeINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterOptionalSequenceExternalInputParserINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterConfig) Lower(value Config) C.RustBuffer {
	return LowerIntoRustBuffer[Config](c, value)
}

func (c FfiConverterConfig) Write(writer io.Writer, value Config) {
	FfiConverterOptionalStringINSTANCE.Write(writer, value.ApiKey)
	FfiConverterNetworkINSTANCE.Write(writer, value.Network)
	FfiConverterUint32INSTANCE.Write(writer, value.SyncIntervalSecs)
	FfiConverterOptionalMaxFeeINSTANCE.Write(writer, value.MaxDepositClaimFee)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.LnurlDomain)
	FfiConverterBoolINSTANCE.Write(writer, value.PreferSparkOverLightning)
	FfiConverterOptionalSequenceExternalInputParserINSTANCE.Write(writer, value.ExternalInputParsers)
	FfiConverterBoolINSTANCE.Write(writer, value.UseDefaultExternalInputParsers)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.RealTimeSyncServerUrl)
	FfiConverterBoolINSTANCE.Write(writer, value.PrivateEnabledDefault)
}

type FfiDestroyerConfig struct{}

func (_ FfiDestroyerConfig) Destroy(value Config) {
	value.Destroy()
}

type ConnectRequest struct {
	Config     Config
	Seed       Seed
	StorageDir string
}

func (r *ConnectRequest) Destroy() {
	FfiDestroyerConfig{}.Destroy(r.Config)
	FfiDestroyerSeed{}.Destroy(r.Seed)
	FfiDestroyerString{}.Destroy(r.StorageDir)
}

type FfiConverterConnectRequest struct{}

var FfiConverterConnectRequestINSTANCE = FfiConverterConnectRequest{}

func (c FfiConverterConnectRequest) Lift(rb RustBufferI) ConnectRequest {
	return LiftFromRustBuffer[ConnectRequest](c, rb)
}

func (c FfiConverterConnectRequest) Read(reader io.Reader) ConnectRequest {
	return ConnectRequest{
		FfiConverterConfigINSTANCE.Read(reader),
		FfiConverterSeedINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterConnectRequest) Lower(value ConnectRequest) C.RustBuffer {
	return LowerIntoRustBuffer[ConnectRequest](c, value)
}

func (c FfiConverterConnectRequest) Write(writer io.Writer, value ConnectRequest) {
	FfiConverterConfigINSTANCE.Write(writer, value.Config)
	FfiConverterSeedINSTANCE.Write(writer, value.Seed)
	FfiConverterStringINSTANCE.Write(writer, value.StorageDir)
}

type FfiDestroyerConnectRequest struct{}

func (_ FfiDestroyerConnectRequest) Destroy(value ConnectRequest) {
	value.Destroy()
}

type CreateIssuerTokenRequest struct {
	Name        string
	Ticker      string
	Decimals    uint32
	IsFreezable bool
	MaxSupply   u128
}

func (r *CreateIssuerTokenRequest) Destroy() {
	FfiDestroyerString{}.Destroy(r.Name)
	FfiDestroyerString{}.Destroy(r.Ticker)
	FfiDestroyerUint32{}.Destroy(r.Decimals)
	FfiDestroyerBool{}.Destroy(r.IsFreezable)
	FfiDestroyerTypeu128{}.Destroy(r.MaxSupply)
}

type FfiConverterCreateIssuerTokenRequest struct{}

var FfiConverterCreateIssuerTokenRequestINSTANCE = FfiConverterCreateIssuerTokenRequest{}

func (c FfiConverterCreateIssuerTokenRequest) Lift(rb RustBufferI) CreateIssuerTokenRequest {
	return LiftFromRustBuffer[CreateIssuerTokenRequest](c, rb)
}

func (c FfiConverterCreateIssuerTokenRequest) Read(reader io.Reader) CreateIssuerTokenRequest {
	return CreateIssuerTokenRequest{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterTypeu128INSTANCE.Read(reader),
	}
}

func (c FfiConverterCreateIssuerTokenRequest) Lower(value CreateIssuerTokenRequest) C.RustBuffer {
	return LowerIntoRustBuffer[CreateIssuerTokenRequest](c, value)
}

func (c FfiConverterCreateIssuerTokenRequest) Write(writer io.Writer, value CreateIssuerTokenRequest) {
	FfiConverterStringINSTANCE.Write(writer, value.Name)
	FfiConverterStringINSTANCE.Write(writer, value.Ticker)
	FfiConverterUint32INSTANCE.Write(writer, value.Decimals)
	FfiConverterBoolINSTANCE.Write(writer, value.IsFreezable)
	FfiConverterTypeu128INSTANCE.Write(writer, value.MaxSupply)
}

type FfiDestroyerCreateIssuerTokenRequest struct{}

func (_ FfiDestroyerCreateIssuerTokenRequest) Destroy(value CreateIssuerTokenRequest) {
	value.Destroy()
}

type Credentials struct {
	Username string
	Password string
}

func (r *Credentials) Destroy() {
	FfiDestroyerString{}.Destroy(r.Username)
	FfiDestroyerString{}.Destroy(r.Password)
}

type FfiConverterCredentials struct{}

var FfiConverterCredentialsINSTANCE = FfiConverterCredentials{}

func (c FfiConverterCredentials) Lift(rb RustBufferI) Credentials {
	return LiftFromRustBuffer[Credentials](c, rb)
}

func (c FfiConverterCredentials) Read(reader io.Reader) Credentials {
	return Credentials{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterCredentials) Lower(value Credentials) C.RustBuffer {
	return LowerIntoRustBuffer[Credentials](c, value)
}

func (c FfiConverterCredentials) Write(writer io.Writer, value Credentials) {
	FfiConverterStringINSTANCE.Write(writer, value.Username)
	FfiConverterStringINSTANCE.Write(writer, value.Password)
}

type FfiDestroyerCredentials struct{}

func (_ FfiDestroyerCredentials) Destroy(value Credentials) {
	value.Destroy()
}

// Details about a supported currency in the fiat rate feed
type CurrencyInfo struct {
	Name            string
	FractionSize    uint32
	Spacing         *uint32
	Symbol          *Symbol
	UniqSymbol      *Symbol
	LocalizedName   []LocalizedName
	LocaleOverrides []LocaleOverrides
}

func (r *CurrencyInfo) Destroy() {
	FfiDestroyerString{}.Destroy(r.Name)
	FfiDestroyerUint32{}.Destroy(r.FractionSize)
	FfiDestroyerOptionalUint32{}.Destroy(r.Spacing)
	FfiDestroyerOptionalSymbol{}.Destroy(r.Symbol)
	FfiDestroyerOptionalSymbol{}.Destroy(r.UniqSymbol)
	FfiDestroyerSequenceLocalizedName{}.Destroy(r.LocalizedName)
	FfiDestroyerSequenceLocaleOverrides{}.Destroy(r.LocaleOverrides)
}

type FfiConverterCurrencyInfo struct{}

var FfiConverterCurrencyInfoINSTANCE = FfiConverterCurrencyInfo{}

func (c FfiConverterCurrencyInfo) Lift(rb RustBufferI) CurrencyInfo {
	return LiftFromRustBuffer[CurrencyInfo](c, rb)
}

func (c FfiConverterCurrencyInfo) Read(reader io.Reader) CurrencyInfo {
	return CurrencyInfo{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalSymbolINSTANCE.Read(reader),
		FfiConverterOptionalSymbolINSTANCE.Read(reader),
		FfiConverterSequenceLocalizedNameINSTANCE.Read(reader),
		FfiConverterSequenceLocaleOverridesINSTANCE.Read(reader),
	}
}

func (c FfiConverterCurrencyInfo) Lower(value CurrencyInfo) C.RustBuffer {
	return LowerIntoRustBuffer[CurrencyInfo](c, value)
}

func (c FfiConverterCurrencyInfo) Write(writer io.Writer, value CurrencyInfo) {
	FfiConverterStringINSTANCE.Write(writer, value.Name)
	FfiConverterUint32INSTANCE.Write(writer, value.FractionSize)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.Spacing)
	FfiConverterOptionalSymbolINSTANCE.Write(writer, value.Symbol)
	FfiConverterOptionalSymbolINSTANCE.Write(writer, value.UniqSymbol)
	FfiConverterSequenceLocalizedNameINSTANCE.Write(writer, value.LocalizedName)
	FfiConverterSequenceLocaleOverridesINSTANCE.Write(writer, value.LocaleOverrides)
}

type FfiDestroyerCurrencyInfo struct{}

func (_ FfiDestroyerCurrencyInfo) Destroy(value CurrencyInfo) {
	value.Destroy()
}

type DepositInfo struct {
	Txid       string
	Vout       uint32
	AmountSats uint64
	RefundTx   *string
	RefundTxId *string
	ClaimError *DepositClaimError
}

func (r *DepositInfo) Destroy() {
	FfiDestroyerString{}.Destroy(r.Txid)
	FfiDestroyerUint32{}.Destroy(r.Vout)
	FfiDestroyerUint64{}.Destroy(r.AmountSats)
	FfiDestroyerOptionalString{}.Destroy(r.RefundTx)
	FfiDestroyerOptionalString{}.Destroy(r.RefundTxId)
	FfiDestroyerOptionalDepositClaimError{}.Destroy(r.ClaimError)
}

type FfiConverterDepositInfo struct{}

var FfiConverterDepositInfoINSTANCE = FfiConverterDepositInfo{}

func (c FfiConverterDepositInfo) Lift(rb RustBufferI) DepositInfo {
	return LiftFromRustBuffer[DepositInfo](c, rb)
}

func (c FfiConverterDepositInfo) Read(reader io.Reader) DepositInfo {
	return DepositInfo{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalDepositClaimErrorINSTANCE.Read(reader),
	}
}

func (c FfiConverterDepositInfo) Lower(value DepositInfo) C.RustBuffer {
	return LowerIntoRustBuffer[DepositInfo](c, value)
}

func (c FfiConverterDepositInfo) Write(writer io.Writer, value DepositInfo) {
	FfiConverterStringINSTANCE.Write(writer, value.Txid)
	FfiConverterUint32INSTANCE.Write(writer, value.Vout)
	FfiConverterUint64INSTANCE.Write(writer, value.AmountSats)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.RefundTx)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.RefundTxId)
	FfiConverterOptionalDepositClaimErrorINSTANCE.Write(writer, value.ClaimError)
}

type FfiDestroyerDepositInfo struct{}

func (_ FfiDestroyerDepositInfo) Destroy(value DepositInfo) {
	value.Destroy()
}

// Configuration for an external input parser
type ExternalInputParser struct {
	// An arbitrary parser provider id
	ProviderId string
	// The external parser will be used when an input conforms to this regex
	InputRegex string
	// The URL of the parser containing a placeholder `<input>` that will be replaced with the
	// input to be parsed. The input is sanitized using percent encoding.
	ParserUrl string
}

func (r *ExternalInputParser) Destroy() {
	FfiDestroyerString{}.Destroy(r.ProviderId)
	FfiDestroyerString{}.Destroy(r.InputRegex)
	FfiDestroyerString{}.Destroy(r.ParserUrl)
}

type FfiConverterExternalInputParser struct{}

var FfiConverterExternalInputParserINSTANCE = FfiConverterExternalInputParser{}

func (c FfiConverterExternalInputParser) Lift(rb RustBufferI) ExternalInputParser {
	return LiftFromRustBuffer[ExternalInputParser](c, rb)
}

func (c FfiConverterExternalInputParser) Read(reader io.Reader) ExternalInputParser {
	return ExternalInputParser{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterExternalInputParser) Lower(value ExternalInputParser) C.RustBuffer {
	return LowerIntoRustBuffer[ExternalInputParser](c, value)
}

func (c FfiConverterExternalInputParser) Write(writer io.Writer, value ExternalInputParser) {
	FfiConverterStringINSTANCE.Write(writer, value.ProviderId)
	FfiConverterStringINSTANCE.Write(writer, value.InputRegex)
	FfiConverterStringINSTANCE.Write(writer, value.ParserUrl)
}

type FfiDestroyerExternalInputParser struct{}

func (_ FfiDestroyerExternalInputParser) Destroy(value ExternalInputParser) {
	value.Destroy()
}

// Wrapper around the [`CurrencyInfo`] of a fiat currency
type FiatCurrency struct {
	Id   string
	Info CurrencyInfo
}

func (r *FiatCurrency) Destroy() {
	FfiDestroyerString{}.Destroy(r.Id)
	FfiDestroyerCurrencyInfo{}.Destroy(r.Info)
}

type FfiConverterFiatCurrency struct{}

var FfiConverterFiatCurrencyINSTANCE = FfiConverterFiatCurrency{}

func (c FfiConverterFiatCurrency) Lift(rb RustBufferI) FiatCurrency {
	return LiftFromRustBuffer[FiatCurrency](c, rb)
}

func (c FfiConverterFiatCurrency) Read(reader io.Reader) FiatCurrency {
	return FiatCurrency{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterCurrencyInfoINSTANCE.Read(reader),
	}
}

func (c FfiConverterFiatCurrency) Lower(value FiatCurrency) C.RustBuffer {
	return LowerIntoRustBuffer[FiatCurrency](c, value)
}

func (c FfiConverterFiatCurrency) Write(writer io.Writer, value FiatCurrency) {
	FfiConverterStringINSTANCE.Write(writer, value.Id)
	FfiConverterCurrencyInfoINSTANCE.Write(writer, value.Info)
}

type FfiDestroyerFiatCurrency struct{}

func (_ FfiDestroyerFiatCurrency) Destroy(value FiatCurrency) {
	value.Destroy()
}

type FreezeIssuerTokenRequest struct {
	Address string
}

func (r *FreezeIssuerTokenRequest) Destroy() {
	FfiDestroyerString{}.Destroy(r.Address)
}

type FfiConverterFreezeIssuerTokenRequest struct{}

var FfiConverterFreezeIssuerTokenRequestINSTANCE = FfiConverterFreezeIssuerTokenRequest{}

func (c FfiConverterFreezeIssuerTokenRequest) Lift(rb RustBufferI) FreezeIssuerTokenRequest {
	return LiftFromRustBuffer[FreezeIssuerTokenRequest](c, rb)
}

func (c FfiConverterFreezeIssuerTokenRequest) Read(reader io.Reader) FreezeIssuerTokenRequest {
	return FreezeIssuerTokenRequest{
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterFreezeIssuerTokenRequest) Lower(value FreezeIssuerTokenRequest) C.RustBuffer {
	return LowerIntoRustBuffer[FreezeIssuerTokenRequest](c, value)
}

func (c FfiConverterFreezeIssuerTokenRequest) Write(writer io.Writer, value FreezeIssuerTokenRequest) {
	FfiConverterStringINSTANCE.Write(writer, value.Address)
}

type FfiDestroyerFreezeIssuerTokenRequest struct{}

func (_ FfiDestroyerFreezeIssuerTokenRequest) Destroy(value FreezeIssuerTokenRequest) {
	value.Destroy()
}

type FreezeIssuerTokenResponse struct {
	ImpactedOutputIds   []string
	ImpactedTokenAmount u128
}

func (r *FreezeIssuerTokenResponse) Destroy() {
	FfiDestroyerSequenceString{}.Destroy(r.ImpactedOutputIds)
	FfiDestroyerTypeu128{}.Destroy(r.ImpactedTokenAmount)
}

type FfiConverterFreezeIssuerTokenResponse struct{}

var FfiConverterFreezeIssuerTokenResponseINSTANCE = FfiConverterFreezeIssuerTokenResponse{}

func (c FfiConverterFreezeIssuerTokenResponse) Lift(rb RustBufferI) FreezeIssuerTokenResponse {
	return LiftFromRustBuffer[FreezeIssuerTokenResponse](c, rb)
}

func (c FfiConverterFreezeIssuerTokenResponse) Read(reader io.Reader) FreezeIssuerTokenResponse {
	return FreezeIssuerTokenResponse{
		FfiConverterSequenceStringINSTANCE.Read(reader),
		FfiConverterTypeu128INSTANCE.Read(reader),
	}
}

func (c FfiConverterFreezeIssuerTokenResponse) Lower(value FreezeIssuerTokenResponse) C.RustBuffer {
	return LowerIntoRustBuffer[FreezeIssuerTokenResponse](c, value)
}

func (c FfiConverterFreezeIssuerTokenResponse) Write(writer io.Writer, value FreezeIssuerTokenResponse) {
	FfiConverterSequenceStringINSTANCE.Write(writer, value.ImpactedOutputIds)
	FfiConverterTypeu128INSTANCE.Write(writer, value.ImpactedTokenAmount)
}

type FfiDestroyerFreezeIssuerTokenResponse struct{}

func (_ FfiDestroyerFreezeIssuerTokenResponse) Destroy(value FreezeIssuerTokenResponse) {
	value.Destroy()
}

// Request to get the balance of the wallet
type GetInfoRequest struct {
	EnsureSynced *bool
}

func (r *GetInfoRequest) Destroy() {
	FfiDestroyerOptionalBool{}.Destroy(r.EnsureSynced)
}

type FfiConverterGetInfoRequest struct{}

var FfiConverterGetInfoRequestINSTANCE = FfiConverterGetInfoRequest{}

func (c FfiConverterGetInfoRequest) Lift(rb RustBufferI) GetInfoRequest {
	return LiftFromRustBuffer[GetInfoRequest](c, rb)
}

func (c FfiConverterGetInfoRequest) Read(reader io.Reader) GetInfoRequest {
	return GetInfoRequest{
		FfiConverterOptionalBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterGetInfoRequest) Lower(value GetInfoRequest) C.RustBuffer {
	return LowerIntoRustBuffer[GetInfoRequest](c, value)
}

func (c FfiConverterGetInfoRequest) Write(writer io.Writer, value GetInfoRequest) {
	FfiConverterOptionalBoolINSTANCE.Write(writer, value.EnsureSynced)
}

type FfiDestroyerGetInfoRequest struct{}

func (_ FfiDestroyerGetInfoRequest) Destroy(value GetInfoRequest) {
	value.Destroy()
}

// Response containing the balance of the wallet
type GetInfoResponse struct {
	// The balance in satoshis
	BalanceSats uint64
	// The balances of the tokens in the wallet keyed by the token identifier
	TokenBalances map[string]TokenBalance
}

func (r *GetInfoResponse) Destroy() {
	FfiDestroyerUint64{}.Destroy(r.BalanceSats)
	FfiDestroyerMapStringTokenBalance{}.Destroy(r.TokenBalances)
}

type FfiConverterGetInfoResponse struct{}

var FfiConverterGetInfoResponseINSTANCE = FfiConverterGetInfoResponse{}

func (c FfiConverterGetInfoResponse) Lift(rb RustBufferI) GetInfoResponse {
	return LiftFromRustBuffer[GetInfoResponse](c, rb)
}

func (c FfiConverterGetInfoResponse) Read(reader io.Reader) GetInfoResponse {
	return GetInfoResponse{
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterMapStringTokenBalanceINSTANCE.Read(reader),
	}
}

func (c FfiConverterGetInfoResponse) Lower(value GetInfoResponse) C.RustBuffer {
	return LowerIntoRustBuffer[GetInfoResponse](c, value)
}

func (c FfiConverterGetInfoResponse) Write(writer io.Writer, value GetInfoResponse) {
	FfiConverterUint64INSTANCE.Write(writer, value.BalanceSats)
	FfiConverterMapStringTokenBalanceINSTANCE.Write(writer, value.TokenBalances)
}

type FfiDestroyerGetInfoResponse struct{}

func (_ FfiDestroyerGetInfoResponse) Destroy(value GetInfoResponse) {
	value.Destroy()
}

type GetPaymentRequest struct {
	PaymentId string
}

func (r *GetPaymentRequest) Destroy() {
	FfiDestroyerString{}.Destroy(r.PaymentId)
}

type FfiConverterGetPaymentRequest struct{}

var FfiConverterGetPaymentRequestINSTANCE = FfiConverterGetPaymentRequest{}

func (c FfiConverterGetPaymentRequest) Lift(rb RustBufferI) GetPaymentRequest {
	return LiftFromRustBuffer[GetPaymentRequest](c, rb)
}

func (c FfiConverterGetPaymentRequest) Read(reader io.Reader) GetPaymentRequest {
	return GetPaymentRequest{
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterGetPaymentRequest) Lower(value GetPaymentRequest) C.RustBuffer {
	return LowerIntoRustBuffer[GetPaymentRequest](c, value)
}

func (c FfiConverterGetPaymentRequest) Write(writer io.Writer, value GetPaymentRequest) {
	FfiConverterStringINSTANCE.Write(writer, value.PaymentId)
}

type FfiDestroyerGetPaymentRequest struct{}

func (_ FfiDestroyerGetPaymentRequest) Destroy(value GetPaymentRequest) {
	value.Destroy()
}

type GetPaymentResponse struct {
	Payment Payment
}

func (r *GetPaymentResponse) Destroy() {
	FfiDestroyerPayment{}.Destroy(r.Payment)
}

type FfiConverterGetPaymentResponse struct{}

var FfiConverterGetPaymentResponseINSTANCE = FfiConverterGetPaymentResponse{}

func (c FfiConverterGetPaymentResponse) Lift(rb RustBufferI) GetPaymentResponse {
	return LiftFromRustBuffer[GetPaymentResponse](c, rb)
}

func (c FfiConverterGetPaymentResponse) Read(reader io.Reader) GetPaymentResponse {
	return GetPaymentResponse{
		FfiConverterPaymentINSTANCE.Read(reader),
	}
}

func (c FfiConverterGetPaymentResponse) Lower(value GetPaymentResponse) C.RustBuffer {
	return LowerIntoRustBuffer[GetPaymentResponse](c, value)
}

func (c FfiConverterGetPaymentResponse) Write(writer io.Writer, value GetPaymentResponse) {
	FfiConverterPaymentINSTANCE.Write(writer, value.Payment)
}

type FfiDestroyerGetPaymentResponse struct{}

func (_ FfiDestroyerGetPaymentResponse) Destroy(value GetPaymentResponse) {
	value.Destroy()
}

type GetTokensMetadataRequest struct {
	TokenIdentifiers []string
}

func (r *GetTokensMetadataRequest) Destroy() {
	FfiDestroyerSequenceString{}.Destroy(r.TokenIdentifiers)
}

type FfiConverterGetTokensMetadataRequest struct{}

var FfiConverterGetTokensMetadataRequestINSTANCE = FfiConverterGetTokensMetadataRequest{}

func (c FfiConverterGetTokensMetadataRequest) Lift(rb RustBufferI) GetTokensMetadataRequest {
	return LiftFromRustBuffer[GetTokensMetadataRequest](c, rb)
}

func (c FfiConverterGetTokensMetadataRequest) Read(reader io.Reader) GetTokensMetadataRequest {
	return GetTokensMetadataRequest{
		FfiConverterSequenceStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterGetTokensMetadataRequest) Lower(value GetTokensMetadataRequest) C.RustBuffer {
	return LowerIntoRustBuffer[GetTokensMetadataRequest](c, value)
}

func (c FfiConverterGetTokensMetadataRequest) Write(writer io.Writer, value GetTokensMetadataRequest) {
	FfiConverterSequenceStringINSTANCE.Write(writer, value.TokenIdentifiers)
}

type FfiDestroyerGetTokensMetadataRequest struct{}

func (_ FfiDestroyerGetTokensMetadataRequest) Destroy(value GetTokensMetadataRequest) {
	value.Destroy()
}

type GetTokensMetadataResponse struct {
	TokensMetadata []TokenMetadata
}

func (r *GetTokensMetadataResponse) Destroy() {
	FfiDestroyerSequenceTokenMetadata{}.Destroy(r.TokensMetadata)
}

type FfiConverterGetTokensMetadataResponse struct{}

var FfiConverterGetTokensMetadataResponseINSTANCE = FfiConverterGetTokensMetadataResponse{}

func (c FfiConverterGetTokensMetadataResponse) Lift(rb RustBufferI) GetTokensMetadataResponse {
	return LiftFromRustBuffer[GetTokensMetadataResponse](c, rb)
}

func (c FfiConverterGetTokensMetadataResponse) Read(reader io.Reader) GetTokensMetadataResponse {
	return GetTokensMetadataResponse{
		FfiConverterSequenceTokenMetadataINSTANCE.Read(reader),
	}
}

func (c FfiConverterGetTokensMetadataResponse) Lower(value GetTokensMetadataResponse) C.RustBuffer {
	return LowerIntoRustBuffer[GetTokensMetadataResponse](c, value)
}

func (c FfiConverterGetTokensMetadataResponse) Write(writer io.Writer, value GetTokensMetadataResponse) {
	FfiConverterSequenceTokenMetadataINSTANCE.Write(writer, value.TokensMetadata)
}

type FfiDestroyerGetTokensMetadataResponse struct{}

func (_ FfiDestroyerGetTokensMetadataResponse) Destroy(value GetTokensMetadataResponse) {
	value.Destroy()
}

type IncomingChange struct {
	NewState Record
	OldState *Record
}

func (r *IncomingChange) Destroy() {
	FfiDestroyerRecord{}.Destroy(r.NewState)
	FfiDestroyerOptionalRecord{}.Destroy(r.OldState)
}

type FfiConverterIncomingChange struct{}

var FfiConverterIncomingChangeINSTANCE = FfiConverterIncomingChange{}

func (c FfiConverterIncomingChange) Lift(rb RustBufferI) IncomingChange {
	return LiftFromRustBuffer[IncomingChange](c, rb)
}

func (c FfiConverterIncomingChange) Read(reader io.Reader) IncomingChange {
	return IncomingChange{
		FfiConverterRecordINSTANCE.Read(reader),
		FfiConverterOptionalRecordINSTANCE.Read(reader),
	}
}

func (c FfiConverterIncomingChange) Lower(value IncomingChange) C.RustBuffer {
	return LowerIntoRustBuffer[IncomingChange](c, value)
}

func (c FfiConverterIncomingChange) Write(writer io.Writer, value IncomingChange) {
	FfiConverterRecordINSTANCE.Write(writer, value.NewState)
	FfiConverterOptionalRecordINSTANCE.Write(writer, value.OldState)
}

type FfiDestroyerIncomingChange struct{}

func (_ FfiDestroyerIncomingChange) Destroy(value IncomingChange) {
	value.Destroy()
}

type LightningAddressDetails struct {
	Address    string
	PayRequest LnurlPayRequestDetails
}

func (r *LightningAddressDetails) Destroy() {
	FfiDestroyerString{}.Destroy(r.Address)
	FfiDestroyerLnurlPayRequestDetails{}.Destroy(r.PayRequest)
}

type FfiConverterLightningAddressDetails struct{}

var FfiConverterLightningAddressDetailsINSTANCE = FfiConverterLightningAddressDetails{}

func (c FfiConverterLightningAddressDetails) Lift(rb RustBufferI) LightningAddressDetails {
	return LiftFromRustBuffer[LightningAddressDetails](c, rb)
}

func (c FfiConverterLightningAddressDetails) Read(reader io.Reader) LightningAddressDetails {
	return LightningAddressDetails{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterLnurlPayRequestDetailsINSTANCE.Read(reader),
	}
}

func (c FfiConverterLightningAddressDetails) Lower(value LightningAddressDetails) C.RustBuffer {
	return LowerIntoRustBuffer[LightningAddressDetails](c, value)
}

func (c FfiConverterLightningAddressDetails) Write(writer io.Writer, value LightningAddressDetails) {
	FfiConverterStringINSTANCE.Write(writer, value.Address)
	FfiConverterLnurlPayRequestDetailsINSTANCE.Write(writer, value.PayRequest)
}

type FfiDestroyerLightningAddressDetails struct{}

func (_ FfiDestroyerLightningAddressDetails) Destroy(value LightningAddressDetails) {
	value.Destroy()
}

type LightningAddressInfo struct {
	Description      string
	LightningAddress string
	Lnurl            string
	Username         string
}

func (r *LightningAddressInfo) Destroy() {
	FfiDestroyerString{}.Destroy(r.Description)
	FfiDestroyerString{}.Destroy(r.LightningAddress)
	FfiDestroyerString{}.Destroy(r.Lnurl)
	FfiDestroyerString{}.Destroy(r.Username)
}

type FfiConverterLightningAddressInfo struct{}

var FfiConverterLightningAddressInfoINSTANCE = FfiConverterLightningAddressInfo{}

func (c FfiConverterLightningAddressInfo) Lift(rb RustBufferI) LightningAddressInfo {
	return LiftFromRustBuffer[LightningAddressInfo](c, rb)
}

func (c FfiConverterLightningAddressInfo) Read(reader io.Reader) LightningAddressInfo {
	return LightningAddressInfo{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterLightningAddressInfo) Lower(value LightningAddressInfo) C.RustBuffer {
	return LowerIntoRustBuffer[LightningAddressInfo](c, value)
}

func (c FfiConverterLightningAddressInfo) Write(writer io.Writer, value LightningAddressInfo) {
	FfiConverterStringINSTANCE.Write(writer, value.Description)
	FfiConverterStringINSTANCE.Write(writer, value.LightningAddress)
	FfiConverterStringINSTANCE.Write(writer, value.Lnurl)
	FfiConverterStringINSTANCE.Write(writer, value.Username)
}

type FfiDestroyerLightningAddressInfo struct{}

func (_ FfiDestroyerLightningAddressInfo) Destroy(value LightningAddressInfo) {
	value.Destroy()
}

// Response from listing fiat currencies
type ListFiatCurrenciesResponse struct {
	// The list of fiat currencies
	Currencies []FiatCurrency
}

func (r *ListFiatCurrenciesResponse) Destroy() {
	FfiDestroyerSequenceFiatCurrency{}.Destroy(r.Currencies)
}

type FfiConverterListFiatCurrenciesResponse struct{}

var FfiConverterListFiatCurrenciesResponseINSTANCE = FfiConverterListFiatCurrenciesResponse{}

func (c FfiConverterListFiatCurrenciesResponse) Lift(rb RustBufferI) ListFiatCurrenciesResponse {
	return LiftFromRustBuffer[ListFiatCurrenciesResponse](c, rb)
}

func (c FfiConverterListFiatCurrenciesResponse) Read(reader io.Reader) ListFiatCurrenciesResponse {
	return ListFiatCurrenciesResponse{
		FfiConverterSequenceFiatCurrencyINSTANCE.Read(reader),
	}
}

func (c FfiConverterListFiatCurrenciesResponse) Lower(value ListFiatCurrenciesResponse) C.RustBuffer {
	return LowerIntoRustBuffer[ListFiatCurrenciesResponse](c, value)
}

func (c FfiConverterListFiatCurrenciesResponse) Write(writer io.Writer, value ListFiatCurrenciesResponse) {
	FfiConverterSequenceFiatCurrencyINSTANCE.Write(writer, value.Currencies)
}

type FfiDestroyerListFiatCurrenciesResponse struct{}

func (_ FfiDestroyerListFiatCurrenciesResponse) Destroy(value ListFiatCurrenciesResponse) {
	value.Destroy()
}

// Response from listing fiat rates
type ListFiatRatesResponse struct {
	// The list of fiat rates
	Rates []Rate
}

func (r *ListFiatRatesResponse) Destroy() {
	FfiDestroyerSequenceRate{}.Destroy(r.Rates)
}

type FfiConverterListFiatRatesResponse struct{}

var FfiConverterListFiatRatesResponseINSTANCE = FfiConverterListFiatRatesResponse{}

func (c FfiConverterListFiatRatesResponse) Lift(rb RustBufferI) ListFiatRatesResponse {
	return LiftFromRustBuffer[ListFiatRatesResponse](c, rb)
}

func (c FfiConverterListFiatRatesResponse) Read(reader io.Reader) ListFiatRatesResponse {
	return ListFiatRatesResponse{
		FfiConverterSequenceRateINSTANCE.Read(reader),
	}
}

func (c FfiConverterListFiatRatesResponse) Lower(value ListFiatRatesResponse) C.RustBuffer {
	return LowerIntoRustBuffer[ListFiatRatesResponse](c, value)
}

func (c FfiConverterListFiatRatesResponse) Write(writer io.Writer, value ListFiatRatesResponse) {
	FfiConverterSequenceRateINSTANCE.Write(writer, value.Rates)
}

type FfiDestroyerListFiatRatesResponse struct{}

func (_ FfiDestroyerListFiatRatesResponse) Destroy(value ListFiatRatesResponse) {
	value.Destroy()
}

// Request to list payments with optional filters and pagination
type ListPaymentsRequest struct {
	TypeFilter   *[]PaymentType
	StatusFilter *[]PaymentStatus
	AssetFilter  *AssetFilter
	// Only include payments with specific Spark HTLC statuses
	SparkHtlcStatusFilter *[]SparkHtlcStatus
	// Only include payments created after this timestamp (inclusive)
	FromTimestamp *uint64
	// Only include payments created before this timestamp (exclusive)
	ToTimestamp *uint64
	// Number of records to skip
	Offset *uint32
	// Maximum number of records to return
	Limit         *uint32
	SortAscending *bool
}

func (r *ListPaymentsRequest) Destroy() {
	FfiDestroyerOptionalSequencePaymentType{}.Destroy(r.TypeFilter)
	FfiDestroyerOptionalSequencePaymentStatus{}.Destroy(r.StatusFilter)
	FfiDestroyerOptionalAssetFilter{}.Destroy(r.AssetFilter)
	FfiDestroyerOptionalSequenceSparkHtlcStatus{}.Destroy(r.SparkHtlcStatusFilter)
	FfiDestroyerOptionalUint64{}.Destroy(r.FromTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.ToTimestamp)
	FfiDestroyerOptionalUint32{}.Destroy(r.Offset)
	FfiDestroyerOptionalUint32{}.Destroy(r.Limit)
	FfiDestroyerOptionalBool{}.Destroy(r.SortAscending)
}

type FfiConverterListPaymentsRequest struct{}

var FfiConverterListPaymentsRequestINSTANCE = FfiConverterListPaymentsRequest{}

func (c FfiConverterListPaymentsRequest) Lift(rb RustBufferI) ListPaymentsRequest {
	return LiftFromRustBuffer[ListPaymentsRequest](c, rb)
}

func (c FfiConverterListPaymentsRequest) Read(reader io.Reader) ListPaymentsRequest {
	return ListPaymentsRequest{
		FfiConverterOptionalSequencePaymentTypeINSTANCE.Read(reader),
		FfiConverterOptionalSequencePaymentStatusINSTANCE.Read(reader),
		FfiConverterOptionalAssetFilterINSTANCE.Read(reader),
		FfiConverterOptionalSequenceSparkHtlcStatusINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterListPaymentsRequest) Lower(value ListPaymentsRequest) C.RustBuffer {
	return LowerIntoRustBuffer[ListPaymentsRequest](c, value)
}

func (c FfiConverterListPaymentsRequest) Write(writer io.Writer, value ListPaymentsRequest) {
	FfiConverterOptionalSequencePaymentTypeINSTANCE.Write(writer, value.TypeFilter)
	FfiConverterOptionalSequencePaymentStatusINSTANCE.Write(writer, value.StatusFilter)
	FfiConverterOptionalAssetFilterINSTANCE.Write(writer, value.AssetFilter)
	FfiConverterOptionalSequenceSparkHtlcStatusINSTANCE.Write(writer, value.SparkHtlcStatusFilter)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.FromTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.ToTimestamp)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.Offset)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.Limit)
	FfiConverterOptionalBoolINSTANCE.Write(writer, value.SortAscending)
}

type FfiDestroyerListPaymentsRequest struct{}

func (_ FfiDestroyerListPaymentsRequest) Destroy(value ListPaymentsRequest) {
	value.Destroy()
}

// Response from listing payments
type ListPaymentsResponse struct {
	// The list of payments
	Payments []Payment
}

func (r *ListPaymentsResponse) Destroy() {
	FfiDestroyerSequencePayment{}.Destroy(r.Payments)
}

type FfiConverterListPaymentsResponse struct{}

var FfiConverterListPaymentsResponseINSTANCE = FfiConverterListPaymentsResponse{}

func (c FfiConverterListPaymentsResponse) Lift(rb RustBufferI) ListPaymentsResponse {
	return LiftFromRustBuffer[ListPaymentsResponse](c, rb)
}

func (c FfiConverterListPaymentsResponse) Read(reader io.Reader) ListPaymentsResponse {
	return ListPaymentsResponse{
		FfiConverterSequencePaymentINSTANCE.Read(reader),
	}
}

func (c FfiConverterListPaymentsResponse) Lower(value ListPaymentsResponse) C.RustBuffer {
	return LowerIntoRustBuffer[ListPaymentsResponse](c, value)
}

func (c FfiConverterListPaymentsResponse) Write(writer io.Writer, value ListPaymentsResponse) {
	FfiConverterSequencePaymentINSTANCE.Write(writer, value.Payments)
}

type FfiDestroyerListPaymentsResponse struct{}

func (_ FfiDestroyerListPaymentsResponse) Destroy(value ListPaymentsResponse) {
	value.Destroy()
}

type ListUnclaimedDepositsRequest struct {
}

func (r *ListUnclaimedDepositsRequest) Destroy() {
}

type FfiConverterListUnclaimedDepositsRequest struct{}

var FfiConverterListUnclaimedDepositsRequestINSTANCE = FfiConverterListUnclaimedDepositsRequest{}

func (c FfiConverterListUnclaimedDepositsRequest) Lift(rb RustBufferI) ListUnclaimedDepositsRequest {
	return LiftFromRustBuffer[ListUnclaimedDepositsRequest](c, rb)
}

func (c FfiConverterListUnclaimedDepositsRequest) Read(reader io.Reader) ListUnclaimedDepositsRequest {
	return ListUnclaimedDepositsRequest{}
}

func (c FfiConverterListUnclaimedDepositsRequest) Lower(value ListUnclaimedDepositsRequest) C.RustBuffer {
	return LowerIntoRustBuffer[ListUnclaimedDepositsRequest](c, value)
}

func (c FfiConverterListUnclaimedDepositsRequest) Write(writer io.Writer, value ListUnclaimedDepositsRequest) {
}

type FfiDestroyerListUnclaimedDepositsRequest struct{}

func (_ FfiDestroyerListUnclaimedDepositsRequest) Destroy(value ListUnclaimedDepositsRequest) {
	value.Destroy()
}

type ListUnclaimedDepositsResponse struct {
	Deposits []DepositInfo
}

func (r *ListUnclaimedDepositsResponse) Destroy() {
	FfiDestroyerSequenceDepositInfo{}.Destroy(r.Deposits)
}

type FfiConverterListUnclaimedDepositsResponse struct{}

var FfiConverterListUnclaimedDepositsResponseINSTANCE = FfiConverterListUnclaimedDepositsResponse{}

func (c FfiConverterListUnclaimedDepositsResponse) Lift(rb RustBufferI) ListUnclaimedDepositsResponse {
	return LiftFromRustBuffer[ListUnclaimedDepositsResponse](c, rb)
}

func (c FfiConverterListUnclaimedDepositsResponse) Read(reader io.Reader) ListUnclaimedDepositsResponse {
	return ListUnclaimedDepositsResponse{
		FfiConverterSequenceDepositInfoINSTANCE.Read(reader),
	}
}

func (c FfiConverterListUnclaimedDepositsResponse) Lower(value ListUnclaimedDepositsResponse) C.RustBuffer {
	return LowerIntoRustBuffer[ListUnclaimedDepositsResponse](c, value)
}

func (c FfiConverterListUnclaimedDepositsResponse) Write(writer io.Writer, value ListUnclaimedDepositsResponse) {
	FfiConverterSequenceDepositInfoINSTANCE.Write(writer, value.Deposits)
}

type FfiDestroyerListUnclaimedDepositsResponse struct{}

func (_ FfiDestroyerListUnclaimedDepositsResponse) Destroy(value ListUnclaimedDepositsResponse) {
	value.Destroy()
}

// Wrapped in a [`InputType::LnurlAuth`], this is the result of [`parse`](breez_sdk_common::input::parse) when given a LNURL-auth endpoint.
//
// It represents the endpoint's parameters for the LNURL workflow.
//
// See <https://github.com/lnurl/luds/blob/luds/04.md>
type LnurlAuthRequestDetails struct {
	// Hex encoded 32 bytes of challenge
	K1 string
	// When available, one of: register, login, link, auth
	Action *string
	// Indicates the domain of the LNURL-auth service, to be shown to the user when asking for
	// auth confirmation, as per LUD-04 spec.
	Domain string
	// Indicates the URL of the LNURL-auth service, including the query arguments. This will be
	// extended with the signed challenge and the linking key, then called in the second step of the workflow.
	Url string
}

func (r *LnurlAuthRequestDetails) Destroy() {
	FfiDestroyerString{}.Destroy(r.K1)
	FfiDestroyerOptionalString{}.Destroy(r.Action)
	FfiDestroyerString{}.Destroy(r.Domain)
	FfiDestroyerString{}.Destroy(r.Url)
}

type FfiConverterLnurlAuthRequestDetails struct{}

var FfiConverterLnurlAuthRequestDetailsINSTANCE = FfiConverterLnurlAuthRequestDetails{}

func (c FfiConverterLnurlAuthRequestDetails) Lift(rb RustBufferI) LnurlAuthRequestDetails {
	return LiftFromRustBuffer[LnurlAuthRequestDetails](c, rb)
}

func (c FfiConverterLnurlAuthRequestDetails) Read(reader io.Reader) LnurlAuthRequestDetails {
	return LnurlAuthRequestDetails{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterLnurlAuthRequestDetails) Lower(value LnurlAuthRequestDetails) C.RustBuffer {
	return LowerIntoRustBuffer[LnurlAuthRequestDetails](c, value)
}

func (c FfiConverterLnurlAuthRequestDetails) Write(writer io.Writer, value LnurlAuthRequestDetails) {
	FfiConverterStringINSTANCE.Write(writer, value.K1)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Action)
	FfiConverterStringINSTANCE.Write(writer, value.Domain)
	FfiConverterStringINSTANCE.Write(writer, value.Url)
}

type FfiDestroyerLnurlAuthRequestDetails struct{}

func (_ FfiDestroyerLnurlAuthRequestDetails) Destroy(value LnurlAuthRequestDetails) {
	value.Destroy()
}

// Represents the payment LNURL info
type LnurlPayInfo struct {
	LnAddress              *string
	Comment                *string
	Domain                 *string
	Metadata               *string
	ProcessedSuccessAction *SuccessActionProcessed
	RawSuccessAction       *SuccessAction
}

func (r *LnurlPayInfo) Destroy() {
	FfiDestroyerOptionalString{}.Destroy(r.LnAddress)
	FfiDestroyerOptionalString{}.Destroy(r.Comment)
	FfiDestroyerOptionalString{}.Destroy(r.Domain)
	FfiDestroyerOptionalString{}.Destroy(r.Metadata)
	FfiDestroyerOptionalSuccessActionProcessed{}.Destroy(r.ProcessedSuccessAction)
	FfiDestroyerOptionalSuccessAction{}.Destroy(r.RawSuccessAction)
}

type FfiConverterLnurlPayInfo struct{}

var FfiConverterLnurlPayInfoINSTANCE = FfiConverterLnurlPayInfo{}

func (c FfiConverterLnurlPayInfo) Lift(rb RustBufferI) LnurlPayInfo {
	return LiftFromRustBuffer[LnurlPayInfo](c, rb)
}

func (c FfiConverterLnurlPayInfo) Read(reader io.Reader) LnurlPayInfo {
	return LnurlPayInfo{
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalSuccessActionProcessedINSTANCE.Read(reader),
		FfiConverterOptionalSuccessActionINSTANCE.Read(reader),
	}
}

func (c FfiConverterLnurlPayInfo) Lower(value LnurlPayInfo) C.RustBuffer {
	return LowerIntoRustBuffer[LnurlPayInfo](c, value)
}

func (c FfiConverterLnurlPayInfo) Write(writer io.Writer, value LnurlPayInfo) {
	FfiConverterOptionalStringINSTANCE.Write(writer, value.LnAddress)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Comment)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Domain)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Metadata)
	FfiConverterOptionalSuccessActionProcessedINSTANCE.Write(writer, value.ProcessedSuccessAction)
	FfiConverterOptionalSuccessActionINSTANCE.Write(writer, value.RawSuccessAction)
}

type FfiDestroyerLnurlPayInfo struct{}

func (_ FfiDestroyerLnurlPayInfo) Destroy(value LnurlPayInfo) {
	value.Destroy()
}

type LnurlPayRequest struct {
	PrepareResponse PrepareLnurlPayResponse
	// If set, providing the same idempotency key for multiple requests will ensure that only one
	// payment is made. If an idempotency key is re-used, the same payment will be returned.
	// The idempotency key must be a valid UUID.
	IdempotencyKey *string
}

func (r *LnurlPayRequest) Destroy() {
	FfiDestroyerPrepareLnurlPayResponse{}.Destroy(r.PrepareResponse)
	FfiDestroyerOptionalString{}.Destroy(r.IdempotencyKey)
}

type FfiConverterLnurlPayRequest struct{}

var FfiConverterLnurlPayRequestINSTANCE = FfiConverterLnurlPayRequest{}

func (c FfiConverterLnurlPayRequest) Lift(rb RustBufferI) LnurlPayRequest {
	return LiftFromRustBuffer[LnurlPayRequest](c, rb)
}

func (c FfiConverterLnurlPayRequest) Read(reader io.Reader) LnurlPayRequest {
	return LnurlPayRequest{
		FfiConverterPrepareLnurlPayResponseINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterLnurlPayRequest) Lower(value LnurlPayRequest) C.RustBuffer {
	return LowerIntoRustBuffer[LnurlPayRequest](c, value)
}

func (c FfiConverterLnurlPayRequest) Write(writer io.Writer, value LnurlPayRequest) {
	FfiConverterPrepareLnurlPayResponseINSTANCE.Write(writer, value.PrepareResponse)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.IdempotencyKey)
}

type FfiDestroyerLnurlPayRequest struct{}

func (_ FfiDestroyerLnurlPayRequest) Destroy(value LnurlPayRequest) {
	value.Destroy()
}

type LnurlPayRequestDetails struct {
	Callback string
	// The minimum amount, in millisats, that this LNURL-pay endpoint accepts
	MinSendable uint64
	// The maximum amount, in millisats, that this LNURL-pay endpoint accepts
	MaxSendable uint64
	// As per LUD-06, `metadata` is a raw string (e.g. a json representation of the inner map).
	// Use `metadata_vec()` to get the parsed items.
	MetadataStr string
	// The comment length accepted by this endpoint
	//
	// See <https://github.com/lnurl/luds/blob/luds/12.md>
	CommentAllowed uint16
	// Indicates the domain of the LNURL-pay service, to be shown to the user when asking for
	// payment input, as per LUD-06 spec.
	//
	// Note: this is not the domain of the callback, but the domain of the LNURL-pay endpoint.
	Domain string
	Url    string
	// Optional lightning address if that was used to resolve the lnurl.
	Address *string
	// Value indicating whether the recipient supports Nostr Zaps through NIP-57.
	//
	// See <https://github.com/nostr-protocol/nips/blob/master/57.md>
	AllowsNostr *bool
	// Optional recipient's lnurl provider's Nostr pubkey for NIP-57. If it exists it should be a
	// valid BIP 340 public key in hex.
	//
	// See <https://github.com/nostr-protocol/nips/blob/master/57.md>
	// See <https://github.com/bitcoin/bips/blob/master/bip-0340.mediawiki>
	NostrPubkey *string
}

func (r *LnurlPayRequestDetails) Destroy() {
	FfiDestroyerString{}.Destroy(r.Callback)
	FfiDestroyerUint64{}.Destroy(r.MinSendable)
	FfiDestroyerUint64{}.Destroy(r.MaxSendable)
	FfiDestroyerString{}.Destroy(r.MetadataStr)
	FfiDestroyerUint16{}.Destroy(r.CommentAllowed)
	FfiDestroyerString{}.Destroy(r.Domain)
	FfiDestroyerString{}.Destroy(r.Url)
	FfiDestroyerOptionalString{}.Destroy(r.Address)
	FfiDestroyerOptionalBool{}.Destroy(r.AllowsNostr)
	FfiDestroyerOptionalString{}.Destroy(r.NostrPubkey)
}

type FfiConverterLnurlPayRequestDetails struct{}

var FfiConverterLnurlPayRequestDetailsINSTANCE = FfiConverterLnurlPayRequestDetails{}

func (c FfiConverterLnurlPayRequestDetails) Lift(rb RustBufferI) LnurlPayRequestDetails {
	return LiftFromRustBuffer[LnurlPayRequestDetails](c, rb)
}

func (c FfiConverterLnurlPayRequestDetails) Read(reader io.Reader) LnurlPayRequestDetails {
	return LnurlPayRequestDetails{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterUint16INSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalBoolINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterLnurlPayRequestDetails) Lower(value LnurlPayRequestDetails) C.RustBuffer {
	return LowerIntoRustBuffer[LnurlPayRequestDetails](c, value)
}

func (c FfiConverterLnurlPayRequestDetails) Write(writer io.Writer, value LnurlPayRequestDetails) {
	FfiConverterStringINSTANCE.Write(writer, value.Callback)
	FfiConverterUint64INSTANCE.Write(writer, value.MinSendable)
	FfiConverterUint64INSTANCE.Write(writer, value.MaxSendable)
	FfiConverterStringINSTANCE.Write(writer, value.MetadataStr)
	FfiConverterUint16INSTANCE.Write(writer, value.CommentAllowed)
	FfiConverterStringINSTANCE.Write(writer, value.Domain)
	FfiConverterStringINSTANCE.Write(writer, value.Url)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Address)
	FfiConverterOptionalBoolINSTANCE.Write(writer, value.AllowsNostr)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.NostrPubkey)
}

type FfiDestroyerLnurlPayRequestDetails struct{}

func (_ FfiDestroyerLnurlPayRequestDetails) Destroy(value LnurlPayRequestDetails) {
	value.Destroy()
}

type LnurlPayResponse struct {
	Payment       Payment
	SuccessAction *SuccessActionProcessed
}

func (r *LnurlPayResponse) Destroy() {
	FfiDestroyerPayment{}.Destroy(r.Payment)
	FfiDestroyerOptionalSuccessActionProcessed{}.Destroy(r.SuccessAction)
}

type FfiConverterLnurlPayResponse struct{}

var FfiConverterLnurlPayResponseINSTANCE = FfiConverterLnurlPayResponse{}

func (c FfiConverterLnurlPayResponse) Lift(rb RustBufferI) LnurlPayResponse {
	return LiftFromRustBuffer[LnurlPayResponse](c, rb)
}

func (c FfiConverterLnurlPayResponse) Read(reader io.Reader) LnurlPayResponse {
	return LnurlPayResponse{
		FfiConverterPaymentINSTANCE.Read(reader),
		FfiConverterOptionalSuccessActionProcessedINSTANCE.Read(reader),
	}
}

func (c FfiConverterLnurlPayResponse) Lower(value LnurlPayResponse) C.RustBuffer {
	return LowerIntoRustBuffer[LnurlPayResponse](c, value)
}

func (c FfiConverterLnurlPayResponse) Write(writer io.Writer, value LnurlPayResponse) {
	FfiConverterPaymentINSTANCE.Write(writer, value.Payment)
	FfiConverterOptionalSuccessActionProcessedINSTANCE.Write(writer, value.SuccessAction)
}

type FfiDestroyerLnurlPayResponse struct{}

func (_ FfiDestroyerLnurlPayResponse) Destroy(value LnurlPayResponse) {
	value.Destroy()
}

type LnurlReceiveMetadata struct {
	NostrZapRequest *string
	NostrZapReceipt *string
	SenderComment   *string
}

func (r *LnurlReceiveMetadata) Destroy() {
	FfiDestroyerOptionalString{}.Destroy(r.NostrZapRequest)
	FfiDestroyerOptionalString{}.Destroy(r.NostrZapReceipt)
	FfiDestroyerOptionalString{}.Destroy(r.SenderComment)
}

type FfiConverterLnurlReceiveMetadata struct{}

var FfiConverterLnurlReceiveMetadataINSTANCE = FfiConverterLnurlReceiveMetadata{}

func (c FfiConverterLnurlReceiveMetadata) Lift(rb RustBufferI) LnurlReceiveMetadata {
	return LiftFromRustBuffer[LnurlReceiveMetadata](c, rb)
}

func (c FfiConverterLnurlReceiveMetadata) Read(reader io.Reader) LnurlReceiveMetadata {
	return LnurlReceiveMetadata{
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterLnurlReceiveMetadata) Lower(value LnurlReceiveMetadata) C.RustBuffer {
	return LowerIntoRustBuffer[LnurlReceiveMetadata](c, value)
}

func (c FfiConverterLnurlR