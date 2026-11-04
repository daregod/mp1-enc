package wav

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func buildWAV(channels, rate int, samples []int16) []byte {
	data := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(s))
	}
	blockAlign := channels * 2
	byteRate := rate * blockAlign

	body := make([]byte, 0, 36+len(data))
	body = append(body, []byte("fmt ")...)
	fmtChunk := make([]byte, 16)
	binary.LittleEndian.PutUint16(fmtChunk[0:], 1)
	binary.LittleEndian.PutUint16(fmtChunk[2:], uint16(channels))
	binary.LittleEndian.PutUint32(fmtChunk[4:], uint32(rate))
	binary.LittleEndian.PutUint32(fmtChunk[8:], uint32(byteRate))
	binary.LittleEndian.PutUint16(fmtChunk[12:], uint16(blockAlign))
	binary.LittleEndian.PutUint16(fmtChunk[14:], 16)
	body = append(body, byte(len(fmtChunk)), 0, 0, 0)
	body = append(body, fmtChunk...)
	body = append(body, []byte("data")...)
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))
	body = append(body, size...)
	body = append(body, data...)

	out := append([]byte("RIFF"), size...)
	binary.LittleEndian.PutUint32(out[4:], uint32(4+len(body)))
	out = append(out, []byte("WAVE")...)
	return append(out, body...)
}

func TestRead(t *testing.T) {
	samples := []int16{0, 1000, -1000, 32767}
	path := filepath.Join(t.TempDir(), "test.wav")
	if err := os.WriteFile(path, buildWAV(1, 44100, samples), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Channels != 1 || f.SampleRate != 44100 || f.BitsPerSample != 16 {
		t.Fatalf("file = %+v", f)
	}
	if len(f.PCM) != len(samples) {
		t.Fatalf("pcm = %d samples", len(f.PCM))
	}
	for i := range samples {
		if f.PCM[i] != samples[i] {
			t.Fatalf("pcm[%d] = %d, want %d", i, f.PCM[i], samples[i])
		}
	}
}
func buildExtensibleWAV(channels, rate int, samples []int16) []byte {
	data := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(s))
	}
	blockAlign := channels * 2
	byteRate := rate * blockAlign
	fmtChunk := make([]byte, 40)
	binary.LittleEndian.PutUint16(fmtChunk[0:], 0xFFFE)
	binary.LittleEndian.PutUint16(fmtChunk[2:], uint16(channels))
	binary.LittleEndian.PutUint32(fmtChunk[4:], uint32(rate))
	binary.LittleEndian.PutUint32(fmtChunk[8:], uint32(byteRate))
	binary.LittleEndian.PutUint16(fmtChunk[12:], uint16(blockAlign))
	binary.LittleEndian.PutUint16(fmtChunk[14:], 16)
	binary.LittleEndian.PutUint16(fmtChunk[16:], 22)
	binary.LittleEndian.PutUint16(fmtChunk[18:], 16)
	copy(fmtChunk[24:], []byte{1, 0, 0, 0, 0, 0, 0x10, 0, 0x80, 0, 0, 0xaa, 0, 0x38, 0x9b, 0x71})

	body := append([]byte("fmt "), make([]byte, 4)...)
	binary.LittleEndian.PutUint32(body[4:], uint32(len(fmtChunk)))
	body = append(body, fmtChunk...)
	body = append(body, []byte("data")...)
	body = append(body, make([]byte, 4)...)
	binary.LittleEndian.PutUint32(body[len(body)-4:], uint32(len(data)))
	body = append(body, data...)

	out := append([]byte("RIFF"), make([]byte, 4)...)
	binary.LittleEndian.PutUint32(out[4:], uint32(4+len(body)))
	out = append(out, []byte("WAVE")...)
	return append(out, body...)
}

func TestReadExtensible(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ext.wav")
	if err := os.WriteFile(path, buildExtensibleWAV(1, 48000, []int16{1, -2, 3}), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.BitsPerSample != 16 || f.SampleRate != 48000 || len(f.PCM) != 3 {
		t.Fatalf("file = %+v", f)
	}
}

func TestReadBadBlockAlign(t *testing.T) {
	data := buildWAV(1, 44100, []int16{1})
	binary.LittleEndian.PutUint16(data[32:34], 4)
	path := filepath.Join(t.TempDir(), "bad.wav")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path); err == nil {
		t.Fatal("bad block align must fail")
	}
}

func FuzzParse(f *testing.F) {
	f.Add(buildWAV(1, 44100, []int16{1, 2}))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parse(data, "fuzz.wav")
	})
}
