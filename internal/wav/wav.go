package wav

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	formatPCM        = 1
	formatExtensible = 0xFFFE
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
	return parse(data, path)
}

func parse(data []byte, path string) (*File, error) {
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
			return nil, fmt.Errorf("%s: truncated %q chunk", path, id)
		}
		switch id {
		case "fmt ":
			if size < 16 {
				return nil, fmt.Errorf("%s: short fmt chunk", path)
			}
			format := binary.LittleEndian.Uint16(data[start : start+2])
			f.Channels = int(binary.LittleEndian.Uint16(data[start+2 : start+4]))
			f.SampleRate = int(binary.LittleEndian.Uint32(data[start+4 : start+8]))
			byteRate := binary.LittleEndian.Uint32(data[start+8 : start+12])
			blockAlign := int(binary.LittleEndian.Uint16(data[start+12 : start+14]))
			f.BitsPerSample = int(binary.LittleEndian.Uint16(data[start+14 : start+16]))
			switch format {
			case formatPCM:
			case formatExtensible:
				if size < 40 {
					return nil, fmt.Errorf("%s: short WAVE_FORMAT_EXTENSIBLE fmt chunk", path)
				}
				validBits := int(binary.LittleEndian.Uint16(data[start+18 : start+20]))
				if validBits != 0 {
					f.BitsPerSample = validBits
				}
				if data[start+24] != 1 || data[start+25] != 0 ||
					string(data[start+26:start+40]) != "\x00\x00\x00\x00\x10\x00\x80\x00\x00\xaa\x00\x38\x9b\x71" {
					return nil, fmt.Errorf("%s: unsupported WAVE_FORMAT_EXTENSIBLE subformat", path)
				}
			default:
				return nil, fmt.Errorf("%s: unsupported WAV format %d (need PCM)", path, format)
			}
			if f.Channels != 1 && f.Channels != 2 {
				return nil, fmt.Errorf("%s: %d channels (only mono/stereo supported)", path, f.Channels)
			}
			if f.BitsPerSample != 16 {
				return nil, fmt.Errorf("%s: %d bits per sample (only 16 supported)", path, f.BitsPerSample)
			}
			if blockAlign != f.Channels*2 {
				return nil, fmt.Errorf("%s: bad block align %d for %d channels", path, blockAlign, f.Channels)
			}
			if byteRate != uint32(f.SampleRate*blockAlign) {
				return nil, fmt.Errorf("%s: bad byte rate %d", path, byteRate)
			}
		case "data":
			if size%2 != 0 {
				return nil, fmt.Errorf("%s: odd data chunk size %d", path, size)
			}
			pcmData = data[start : start+size]
		}
		offset = start + size
		if size%2 == 1 {
			offset++
		}
	}

	if f.Channels == 0 {
		return nil, fmt.Errorf("%s: no fmt chunk", path)
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
