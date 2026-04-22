// TLS Handshake for GOST TLS.
package tls

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"math/big"

	"github.com/nativemind/https-vpn/crypto/ru/gost"
)

// Handshake performs the TLS handshake.
type Handshake struct {
	record       *RecordLayer
	isClient     bool
	suite        CipherSuiteInfo
	clientRandom []byte
	serverRandom []byte
	masterSecret []byte

	// Certificate and keys
	cert       []byte
	privateKey *gost.PrivateKey

	// Handshake transcript for Finished verification
	transcript []byte
}

// NewServerHandshake creates a server-side handshake.
func NewServerHandshake(record *RecordLayer, cert []byte, key *gost.PrivateKey) *Handshake {
	return &Handshake{
		record:     record,
		isClient:   false,
		cert:       cert,
		privateKey: key,
	}
}

// NewClientHandshake creates a client-side handshake.
func NewClientHandshake(record *RecordLayer) *Handshake {
	return &Handshake{
		record:   record,
		isClient: true,
	}
}

// DoHandshake performs the full TLS handshake.
func (h *Handshake) DoHandshake() error {
	if h.isClient {
		return h.doClientHandshake()
	}
	return h.doServerHandshake()
}

func (h *Handshake) doServerHandshake() error {
	// 1. Receive ClientHello
	clientHello, err := h.readClientHello()
	if err != nil {
		return err
	}

	// 2. Select cipher suite
	h.suite, err = h.selectCipherSuite(clientHello.cipherSuites)
	if err != nil {
		return err
	}

	// 3. Send ServerHello
	if err := h.sendServerHello(); err != nil {
		return err
	}

	// 4. Send Certificate
	if err := h.sendCertificate(); err != nil {
		return err
	}

	// 5. Send ServerKeyExchange (with ephemeral GOST key)
	ephemeralPriv, err := h.sendServerKeyExchange()
	if err != nil {
		return err
	}

	// 6. Send ServerHelloDone
	if err := h.sendServerHelloDone(); err != nil {
		return err
	}

	// 7. Receive ClientKeyExchange
	preMasterSecret, err := h.readClientKeyExchange(ephemeralPriv)
	if err != nil {
		return err
	}

	// 8. Compute master secret
	h.masterSecret = ComputeMasterSecret(preMasterSecret, h.clientRandom, h.serverRandom, h.suite.Hash)

	// 9. Receive ChangeCipherSpec
	if err := h.readChangeCipherSpec(); err != nil {
		return err
	}

	// 10. Derive keys and set read cipher
	clientKey, serverKey, clientIV, serverIV := DeriveTrafficKeys(
		h.masterSecret, h.clientRandom, h.serverRandom, h.suite)

	readCipher, err := CreateRecordCipher(clientKey, h.suite)
	if err != nil {
		return err
	}
	h.record.SetReadCipher(readCipher, clientIV)

	// 11. Receive Finished
	if err := h.readFinished("client finished"); err != nil {
		return err
	}

	// 12. Send ChangeCipherSpec
	if err := h.record.WriteChangeCipherSpec(); err != nil {
		return err
	}

	// 13. Set write cipher
	writeCipher, err := CreateRecordCipher(serverKey, h.suite)
	if err != nil {
		return err
	}
	h.record.SetWriteCipher(writeCipher, serverIV)

	// 14. Send Finished
	if err := h.sendFinished("server finished"); err != nil {
		return err
	}

	return nil
}

func (h *Handshake) doClientHandshake() error {
	// Client handshake - simplified
	// 1. Send ClientHello
	if err := h.sendClientHello(); err != nil {
		return err
	}

	// 2-6. Receive server messages
	// ... (would implement full client flow)

	return errors.New("client handshake not fully implemented")
}

// ClientHello parsed data
type clientHelloData struct {
	version      uint16
	random       []byte
	sessionID    []byte
	cipherSuites []uint16
	extensions   map[uint16][]byte
}

func (h *Handshake) readClientHello() (*clientHelloData, error) {
	recordType, data, err := h.record.ReadRecord()
	if err != nil {
		return nil, err
	}
	if recordType != RecordTypeHandshake || len(data) < 4 {
		return nil, errors.New("expected handshake record")
	}
	if data[0] != HandshakeTypeClientHello {
		return nil, errors.New("expected ClientHello")
	}

	// Add to transcript
	h.transcript = append(h.transcript, data...)

	// Parse ClientHello
	msgLen := int(data[1])<<16 | int(data[2])<<8 | int(data[3])
	if len(data) < 4+msgLen {
		return nil, errors.New("ClientHello too short")
	}

	body := data[4:]
	ch := &clientHelloData{
		extensions: make(map[uint16][]byte),
	}

	// Version (2 bytes)
	ch.version = binary.BigEndian.Uint16(body[0:2])
	body = body[2:]

	// Random (32 bytes)
	ch.random = make([]byte, 32)
	copy(ch.random, body[:32])
	h.clientRandom = ch.random
	body = body[32:]

	// Session ID
	sidLen := int(body[0])
	body = body[1:]
	ch.sessionID = body[:sidLen]
	body = body[sidLen:]

	// Cipher suites
	csLen := int(binary.BigEndian.Uint16(body[0:2]))
	body = body[2:]
	for i := 0; i < csLen; i += 2 {
		cs := binary.BigEndian.Uint16(body[i : i+2])
		ch.cipherSuites = append(ch.cipherSuites, cs)
	}
	body = body[csLen:]

	// Compression methods (skip)
	compLen := int(body[0])
	body = body[1+compLen:]

	// Extensions (if present)
	if len(body) >= 2 {
		extLen := int(binary.BigEndian.Uint16(body[0:2]))
		body = body[2:]
		for len(body) >= 4 && extLen > 0 {
			extType := binary.BigEndian.Uint16(body[0:2])
			extDataLen := int(binary.BigEndian.Uint16(body[2:4]))
			body = body[4:]
			if len(body) >= extDataLen {
				ch.extensions[extType] = body[:extDataLen]
				body = body[extDataLen:]
			}
			extLen -= 4 + extDataLen
		}
	}

	return ch, nil
}

func (h *Handshake) selectCipherSuite(offered []uint16) (CipherSuiteInfo, error) {
	supported := SupportedCipherSuites()
	for _, s := range supported {
		for _, o := range offered {
			if s == o {
				return CipherSuites[s], nil
			}
		}
	}
	return CipherSuiteInfo{}, errors.New("no common cipher suite")
}

func (h *Handshake) sendServerHello() error {
	// Generate server random
	h.serverRandom = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, h.serverRandom); err != nil {
		return err
	}

	var buf bytes.Buffer

	// Version
	binary.Write(&buf, binary.BigEndian, VersionTLS12)

	// Random
	buf.Write(h.serverRandom)

	// Session ID (empty for now)
	buf.WriteByte(0)

	// Cipher suite
	binary.Write(&buf, binary.BigEndian, h.suite.ID)

	// Compression method (null)
	buf.WriteByte(0)

	// Extensions (empty for now)
	binary.Write(&buf, binary.BigEndian, uint16(0))

	msg := buf.Bytes()
	h.transcript = append(h.transcript, HandshakeTypeServerHello)
	h.transcript = append(h.transcript, byte(len(msg)>>16), byte(len(msg)>>8), byte(len(msg)))
	h.transcript = append(h.transcript, msg...)

	return h.record.WriteHandshake(HandshakeTypeServerHello, msg)
}

func (h *Handshake) sendCertificate() error {
	// Certificate message: length (3) + cert_list
	// cert_list: cert_length (3) + cert_data
	var buf bytes.Buffer

	// Total length placeholder
	certListLen := 3 + len(h.cert)
	buf.WriteByte(byte(certListLen >> 16))
	buf.WriteByte(byte(certListLen >> 8))
	buf.WriteByte(byte(certListLen))

	// Certificate entry
	buf.WriteByte(byte(len(h.cert) >> 16))
	buf.WriteByte(byte(len(h.cert) >> 8))
	buf.WriteByte(byte(len(h.cert)))
	buf.Write(h.cert)

	msg := buf.Bytes()
	h.transcript = append(h.transcript, HandshakeTypeCertificate)
	h.transcript = append(h.transcript, byte(len(msg)>>16), byte(len(msg)>>8), byte(len(msg)))
	h.transcript = append(h.transcript, msg...)

	return h.record.WriteHandshake(HandshakeTypeCertificate, msg)
}

func (h *Handshake) sendServerKeyExchange() (*gost.PrivateKey, error) {
	// Generate ephemeral key pair
	var curve *gost.Curve
	if h.suite.Hash == 512 {
		curve = gost.TC26512A()
	} else {
		curve = gost.TC26256A()
	}

	ephemeralKey, err := gost.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	// Build ServerKeyExchange message
	var buf bytes.Buffer

	// Curve type (named_curve = 3)
	buf.WriteByte(3)

	// Named curve
	if h.suite.Hash == 512 {
		binary.Write(&buf, binary.BigEndian, CurveGOSTR34102012_512_A)
	} else {
		binary.Write(&buf, binary.BigEndian, CurveGOSTR34102012_256_A)
	}

	// Public key (point format: 04 || X || Y)
	byteLen := (curve.BitSize + 7) / 8
	pubKeyLen := 1 + 2*byteLen
	buf.WriteByte(byte(pubKeyLen))
	buf.WriteByte(0x04) // Uncompressed point

	xBytes := ephemeralKey.X.Bytes()
	yBytes := ephemeralKey.Y.Bytes()
	buf.Write(make([]byte, byteLen-len(xBytes)))
	buf.Write(xBytes)
	buf.Write(make([]byte, byteLen-len(yBytes)))
	buf.Write(yBytes)

	// Sign with server's private key
	toSign := append(h.clientRandom, h.serverRandom...)
	toSign = append(toSign, buf.Bytes()...)

	// Hash and sign
	var hash []byte
	if h.suite.Hash == 512 {
		hasher := gost.NewStreebog512()
		hasher.Write(toSign)
		hash = hasher.Sum(nil)
	} else {
		hasher := gost.NewStreebog256()
		hasher.Write(toSign)
		hash = hasher.Sum(nil)
	}

	sig, err := h.privateKey.Sign(rand.Reader, hash, nil)
	if err != nil {
		return nil, err
	}

	// Signature algorithm
	if h.suite.Hash == 512 {
		binary.Write(&buf, binary.BigEndian, SignatureGOSTR34102012_512)
	} else {
		binary.Write(&buf, binary.BigEndian, SignatureGOSTR34102012_256)
	}

	// Signature length and data
	binary.Write(&buf, binary.BigEndian, uint16(len(sig)))
	buf.Write(sig)

	msg := buf.Bytes()
	h.transcript = append(h.transcript, HandshakeTypeServerKeyExchange)
	h.transcript = append(h.transcript, byte(len(msg)>>16), byte(len(msg)>>8), byte(len(msg)))
	h.transcript = append(h.transcript, msg...)

	if err := h.record.WriteHandshake(HandshakeTypeServerKeyExchange, msg); err != nil {
		return nil, err
	}

	return ephemeralKey, nil
}

func (h *Handshake) sendServerHelloDone() error {
	msg := []byte{}
	h.transcript = append(h.transcript, HandshakeTypeServerHelloDone)
	h.transcript = append(h.transcript, 0, 0, 0)

	return h.record.WriteHandshake(HandshakeTypeServerHelloDone, msg)
}

func (h *Handshake) readClientKeyExchange(serverEphemeral *gost.PrivateKey) ([]byte, error) {
	recordType, data, err := h.record.ReadRecord()
	if err != nil {
		return nil, err
	}
	if recordType != RecordTypeHandshake || len(data) < 4 {
		return nil, errors.New("expected handshake record")
	}
	if data[0] != HandshakeTypeClientKeyExchange {
		return nil, errors.New("expected ClientKeyExchange")
	}

	h.transcript = append(h.transcript, data...)

	// Parse client's ephemeral public key
	body := data[4:]
	if len(body) < 1 {
		return nil, errors.New("ClientKeyExchange too short")
	}

	pubKeyLen := int(body[0])
	body = body[1:]
	if len(body) < pubKeyLen || pubKeyLen < 1 {
		return nil, errors.New("invalid public key length")
	}

	// Parse point (04 || X || Y)
	if body[0] != 0x04 {
		return nil, errors.New("expected uncompressed point")
	}
	body = body[1:]
	byteLen := (pubKeyLen - 1) / 2

	clientX := new(big.Int).SetBytes(body[:byteLen])
	clientY := new(big.Int).SetBytes(body[byteLen : 2*byteLen])

	// Compute pre-master secret using VKO
	// UKM is typically clientRandom || serverRandom (or first 8 bytes)
	ukm := append(h.clientRandom[:8], h.serverRandom[:8]...)

	return VKO(serverEphemeral.Curve, serverEphemeral.D, clientX, clientY, ukm)
}

func (h *Handshake) readChangeCipherSpec() error {
	recordType, data, err := h.record.ReadRecord()
	if err != nil {
		return err
	}
	if recordType != RecordTypeChangeCipherSpec || len(data) != 1 || data[0] != 1 {
		return errors.New("expected ChangeCipherSpec")
	}
	return nil
}

func (h *Handshake) readFinished(label string) error {
	recordType, data, err := h.record.ReadRecord()
	if err != nil {
		return err
	}
	if recordType != RecordTypeHandshake || len(data) < 4 {
		return errors.New("expected handshake record")
	}
	if data[0] != HandshakeTypeFinished {
		return errors.New("expected Finished")
	}

	// Compute expected verify_data
	var hashData []byte
	if h.suite.Hash == 512 {
		hasher := gost.NewStreebog512()
		hasher.Write(h.transcript)
		hashData = hasher.Sum(nil)
	} else {
		hasher := gost.NewStreebog256()
		hasher.Write(h.transcript)
		hashData = hasher.Sum(nil)
	}

	expected := ComputeVerifyData(h.masterSecret, label, hashData, h.suite.Hash)
	received := data[4:]

	if !bytes.Equal(expected, received) {
		return errors.New("Finished verify_data mismatch")
	}

	h.transcript = append(h.transcript, data...)
	return nil
}

func (h *Handshake) sendFinished(label string) error {
	// Compute verify_data
	var hashData []byte
	if h.suite.Hash == 512 {
		hasher := gost.NewStreebog512()
		hasher.Write(h.transcript)
		hashData = hasher.Sum(nil)
	} else {
		hasher := gost.NewStreebog256()
		hasher.Write(h.transcript)
		hashData = hasher.Sum(nil)
	}

	verifyData := ComputeVerifyData(h.masterSecret, label, hashData, h.suite.Hash)

	h.transcript = append(h.transcript, HandshakeTypeFinished)
	h.transcript = append(h.transcript, byte(len(verifyData)>>16), byte(len(verifyData)>>8), byte(len(verifyData)))
	h.transcript = append(h.transcript, verifyData...)

	return h.record.WriteHandshake(HandshakeTypeFinished, verifyData)
}

func (h *Handshake) sendClientHello() error {
	h.clientRandom = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, h.clientRandom); err != nil {
		return err
	}

	var buf bytes.Buffer

	// Version
	binary.Write(&buf, binary.BigEndian, VersionTLS12)

	// Random
	buf.Write(h.clientRandom)

	// Session ID (empty)
	buf.WriteByte(0)

	// Cipher suites
	suites := SupportedCipherSuites()
	binary.Write(&buf, binary.BigEndian, uint16(len(suites)*2))
	for _, s := range suites {
		binary.Write(&buf, binary.BigEndian, s)
	}

	// Compression methods
	buf.WriteByte(1) // Length
	buf.WriteByte(0) // null compression

	// Extensions
	binary.Write(&buf, binary.BigEndian, uint16(0))

	msg := buf.Bytes()
	h.transcript = append(h.transcript, HandshakeTypeClientHello)
	h.transcript = append(h.transcript, byte(len(msg)>>16), byte(len(msg)>>8), byte(len(msg)))
	h.transcript = append(h.transcript, msg...)

	return h.record.WriteHandshake(HandshakeTypeClientHello, msg)
}

// GetMasterSecret returns the computed master secret after handshake.
func (h *Handshake) GetMasterSecret() []byte {
	return h.masterSecret
}

// GetSuite returns the negotiated cipher suite.
func (h *Handshake) GetSuite() CipherSuiteInfo {
	return h.suite
}
