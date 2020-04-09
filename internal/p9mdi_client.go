package intelimarketclient

//
// #include <stdlib.h>
// #include "p9mdi_client.h"
//
//#cgo LDFLAGS: -L. -lp9mdi_client
//#cgo CFLAGS: -I../../inc
import "C"

import
(
	"unsafe"
)

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

type P9Connection = C.struct_P9MDI_CONNECTION

func P9mdi_connect(server string, port uint16) *P9Connection {
	var cn *C.struct_P9MDI_CONNECTION

	c_server := C.CString(server)
	defer C.free(unsafe.Pointer(c_server))

	C.p9mdi_connect(
		c_server, 
		C.ushort(port), 
		nil,
		nil,
		nil,
		&cn)

	return cn
}

func P9mdi_disconnect(c_connection *P9Connection)  {
	C.p9mdi_disconnect(c_connection)
}
