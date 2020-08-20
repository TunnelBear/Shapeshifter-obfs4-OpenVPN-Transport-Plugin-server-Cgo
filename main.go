package main

import "C"
import (
	"net"
	"unsafe"
	"reflect"

	"github.com/TunnelBear/shapeshifter-transports/transports/obfs4/v2"
)

var transports = map[int]*obfs4.Transport{}
var listeners = map[int]net.Listener{}
var conns = map[int]net.Conn{}
var nextID = 0

//export Obfs4_initialize_server
func Obfs4_initialize_server(stateDir *C.char) (listenerKey int) {
	goStateString := C.GoString(stateDir)
	transport, _ := obfs4.NewObfs4Server(goStateString)
	transports[nextID] = transport

	// This is the return value
	if transport != nil {
		listenerKey = nextID
	} else {
		listenerKey = -1
	}

	nextID += 1
	return
}

//export Obfs4_initialize_client
func Obfs4_initialize_client(cert *C.char, iatMode int) (listenerKey int) {
	certString := C.GoString(cert)
	transport, _ := obfs4.NewObfs4Client(certString, iatMode, nil) // TODO: support dialer
	transports[nextID] = transport

	// This is the return value
	if transport != nil {
		listenerKey = nextID
	} else {
		listenerKey = -1
	}

	nextID += 1
	return
}

//export Obfs4_listen
func Obfs4_listen(id int, address_string *C.char) {
	goAddressString := C.GoString(address_string)

	var transport = transports[id]
	var listener = transport.Listen(goAddressString)
	listeners[id] = listener
}

//export Obfs4_dial
func Obfs4_dial(id int, address_string *C.char) int {
	goAddressString := C.GoString(address_string)

	var transport = transports[id]
	var conn, err = transport.Dial(goAddressString)

	if err != nil {
		return -1
	} else {
		conns[id] = conn
		return 0
	}
}

//export Obfs4_accept
func Obfs4_accept(id int) {
	var listener = listeners[id]

	conn, err := listener.Accept()
	if err != nil {
		return
	}

	conns[id] = conn
}

//export Obfs4_write
func Obfs4_write(listener_id int, buffer unsafe.Pointer, buffer_length C.int) int {
	var connection = conns[listener_id]
	var bytesBuffer = C.GoBytes(buffer, buffer_length)
	numberOfBytesWritten, error := connection.Write(bytesBuffer)

	if error != nil {
		return -1
	} else {
		return numberOfBytesWritten
	}
}

//export Obfs4_read
func Obfs4_read(listener_id int, buffer unsafe.Pointer, buffer_length int) int {
	var connection = conns[listener_id]
	header := reflect.SliceHeader{uintptr(buffer), buffer_length, buffer_length}
	bytesBuffer := *(*[]byte)(unsafe.Pointer(&header))

	numberOfBytesRead, error := connection.Read(bytesBuffer)

	if error != nil {
		return -1
	} else {
		return numberOfBytesRead
	}
}

//export Obfs4_close_connection
func Obfs4_close_connection(listener_id int) {

	var connection = conns[listener_id]
	connection.Close()
	delete(conns, listener_id)
}

//export Obfs4_get_file_descriptor
func Obfs4_get_file_descriptor(listener_id int) int {
  var connection = conss[listener_id]
	value := reflect.ValueOf(connection)
	netFD := reflect.Indirect(reflect.Indirect(value).FieldByName("fd"))
	fd := int(netFD.FieldByName("sysfd").Int())
	return fd
}

//export Obfs4_get_file_descriptor2
func Obfs4_get_file_descriptor2(listener_id int) int {
  var connection = conss[listener_id]
	value := reflect.ValueOf(connection)
	netFD := reflect.Indirect(reflect.Indirect(value).FieldByName("fd"))
  pfd := reflect.Indirect(netFD.FieldByName("pfd"))
	fd := int(pfd.FieldByName("Sysfd").Int())
	return fd
}

func main() {}
