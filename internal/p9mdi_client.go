package intelimarketclient

/*
#include <stdlib.h>
#include "p9mdi_client.h"

//
// Forward declaration for functions exported from Go
//

void FieldChangeCallback_cgo(int error_code, void* handle, void* cookie, unsigned eventCode, const char* exchange, const char* symbol, const char* key, char* value);
void fieldChangeCallback_Go(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, char* key, char* value);

#cgo LDFLAGS: -L./bin -lp9mdi_client
#cgo pkg-config: glib-2.0
*/
import "C"

import (
	"log"
	"unsafe"
)

var g_traceLogEnabled bool = false

func LogTrace(format string, args ...interface{}) {
	if !g_traceLogEnabled {
		return
	}

	log.Printf(format, args...)
}

const (
	EventCodeSet          uint = 0x14
	EventCodeInsert       uint = 0x15
	EventCodeDelete       uint = 0x16
	EventCodePushBack     uint = 0x17
	EventCodePushFront    uint = 0x18
	EventCodePopBack      uint = 0x19
	EventCodePopFront     uint = 0x1A
	EventCodeClear        uint = 0x1B
	EventCodeSnapshotEnd  uint = 0x23
	EventCodeSubscription uint = 0x66
)

// cookie, exchange, symbol, key, value
type PropertyCallback func(interface{}, string, string, string, string)

type P9GoConnection struct {
	c_connection     *C.struct_P9MDI_CONNECTION
	connectionId     uintptr
	eventCookie      interface{}
	propertyCallback PropertyCallback
}

//
// Vamos manter esse map porque não podemos passar
// um ponteiro de uma estrutura Go (como o P9GoConnection)
// como cookie para parte C. O Go não permite passar para a parte
// C nenhum ponteiro Go que tenha outro ponteiro dentro
// Dá um panic em runtime "cgo argument has Go pointer to Go pointer"
//
var g_lastConnectionId uintptr = 0
var g_activeConnections = map[uintptr]P9GoConnection{}

func P9mdi_connect(hostname string, port uint16, propertyCallback PropertyCallback, eventCookie interface{}) *P9GoConnection {
	LogTrace("P9mdi_connect: hostname=%v, port=%v", hostname, port)
	var cn *C.struct_P9MDI_CONNECTION

	c_hostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(c_hostname))

	C.p9mdi_connect(
		c_hostname,
		C.ushort(port),
		nil,
		nil,
		nil,
		&cn)

	g_lastConnectionId++
	p9_connection := P9GoConnection{}
	p9_connection.c_connection = cn
	p9_connection.connectionId = g_lastConnectionId
	p9_connection.eventCookie = eventCookie
	p9_connection.propertyCallback = propertyCallback

	g_activeConnections[g_lastConnectionId] = p9_connection

	LogTrace("P9mdi_connect: CONNECTED hostname=%v, port=%v, connectionId=%v", hostname, port, p9_connection.connectionId)

	return &p9_connection
}

func P9mdi_disconnect(p9_connection *P9GoConnection) {
	LogTrace("P9mdi_disconnect: connectionId=%v", p9_connection.connectionId)
	C.p9mdi_disconnect(p9_connection.c_connection)

	p9_connection.c_connection = nil
}

func P9mdi_subscribe_instrument_properties(p9_connection *P9GoConnection, symbol string) {
	LogTrace("P9mdi_subscribe_instrument_properties: connectionId=%v, symbol=%v", p9_connection.connectionId, symbol)

	c_exchange := C.CString("bvmf")
	defer C.free(unsafe.Pointer(c_exchange))

	c_symbol := C.CString(symbol)
	defer C.free(unsafe.Pointer(c_symbol))

	var cookie uintptr = uintptr(p9_connection.connectionId)

	result := C.p9mdi_subscribe_instrument_properties(
		p9_connection.c_connection,
		c_exchange,
		c_symbol,
		C.SnapshotPlusIncremental,
		(C.FieldChangeCallback)(unsafe.Pointer(C.FieldChangeCallback_cgo)),
		(unsafe.Pointer)(cookie))

	if result != 0 {
		println("Error on set instrument properties callback")
	}
}

func P9mdi_dispatch_pending_events(p9_connection *P9GoConnection) {
	LogTrace("P9mdi_dispatch_pending_events: connectionId=%v", p9_connection.connectionId)
	C.p9mdi_dispatch_pending_events(p9_connection.c_connection, 1)
}

//export fieldChangeCallback_Go
func fieldChangeCallback_Go(
	errorCode int32,
	handle unsafe.Pointer,
	cookie unsafe.Pointer,
	eventCode uint32,
	c_exchange *C.char,
	c_symbol *C.char,
	c_key *C.char,
	c_value *C.char) {

	exchange := C.GoString(c_exchange)
	symbol := C.GoString(c_symbol)
	key := C.GoString(c_key)
	value := C.GoString(c_value)

	connectionId := uintptr(cookie)
	p9_connection := g_activeConnections[connectionId]

	LogTrace("fieldChangeCallback_Go: connectionId=%v, eventCode=%v, exchange=%v, symbol=%v, key=%v, value=%v",
		p9_connection.connectionId,
		eventCode,
		exchange,
		symbol,
		key,
		value)

	p9_connection.propertyCallback(p9_connection.eventCookie, exchange, symbol, key, value)
}
