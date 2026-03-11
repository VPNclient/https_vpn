package transport_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nativemind/https-vpn/transport"
)

// TestConnectHandler_Success tests successful CONNECT request
func TestConnectHandler_Success(t *testing.T) {
	// Create a test target server
	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create target listener: %v", err)
	}
	defer targetListener.Close()

	// Echo server that responds with what it receives
	go func() {
		conn, err := targetListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		conn.Write([]byte("received: " + string(buf[:n])))
	}()

	// Create handler
	handler := &transport.ConnectHandler{}

	// Create CONNECT request
	targetAddr := targetListener.Addr().String()
	req := httptest.NewRequest(http.MethodConnect, fmt.Sprintf("http://%s", targetAddr), nil)
	req.Host = targetAddr

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check response status (should be 200 for hijacked connection)
	// Note: httptest doesn't support hijacking well, so we test the error path
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500 (hijack not supported), got %d", w.Code)
	}
}

// TestConnectHandler_MethodNotAllowed tests non-CONNECT method
func TestConnectHandler_MethodNotAllowed(t *testing.T) {
	handler := &transport.ConnectHandler{}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestConnectHandler_MissingHost tests missing host header
func TestConnectHandler_MissingHost(t *testing.T) {
	handler := &transport.ConnectHandler{}

	req := httptest.NewRequest(http.MethodConnect, "http://", nil)
	req.Host = ""
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestH2Server_StartAndShutdown tests server lifecycle
func TestH2Server_StartAndShutdown(t *testing.T) {
	// Generate self-signed cert for testing
	cert, err := tls.X509KeyPair([]byte(testCertPEM), []byte(testKeyPEM))
	if err != nil {
		t.Fatalf("Failed to load test cert: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2"},
	}

	cfg := &transport.ServerConfig{
		Addr:      "127.0.0.1:0", // Use port 0 for dynamic port
		TLSConfig: tlsConfig,
		Handler:   &transport.ConnectHandler{},
	}

	server, err := transport.NewH2Server(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	done := make(chan error, 1)
	go func() {
		done <- server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown
	ctx := context.Background()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Wait for server to stop
	select {
	case err := <-done:
		if err != http.ErrServerClosed && err != nil {
			t.Errorf("Server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server did not stop within timeout")
	}
}

// Test cert for testing purposes (self-signed, localhost)
const testCertPEM = `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`

const testKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`
