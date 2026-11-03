package mp1

import "math"

// Config controls the encoder output.
type Config struct {
	BitrateIndex    int  // 1..14, see Layer1Bitrates
	SampleRateIndex int  // 0: 44.1 kHz, 1: 48 kHz, 2: 32 kHz
	CRC             bool // CRC16 error protection
	Copyright       bool
	Original        bool
	Emphasis        int
}

// DefaultConfig matches the Fortran encoder defaults (384 kbit/s, 44.1 kHz).
func DefaultConfig() Config {
	return Config{
		BitrateIndex:    12,
		SampleRateIndex: 0,
		Original:        true,
	}
}

// Layer1Bitrates is ISO Table B.1 for Layer I (kbit/s).
var Layer1Bitrates = [15]int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 414, 448}

// SampleRates is the MPEG-1 sampling frequency table.
var SampleRates = [3]int{44100, 48000, 32000}

// snr is ISO Table C.2 for Layer I (dB).
var snr = [15]float64{0, 7, 16, 25.28, 31.59, 37.75, 43.84, 49.89, 55.93, 61.96, 67.98, 74.01, 80.03, 86.05, 92.01}

var (
	scaleFactor [64]float64
	quantA      [15]float64
	quantB      [15]float64
)

func init() {
	for i := 0; i < 64; i++ {
		scaleFactor[i] = round(2*math.Pow(2, -float64(i)/3), 14)
	}
	for k := 1; k <= 14; k++ {
		tmp := 1 / math.Pow(2, float64(k+1))
		quantA[k] = round(1-tmp, 9)
		quantB[k] = -round(tmp, 9)
	}
}

func round(x float64, digits int) float64 {
	p := math.Pow(10, float64(digits))
	return math.Round(x*p) / p
}

// SampleRateIndex returns the MPEG-1 index for a sampling rate.
func SampleRateIndex(rate int) (int, bool) {
	for i, v := range SampleRates {
		if v == rate {
			return i, true
		}
	}
	return 0, false
}
