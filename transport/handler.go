package transport

import (
	"fmt"
	"io"
	"net"
	"net/http"
)

// ConnectHandler processes HTTP CONNECT requests.
// It dials the target and pipes data between client and target.
type ConnectHandler struct{}

// ServeHTTP handles HTTP CONNECT requests.
func (h *ConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Dial the target
	target := r.Host
	if target == "" {
		http.Error(w, "Missing host", http.StatusBadRequest)
		return
	}

	targetConn, err := net.Dial("tcp", target)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to target: %v", err), http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	// Hijack the connection to get raw access
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, buf, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, fmt.Sprintf("Hijack failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established response
	if buf.Reader.Buffered() > 0 {
		// There's buffered data we need to handle
		io.Copy(clientConn, buf.Reader)
	}

	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		return
	}

	// Pipe data between client and target
	pipe(clientConn, targetConn)
}
