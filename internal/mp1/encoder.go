package mp1

import (
	"fmt"
	"io"

	"github.com/daregod/mp1-enc/internal/wav"
)

// Encode writes an MPEG-1 Layer I stream for the given WAV data.
func Encode(w io.Writer, in *wav.File, cfg Config) (int, error) {
	if cfg.BitrateIndex < 1 || cfg.BitrateIndex > 14 {
		return 0, fmt.Errorf("bitrate index %d out of range 1..14", cfg.BitrateIndex)
	}
	rateIndex, ok := SampleRateIndex(in.SampleRate)
	if !ok {
		return 0, fmt.Errorf("unsupported sample rate %d (need 32000, 44100 or 48000)", in.SampleRate)
	}

	channels := in.Channels
	if channels < 1 || len(in.PCM)%channels != 0 {
		return 0, fmt.Errorf("PCM length %d is not divisible by %d channels", len(in.PCM), channels)
	}
	frameSamples := 384 * channels
	frames := (len(in.PCM) + frameSamples - 1) / frameSamples
	if frames == 0 {
		return 0, fmt.Errorf("no audio frames in input")
	}

	bitrate := Layer1Bitrates[cfg.BitrateIndex]
	slotFloat := round(12.0*1000*float64(bitrate)/float64(in.SampleRate), 3)
	slotSize := int(slotFloat)
	slotFrac := slotFloat - float64(slotSize)
	padding := 0.0

	mode := 0 // stereo
	if channels == 1 {
		mode = 3
	}

	poly := NewPolyphase(channels)
	psycho := &Psycho{}

	for frame := 0; frame < frames; frame++ {
		padding += slotFrac
		padBit := 0
		if padding > 1 {
			padBit = 1
			padding -= 1
		}
		maxBits := (slotSize + padBit) * 32

		start := frame * frameSamples
		end := start + frameSamples
		var samples []int16
		if end > len(in.PCM) {
			samples = make([]int16, frameSamples)
			copy(samples, in.PCM[start:])
		} else {
			samples = in.PCM[start:end]
		}
		poly.PushFrame(samples)
		poly.Filter()

		smr := psycho.Smr(poly.PCM(), in.SampleRate)
		scale := ScaleFactors(poly.Subband())

		bw := NewBitWriter(maxBits)
		writeHeader(bw, cfg, rateIndex, padBit, mode)

		totalBits := 32 + 4*32*channels
		if cfg.CRC {
			totalBits += 16
		}
		alloc := make([][]int, channels)
		for ch := range alloc {
			alloc[ch] = make([]int, 32)
		}
		Allocate(smr, maxBits, &totalBits, alloc)

		if cfg.CRC {
			crc := 0x0000FFFF
			crc16(4, cfg.BitrateIndex, &crc)
			crc16(2, rateIndex, &crc)
			crc16(1, padBit, &crc)
			crc16(1, 0, &crc) // private bit
			crc16(2, mode, &crc)
			crc16(2, 0, &crc) // mode extension
			if cfg.Copyright {
				crc16(1, 1, &crc)
			} else {
				crc16(1, 0, &crc)
			}
			if cfg.Original {
				crc16(1, 1, &crc)
			} else {
				crc16(1, 0, &crc)
			}
			crc16(2, cfg.Emphasis, &crc)
			for band := 0; band < 32; band++ {
				for ch := 0; ch < channels; ch++ {
					crc16(4, alloc[ch][band], &crc)
				}
			}
			bw.Put(16, crc)
		}

		quant := Quantize(alloc, poly.Subband(), scale)
		writeBody(bw, alloc, scale, quant, channels)

		if bw.Len() > maxBits {
			return 0, fmt.Errorf("frame %d overflow: %d > %d bits", frame, bw.Len(), maxBits)
		}
		if _, err := w.Write(bw.Bytes()); err != nil {
			return 0, err
		}
	}
	return frames, nil
}

func writeHeader(bw *BitWriter, cfg Config, rateIndex, padding, mode int) {
	bw.PutString("11111111111")
	bw.Put(2, 3) // MPEG-1
	bw.Put(2, 3) // Layer I
	if cfg.CRC {
		bw.Put(1, 0)
	} else {
		bw.Put(1, 1)
	}
	bw.Put(4, cfg.BitrateIndex)
	bw.Put(2, rateIndex)
	bw.Put(1, padding)
	bw.Put(1, 0) // private bit
	bw.Put(2, mode)
	bw.Put(2, 0) // mode extension
	if cfg.Copyright {
		bw.Put(1, 1)
	} else {
		bw.Put(1, 0)
	}
	if cfg.Original {
		bw.Put(1, 1)
	} else {
		bw.Put(1, 0)
	}
	bw.Put(2, cfg.Emphasis)
}

func writeBody(bw *BitWriter, alloc [][]int, scale [][]int, quant [][][]int, channels int) {
	for band := 0; band < 32; band++ {
		for ch := 0; ch < channels; ch++ {
			bw.Put(4, alloc[ch][band])
		}
	}
	for band := 0; band < 32; band++ {
		for ch := 0; ch < channels; ch++ {
			if alloc[ch][band] != 0 {
				bw.Put(6, scale[ch][band])
			}
		}
	}
	for slot := 0; slot < 12; slot++ {
		for band := 0; band < 32; band++ {
			for ch := 0; ch < channels; ch++ {
				k := alloc[ch][band]
				if k != 0 {
					bw.Put(k+1, quant[ch][slot][band])
				}
			}
		}
	}
}
