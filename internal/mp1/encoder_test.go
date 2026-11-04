package mp1

import (
	"bytes"
	"testing"

	"github.com/daregod/mp1-enc/internal/wav"
)

func TestEncodeFrames(t *testing.T) {
	in := &wav.File{
		Channels:   2,
		SampleRate: 44100,
		PCM:        make([]int16, 2*384*3),
	}
	// Simple signal so the encoder exercises real data.
	for i := range in.PCM {
		in.PCM[i] = int16((i % 200) * 50)
	}

	var buf bytes.Buffer
	frames, err := Encode(&buf, in, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if frames != 3 {
		t.Fatalf("frames = %d, want 3", frames)
	}

	out := buf.Bytes()
	if len(out) == 0 {
		t.Fatal("no output")
	}
	// Header: 11 sync bits, MPEG-1 (11), Layer I (11), no CRC (1) → 0xFFFF.
	if out[0] != 0xFF || out[1] != 0xFF {
		t.Fatalf("sync/layer header = %02X %02X", out[0], out[1])
	}
	// bitrate index 12 (1100), 44.1 kHz (00), no padding, stereo, original.
	if out[2] != 0xC0 || out[3] != 0x04 {
		t.Fatalf("header bytes = %02X %02X, want C0 04", out[2], out[3])
	}

	// Frame size: 12 * 1000 * 384 / 44100 = 104.489 slots → 104/105 slots.
	slot := 12.0 * 1000 * 384 / 44100
	expected := 0
	padding := 0.0
	for i := 0; i < frames; i++ {
		padding += slot - float64(int(slot))
		pad := 0
		if padding > 1 {
			pad = 1
			padding -= 1
		}
		expected += (int(slot) + pad) * 4
	}
	if len(out) != expected {
		t.Fatalf("size = %d bytes, want %d", len(out), expected)
	}
}
func TestEncodePartialFrame(t *testing.T) {
	in := &wav.File{Channels: 1, SampleRate: 44100, PCM: make([]int16, 10)}
	var buf bytes.Buffer
	frames, err := Encode(&buf, in, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if frames != 1 {
		t.Fatalf("frames = %d, want 1", frames)
	}
	if buf.Len() == 0 {
		t.Fatal("no output")
	}
}

func TestEncodeErrors(t *testing.T) {
	cases := []struct {
		name string
		in   *wav.File
		cfg  Config
	}{
		{"bad bitrate", &wav.File{Channels: 1, SampleRate: 44100, PCM: []int16{1}}, Config{BitrateIndex: 15}},
		{"bad rate", &wav.File{Channels: 1, SampleRate: 22050, PCM: []int16{1}}, DefaultConfig()},
		{"odd channels", &wav.File{Channels: 2, SampleRate: 44100, PCM: []int16{1}}, DefaultConfig()},
		{"empty", &wav.File{Channels: 1, SampleRate: 44100}, DefaultConfig()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Encode(&bytes.Buffer{}, tc.in, tc.cfg); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
func TestEncodeCRC(t *testing.T) {
	in := &wav.File{Channels: 1, SampleRate: 44100, PCM: make([]int16, 384)}
	cfg := DefaultConfig()
	cfg.CRC = true
	var buf bytes.Buffer
	if _, err := Encode(&buf, in, cfg); err != nil {
		t.Fatal(err)
	}
	out := buf.Bytes()
	if len(out) < 4 {
		t.Fatal("short output")
	}
	// Protection bit is the low bit of the second header byte: 0 with CRC.
	if out[1]&1 != 0 {
		t.Fatalf("CRC protection bit not set: %02X %02X", out[0], out[1])
	}
}
