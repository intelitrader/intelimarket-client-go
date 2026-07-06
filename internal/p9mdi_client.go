package intelimarketclient

/*
#include <stdlib.h>
#include "p9mdi_client.h"

//
// Forward declaration for functions exported from Go
//

void FieldChangeCallback_cgo(int error_code, void* handle, void* cookie, unsigned eventCode, const char* exchange, const char* symbol, const char* key, char* value);
void TradeCallback_cgo(int error_code, void* handle, void* cookie, unsigned eventCode, const char* exchange, const char* symbol, unsigned position);
void BookEntryCallback_cgo(int error_code, void* handle, void* cookie, unsigned eventCode, const char* exchange, const char* symbol, unsigned position);
void DisconnectionCallback_cgo(struct P9MDI_CONNECTION* connection);

void fieldChangeCallback_Go(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, char* key, char* value);
void tradeCallback_Go(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, unsigned position, struct KEY_AND_VALUE* fields);
void bookEntryCallback_Go(int error_code, void* handle, void* cookie, unsigned eventCode, char* exchange, char* symbol, unsigned position, struct KEY_AND_VALUE* fields);
void disconnectionCallback_Go(struct P9MDI_CONNECTION* connection);

#cgo LDFLAGS: -L./bin -lp9mdi_client
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

func LogTrace(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
    c_line := C.CString(line)
	defer C.free(unsafe.Pointer(c_line))
    C.p9mdi_log(c_line)
}

// cookie, eventCode, exchange, symbol, key, value
type PropertyCallback func(interface{}, uint32, string, string, string, string)

// cookie, eventCode, exchange, symbol, position, key, value
type TradeCallback func(interface{}, uint32, string, string, uint32, map[string]string)

// cookie, eventCode, exchange, symbol, position, key, value
type BookCallback func(interface{}, uint32, string, string, uint32, map[string]string)

type DisconnectionCallback func(*P9GoConnection)

type P9GoConnection struct {
	c_connection          *C.struct_P9MDI_CONNECTION
	connectionId          uintptr
	eventCookie           interface{}
	propertyCallback      PropertyCallback
	tradeCallback         TradeCallback
	bookCallback          BookCallback
	disconnectionCallback DisconnectionCallback
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
var g_connectionsByHandle = map[*C.struct_P9MDI_CONNECTION]uintptr{}

func P9mdi_connect(hostname string, port uint16, logpath string, propertyCallback PropertyCallback, tradeCallback TradeCallback, bookCallback BookCallback, eventCookie interface{}) (*P9GoConnection, error) {
	LogTrace("P9mdi_connect: hostname=%v, port=%v", hostname, port)
	var cn *C.struct_P9MDI_CONNECTION

	c_hostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(c_hostname))

	c_logpath := C.CString(logpath)
	defer C.free(unsafe.Pointer(c_logpath))

	err := C.p9mdi_connect(
		c_hostname,
		C.ushort(port),
		nil,
		nil,
		c_logpath,
		&cn)

	if err != 0 {
		return nil, errors.New("connection error")
	}

	g_lastConnectionId++
	p9_connection := P9GoConnection{}
	p9_connection.c_connection = cn
	p9_connection.connectionId = g_lastConnectionId
	p9_connection.eventCookie = eventCookie
	p9_connection.propertyCallback = propertyCallback
	p9_connection.tradeCallback = tradeCallback
	p9_connection.bookCallback = bookCallback

	g_activeConnections[g_lastConnectionId] = p9_connection
	g_connectionsByHandle[cn] = p9_connection.connectionId

	LogTrace("P9mdi_connect: CONNECTED hostname=%v, port=%v, connectionId=%v", hostname, port, p9_connection.connectionId)

	return &p9_connection, nil
}

func P9mdi_disconnect(p9_connection *P9GoConnection) {
	LogTrace("P9mdi_disconnect: connectionId=%v", p9_connection.connectionId)
	if p9_connection.c_connection != nil {
		delete(g_connectionsByHandle, p9_connection.c_connection)
		C.p9mdi_disconnect(p9_connection.c_connection)
		p9_connection.c_connection = nil
	}
}

func P9mdi_set_asynchronous(p9_connection *P9GoConnection, callback DisconnectionCallback) {
	LogTrace("P9mdi_set_asynchronous: connectionId=%v", p9_connection.connectionId)
	p9_connection.disconnectionCallback = callback
	C.p9mdi_set_asynchronous(
		p9_connection.c_connection,
		(C.DisconnectionCallback)(unsafe.Pointer(C.DisconnectionCallback_cgo)))
}

func P9mdi_subscribe_group(p9_connection *P9GoConnection, groupName string, position int32) {
	LogTrace("P9mdi_subscribe_group: connectionId=%v, groupName=%v", p9_connection.connectionId, groupName)

	cGroupName := C.CString(groupName)
    c_position := C.int(position)
	defer C.free(unsafe.Pointer(cGroupName))

	var cookie uintptr = uintptr(p9_connection.connectionId)

	C.p9mdi_set_subscribe_group_callback(
		p9_connection.c_connection,
		(C.FieldChangeCallback)(unsafe.Pointer(C.FieldChangeCallback_cgo)),
		(C.ListChangeCallback)(unsafe.Pointer(C.BookEntryCallback_cgo)),
		(C.ListChangeCallback)(unsafe.Pointer(C.BookEntryCallback_cgo)),
		(C.ListChangeCallback)(unsafe.Pointer(C.TradeCallback_cgo)),
		(unsafe.Pointer)(cookie))

	C.p9mdi_subscribe_group(
		p9_connection.c_connection,
		cGroupName,
		c_position)
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

func P9mdi_subscribe_instrument_trades(p9_connection *P9GoConnection, symbol string, position int32) {
	LogTrace("P9mdi_subscribe_instrument_trades: connectionId=%v, symbol=%v", p9_connection.connectionId, symbol)

	c_exchange := C.CString("bvmf")
	defer C.free(unsafe.Pointer(c_exchange))

	c_symbol := C.CString(symbol)
	defer C.free(unsafe.Pointer(c_symbol))

    c_position := C.uint(position)

	var cookie uintptr = uintptr(p9_connection.connectionId)

	result := C.p9mdi_subscribe_instrument_trades(
		p9_connection.c_connection,
		c_exchange,
		c_symbol,
		c_position,
		(C.ListChangeCallback)(unsafe.Pointer(C.TradeCallback_cgo)),
		(unsafe.Pointer)(cookie))

	if result != 0 {
		println("Error on set instrument properties callback")
	}
}

func P9mdi_subscribe_instrument_order_book(p9_connection *P9GoConnection, symbol string, side int32, orderBookSize int32) {
	LogTrace("P9mdi_subscribe_instrument_order_book: connectionId=%v, symbol=%v", p9_connection.connectionId, symbol)

	c_exchange := C.CString("bvmf")
	defer C.free(unsafe.Pointer(c_exchange))

	c_symbol := C.CString(symbol)
	defer C.free(unsafe.Pointer(c_symbol))

	c_side := C.uint(side)
	c_size := C.uint(orderBookSize)

	var cookie uintptr = uintptr(p9_connection.connectionId)

	result := C.p9mdi_subscribe_instrument_order_book(
		p9_connection.c_connection,
		c_exchange,
		c_symbol,
		c_side,
		c_size,
		(C.ListChangeCallback)(unsafe.Pointer(C.BookEntryCallback_cgo)),
		(unsafe.Pointer)(cookie))

	if result != 0 {
		println("Error on set instrument order book callback")
	}
}

func P9mdi_ping(p9_connection *P9GoConnection, payload string) int {
	c_payload := C.CString(payload)
	defer C.free(unsafe.Pointer(c_payload))
	result := C.p9mdi_ping(p9_connection.c_connection, c_payload)
	LogTrace("P9mdi_ping: connectionId=%v, payload=%v, result=%v", p9_connection.connectionId, payload, result)
    return int(result)
}

func P9mdi_get_last_error(p9_connection *P9GoConnection) int {
	result := C.p9mdi_get_last_error(p9_connection.c_connection)
    return int(result)
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

	/*
	LogTrace("fieldChangeCallback_Go: connectionId=%v, eventCode=%v, exchange=%v, symbol=%v, key=%v, value=%v",
		p9_connection.connectionId,
		eventCode,
		exchange,
		symbol,
		key,
		value)
	*/

	p9_connection.propertyCallback(p9_connection.eventCookie, eventCode, exchange, symbol, key, value)
}


//export tradeCallback_Go
func tradeCallback_Go(
    errorCode int32, 
    handle unsafe.Pointer, 
    cookie unsafe.Pointer, 
    eventCode uint32, 
    c_exchange *C.char,
    c_symbol *C.char,
    position uint32,
    c_fields *C.struct_KEY_AND_VALUE) {

	exchange := C.GoString(c_exchange)
	symbol := C.GoString(c_symbol)
	key := "empty"
	value := "empty"

	connectionId := uintptr(cookie)
	p9_connection := g_activeConnections[connectionId]

	fields := make(map[string]string)

    for ok := c_fields != nil; ok; ok = c_fields != nil {
        key = C.GoString(c_fields.key)
        value = C.GoString(c_fields.value)
        fields[key] = value
        c_fields = C.p9mdi_get_next_key_value_field(c_fields)
    }

	/*
	LogTrace("tradeCallback_Go: connectionId=%v, eventCode=%v, exchange=%v, symbol=%v, position=%v, fields=%v",
		p9_connection.connectionId,
		eventCode,
		exchange,
		symbol,
        position,
		fields)
	*/

	p9_connection.tradeCallback(p9_connection.eventCookie, eventCode, exchange, symbol, position, fields)
}

//export bookEntryCallback_Go
func bookEntryCallback_Go(
	errorCode int32,
	handle unsafe.Pointer,
	cookie unsafe.Pointer,
	eventCode uint32,
	c_exchange *C.char,
	c_symbol *C.char,
	position uint32,
	c_fields *C.struct_KEY_AND_VALUE) {

	exchange := C.GoString(c_exchange)
	symbol := C.GoString(c_symbol)
	key := "empty"
	value := "empty"

	connectionId := uintptr(cookie)
	p9_connection := g_activeConnections[connectionId]

	fields := make(map[string]string)

	for ok := c_fields != nil; ok; ok = c_fields != nil {
		key = C.GoString(c_fields.key)
		value = C.GoString(c_fields.value)
		fields[key] = value
		c_fields = C.p9mdi_get_next_key_value_field(c_fields)
	}

	/*
	LogTrace("bookEntryCallback_Go: connectionId=%v, eventCode=%v, exchange=%v, symbol=%v, position=%v, fields=%v",
		p9_connection.connectionId,
		eventCode,
		exchange,
		symbol,
		position,
		fields)
	*/

	p9_connection.bookCallback(p9_connection.eventCookie, eventCode, exchange, symbol, position, fields)
}

//export disconnectionCallback_Go
func disconnectionCallback_Go(handle *C.struct_P9MDI_CONNECTION) {
	connectionId, ok := g_connectionsByHandle[handle]
	if !ok {
		return
	}
	p9_connection := g_activeConnections[connectionId]

	LogTrace("disconnectionCallback_Go: connectionId=%v", p9_connection.connectionId)

	if p9_connection.disconnectionCallback != nil {
		p9_connection.disconnectionCallback(&p9_connection)
	}

	p9_connection.c_connection = nil
	delete(g_connectionsByHandle, handle)
}
