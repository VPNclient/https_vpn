# Requirements: https-vpn-openconnect-server-backward-compatibility

> Version: 1.0  
> Status: DRAFT  
> Last Updated: 2026-04-21

## Problem Statement

The current HTTPS VPN server implementation is a minimalist HTTP/2 CONNECT proxy. While this is effective for custom clients, it lacks compatibility with widely used VPN clients such as OpenConnect and Cisco AnyConnect. To increase the utility of the HTTPS VPN server, it should provide backward compatibility with the OpenConnect server (ocserv) protocol, allowing these standard clients to connect and tunnel traffic without modification.

## User Stories

### Primary

**As a** Remote Worker  
**I want** to use my existing OpenConnect client to connect to the HTTPS VPN server  
**So that** I don't have to install new software on my managed device.

### Secondary

**As a** Network Administrator  
**I want** the HTTPS VPN server to be a drop-in replacement for basic ocserv deployments  
**So that** I can benefit from the lightweight codebase and national crypto support while maintaining client compatibility.

## Acceptance Criteria

### Must Have

1. **Protocol Negotiation**: The server must handle the initial HTTPS GET/POST requests used by OpenConnect clients for protocol detection and authentication.
2. **Tunnel Establishment**: The server must support the specific HTTP CONNECT headers and behavior expected by OpenConnect clients to establish the data tunnel.
3. **Backward Compatibility**: Support for older OpenConnect/AnyConnect clients that may rely on legacy headers or specific TLS handshake patterns.
4. **Minimal Footprint**: The implementation must stay within the spirit of the project's "small codebase" goal, avoiding unnecessary complexity.

### Should Have

1. **Basic XML Auth**: Implementation of a minimal XML-based authentication response to satisfy client requirements (even if it's a "permit-all" or simple token check).
2. **MTU Negotiation**: Basic handling of MTU discovery/negotiation headers.

### Won't Have (This Iteration)

1. **DTLS Support**: OpenConnect often uses DTLS for performance, but this project prioritizes "browser-identical" HTTPS traffic (TCP/TLS). DTLS might be added later if needed.
2. **Complex Auth**: Integration with LDAP, RADIUS, or SAML is out of scope for this minimalist implementation.
3. **Advanced Routing**: Complex split-tunneling or dynamic routing protocol support.

## Constraints

- **Architecture**: Must integrate seamlessly with the existing `transport.H2Server` and `ConnectHandler`.
- **Code Size**: Aim to keep the additional logic minimal to maintain the ~700 LOC "certification-ready" advantage.
- **DPI Resistance**: Compatibility features must not introduce unique signatures that make the traffic easily detectable as a VPN.

## Open Questions

- [ ] Which specific versions of OpenConnect/AnyConnect are the primary targets for "backward compatibility"?
- [ ] Is the XML authentication flow strictly required by all clients, or can it be bypassed for a simpler "direct CONNECT" approach?
- [ ] How will IP allocation (CSTP-Address) be handled without a complex stateful IP pool manager?
- [ ] Are there specific legacy headers (e.g., `X-CSTP-*`) that are mandatory for backward compatibility?

## References

- [OpenConnect Protocol Documentation](https://www.infradead.org/openconnect/protocol.html)
- [ocserv - OpenConnect VPN server](https://gitlab.com/openconnect/ocserv)
- [Cisco AnyConnect SSL VPN Protocol (CSTP)](https://tools.ietf.org/html/draft-mavrogiannopoulos-openconnect-03)

---

## Approval

- [ ] Reviewed by: [name]
- [ ] Approved on: [date]
- [ ] Notes: [any conditions or clarifications]
