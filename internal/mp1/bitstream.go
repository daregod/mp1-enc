package mp1

// BitWriter writes MSB-first bits into a frame buffer.
type BitWriter struct {
	buf  []byte
	bits int
}

// NewBitWriter creates a buffer for the given number of bits.
func NewBitWriter(bits int) *BitWriter {
	return &BitWriter{buf: make([]byte, (bits+7)/8)}
}

// Put writes the low n bits of v, most significant bit first.
func (w *BitWriter) Put(n, v int) {
	for i := n - 1; i >= 0; i-- {
		if (v>>uint(i))&1 == 1 {
			w.buf[w.bits/8] |= 1 << uint(7-w.bits%8)
		}
		w.bits++
	}
}

// PutString writes '0'/'1' characters.
func (w *BitWriter) PutString(s string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
			w.buf[w.bits/8] |= 1 << uint(7-w.bits%8)
		}
		w.bits++
	}
}

// Bytes returns the frame bytes.
func (w *BitWriter) Bytes() []byte { return w.buf }

// Len returns the number of bits written.
func (w *BitWriter) Len() int { return w.bits }

// crc16 updates the MPEG audio CRC (polynomial 0x8005).
func crc16(n, value int, crc *int) {
	for i := n - 1; i >= 0; i-- {
		bit1 := (value >> uint(i)) & 1
		bit2 := (*crc >> 15) & 1
		*crc = (*crc << 1) & 0xFFFF
		if bit1^bit2 == 1 {
			*crc ^= 0x8005
		}
	}
}
