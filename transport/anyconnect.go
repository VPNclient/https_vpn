package transport

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// AnyConnectHandler handles AnyConnect/CSTP requests and proxies them to ocserv backend.
type AnyConnectHandler struct {
	BackendAddr string
	Next        http.Handler // Fallback for standard H2 CONNECT
}

const (
	// Cisco ASA 5506-X / ISR 4331 common banners
	CiscoBanner = "Authorized Access Only. All connections are logged."
	CiscoServer = "Cisco AnyConnect VPN Agent"

	// AnyConnectProfileTmpl is a template for the AnyConnect XML Profile
	AnyConnectProfileTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<AnyConnectProfile xmlns="http://schemas.xmlsoap.org/encoding/">
    <ClientInitialization>
        <StrictCertificateTrust>false</StrictCertificateTrust>
    </ClientInitialization>
    <ServerList>
        <HostEntry>
            <HostName>GOST VPN Hospital</HostName>
            <HostAddress>%s</HostAddress>
        </HostEntry>
    </ServerList>
</AnyConnectProfile>`
)

func (h *AnyConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Check for XML profile request
	if r.URL.Path == "/AnyConnectProfile.xml" || r.URL.Path == "/profile.xml" {
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Server", CiscoServer)
		fmt.Fprintf(w, AnyConnectProfileTmpl, r.Host)
		return
	}

	// 2. Detect AnyConnect/CSTP
	isAnyConnect := r.Header.Get("X-CSTP-Version") != "" || 
		strings.Contains(r.UserAgent(), "AnyConnect") ||
		strings.Contains(r.UserAgent(), "OpenConnect") ||
		r.Header.Get("Upgrade") == "anyconnect"

	if !isAnyConnect {
		if h.Next != nil {
			h.Next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// 3. Add Cisco emulation headers to response
	w.Header().Set("X-CSTP-Banner", CiscoBanner)
	w.Header().Set("Server", CiscoServer)

	// 4. Proxy to ocserv backend
	backend, err := net.Dial("tcp", h.BackendAddr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Backend unavailable: %v", err), http.StatusBadGateway)
		return
	}
	defer backend.Close()

	// 5. Handle H2 or Hijacking
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		if r.ProtoMajor == 2 {
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)
			flusher.Flush()
			
			go io.Copy(backend, r.Body)
			io.Copy(w, backend)
			return
		}
		http.Error(w, "Stream hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "Hijack failed", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Forward the current request to the backend
	if err := r.Write(backend); err != nil {
		return
	}

	// Start bidirectional pipe between client and ocserv
	pipe(clientConn, backend)
}
