// TLS Record Layer for GOST TLS.
package tls

import (
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

const (
	maxPlaintext    = 16384     // 2^14
	maxCiphertext   = 16384 + 256 // With overhead
	recordHeaderLen = 5
)

// RecordLayer handles TLS record encryption/decryption.
type RecordLayer struct {
	conn io.ReadWriter

	// Write state
	writeMu     sync.Mutex
	writeSeq    uint64
	writeCipher cipher.AEAD
	writeIV     []byte

	// Read state
	readMu     sync.Mutex
	readSeq    uint64
	readCipher cipher.AEAD
	readIV     []byte

	// Version for record header
	version uint16
}

// NewRecordLayer creates a new record layer.
func NewRecordLayer(conn io.ReadWriter) *RecordLayer {
	return &RecordLayer{
		conn:    conn,
		version: VersionTLS12,
	}
}

// SetWriteCipher sets the cipher for outgoing records.
func (r *RecordLayer) SetWriteCipher(aead cipher.AEAD, iv []byte) {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()
	r.writeCipher = aead
	r.writeIV = make([]byte, len(iv))
	copy(r.writeIV, iv)
	r.writeSeq = 0
}

// SetReadCipher sets the cipher for incoming records.
func (r *RecordLayer) SetReadCipher(aead cipher.AEAD, iv []byte) {
	r.readMu.Lock()
	defer r.readMu.Unlock()
	r.readCipher = aead
	r.readIV = make([]byte, len(iv))
	copy(r.readIV, iv)
	r.readSeq = 0
}

// WriteRecord writes an encrypted TLS record.
func (r *RecordLayer) WriteRecord(recordType uint8, data []byte) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()

	if len(data) > maxPlaintext {
		return errors.New("tls: record too large")
	}

	var payload []byte
	if r.writeCipher != nil {
		// Encrypt
		nonce := r.makeNonce(r.writeIV, r.writeSeq)
		additionalData := r.makeAdditionalData(recordType, len(data))

		payload = r.writeCipher.Seal(nil, nonce, data, additionalData)
		r.writeSeq++
	} else {
		payload = data
	}

	// Build record header
	header := make([]byte, recordHeaderLen)
	header[0] = recordType
	binary.BigEndian.PutUint16(header[1:3], r.version)
	binary.BigEndian.PutUint16(header[3:5], uint16(len(payload)))

	// Write header + payload
	if _, err := r.conn.Write(header); err != nil {
		return err
	}
	if _, err := r.conn.Write(payload); err != nil {
		return err
	}

	return nil
}

// ReadRecord reads and decrypts a TLS record.
func (r *RecordLayer) ReadRecord() (recordType uint8, data []byte, err error) {
	r.readMu.Lock()
	defer r.readMu.Unlock()

	// Read header
	header := make([]byte, recordHeaderLen)
	if _, err = io.ReadFull(r.conn, header); err != nil {
		return 0, nil, err
	}

	recordType = header[0]
	// version := binary.BigEndian.Uint16(header[1:3])
	length := binary.BigEndian.Uint16(header[3:5])

	if length > maxCiphertext {
		return 0, nil, errors.New("tls: record too large")
	}

	// Read payload
	payload := make([]byte, length)
	if _, err = io.ReadFull(r.conn, payload); err != nil {
		return 0, nil, err
	}

	if r.readCipher != nil {
		// Decrypt
		nonce := r.makeNonce(r.readIV, r.readSeq)
		additionalData := r.makeAdditionalData(recordType, int(length)-r.readCipher.Overhead())

		data, err = r.readCipher.Open(nil, nonce, payload, additionalData)
		if err != nil {
			return 0, nil, err
		}
		r.readSeq++
	} else {
		data = payload
	}

	return recordType, data, nil
}

// makeNonce constructs the nonce from IV and sequence number.
// For GOST TLS: nonce = IV XOR seq (padded to IV length)
func (r *RecordLayer) makeNonce(iv []byte, seq uint64) []byte {
	nonce := make([]byte, len(iv))
	copy(nonce, iv)

	// XOR sequence number into the last 8 bytes
	seqBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(seqBytes, seq)

	start := len(nonce) - 8
	if start < 0 {
		start = 0
		seqBytes = seqBytes[8-len(nonce):]
	}
	for i := 0; i < len(seqBytes) && start+i < len(nonce); i++ {
		nonce[start+i] ^= seqBytes[i]
	}

	return nonce
}

// makeAdditionalData constructs the additional authenticated data.
// For TLS 1.2: seq_num + record_type + version + length
func (r *RecordLayer) makeAdditionalData(recordType uint8, plaintextLen int) []byte {
	ad := make([]byte, 13)
	binary.BigEndian.PutUint64(ad[0:8], r.readSeq) // Use current seq
	ad[8] = recordType
	binary.BigEndian.PutUint16(ad[9:11], r.version)
	binary.BigEndian.PutUint16(ad[11:13], uint16(plaintextLen))
	return ad
}

// WriteHandshake writes a handshake message.
func (r *RecordLayer) WriteHandshake(msgType uint8, data []byte) error {
	// Handshake header: type (1) + length (3)
	msg := make([]byte, 4+len(data))
	msg[0] = msgType
	msg[1] = byte(len(data) >> 16)
	msg[2] = byte(len(data) >> 8)
	msg[3] = byte(len(data))
	copy(msg[4:], data)

	return r.WriteRecord(RecordTypeHandshake, msg)
}

// WriteChangeCipherSpec writes a ChangeCipherSpec message.
func (r *RecordLayer) WriteChangeCipherSpec() error {
	return r.WriteRecord(RecordTypeChangeCipherSpec, []byte{1})
}

// WriteAlert writes an alert message.
func (r *RecordLayer) WriteAlert(level, desc uint8) error {
	return r.WriteRecord(RecordTypeAlert, []byte{level, desc})
}
