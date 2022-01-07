package main

// #include <stdint.h>
// void call_fd_callback(void* func_ptr, void* class_ptr, uintptr_t fd);
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
var tcpConns = map[int]*net.TCPConn{}  // store connections to get RawConn for the file descriptor
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
	transport, _ := obfs4.NewObfs4Client(certString, iatMode, tcpDialer{nextID}) 
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
	if connection == nil {
		return -1
	}
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
	if connection == nil {
		return -1
	}
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
	if connection == nil {
		return
	}
	connection.Close()
	delete(conns, listener_id)
}


//export Obfs4_get_fd
func Obfs4_get_fd(listener_id int, function unsafe.Pointer, class_ptr unsafe.Pointer) int {
	if function == nil {
		return -1
	}

	conn := tcpConns[listener_id]

	if conn == nil {
		return -1
	}

	rawConn, error := conn.SyscallConn()

	if error != nil {
		return -1
	}

	error = rawConn.Control(func (fd uintptr) {
		C.call_fd_callback(function, class_ptr, C.uintptr_t(fd)) 
	})

	if error != nil {
		return -1
	}

	return 0
}


// custom dialer to store the tcpConn + get the file descriptor
type tcpDialer struct {
	id int // transport id
}

func (d tcpDialer) Dial(network, addr string) (c net.Conn, err error) {
	tcpAddr, error := net.ResolveTCPAddr(network, addr)

	if error != nil {
		return nil, error
	}

	conn, error := net.DialTCP(network, nil, tcpAddr)

	if error != nil {
		return nil, error
	}

	tcpConns[d.id] = conn

	return conn, nil
}

func main() {}
