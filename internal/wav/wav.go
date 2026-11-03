package wav

import (
	"encoding/binary"
	"fmt"
	"os"
)

// File is a decoded 16-bit PCM WAV file.
type File struct {
	Channels      int
	SampleRate    int
	BitsPerSample int
	PCM           []int16
}

// Read parses a RIFF/WAVE file containing 16-bit PCM.
func Read(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, fmt.Errorf("%s: not a RIFF/WAVE file", path)
	}

	f := &File{}
	var pcmData []byte
	offset := 12
	for offset+8 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		start := offset + 8
		if size < 0 || start+size > len(data) {
			break
		}
		switch id {
		case "fmt ":
			if size < 16 {
				return nil, fmt.Errorf("%s: short fmt chunk", path)
			}
			format := binary.LittleEndian.Uint16(data[start : start+2])
			if format != 1 {
				return nil, fmt.Errorf("%s: unsupported WAV format %d (need PCM)", path, format)
			}
			f.Channels = int(binary.LittleEndian.Uint16(data[start+2 : start+4]))
			f.SampleRate = int(binary.LittleEndian.Uint32(data[start+4 : start+8]))
			f.BitsPerSample = int(binary.LittleEndian.Uint16(data[start+14 : start+16]))
		case "data":
			pcmData = data[start : start+size]
		}
		offset = start + size
		if size%2 == 1 {
			offset++
		}
	}

	if f.Channels != 1 && f.Channels != 2 {
		return nil, fmt.Errorf("%s: %d channels (only mono/stereo supported)", path, f.Channels)
	}
	if f.BitsPerSample != 16 {
		return nil, fmt.Errorf("%s: %d bits per sample (only 16 supported)", path, f.BitsPerSample)
	}
	if len(pcmData) == 0 {
		return nil, fmt.Errorf("%s: no data chunk", path)
	}

	f.PCM = make([]int16, len(pcmData)/2)
	for i := range f.PCM {
		f.PCM[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2 : i*2+2]))
	}
	return f, nil
}
