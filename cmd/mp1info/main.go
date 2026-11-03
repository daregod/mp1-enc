package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
)

var bitrates = [15]int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 414, 448}
var rates = [3]int{44100, 48000, 32000}

func main() {
	pads := flag.Bool("pads", false, "print the padding bit sequence")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: mp1info [-pads] file.mp1")
		os.Exit(2)
	}
	for _, path := range flag.Args() {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		off := 0
		frames := 0
		var padSeq []byte
		for off+4 <= len(data) {
			h := binary.BigEndian.Uint32(data[off:])
			if h>>21 != 0x7FF {
				break
			}
			version := (h >> 19) & 3
			layer := (h >> 17) & 3
			crc := (h >> 16) & 1
			brIndex := (h >> 12) & 0xF
			srIndex := (h >> 10) & 3
			padding := (h >> 9) & 1
			if int(brIndex) >= len(bitrates) || int(srIndex) >= len(rates) {
				break
			}
			slots := 12000.0 * float64(bitrates[brIndex]) / float64(rates[srIndex])
			length := (int(slots) + int(padding)) * 4
			if length <= 0 || off+length > len(data) {
				break
			}
			if frames == 0 {
				fmt.Printf("%s: version=%d layer=%d crc=%d bitrate=%d rate=%d mode=%d\n",
					path, version, layer, crc, bitrates[brIndex], rates[srIndex], (h>>6)&3)
			}
			frames++
			padSeq = append(padSeq, byte('0'+padding))
			off += length
		}
		fmt.Printf("  frames=%d bytes=%d trailer=%d\n", frames, off, len(data)-off)
		if *pads {
			fmt.Printf("  pads=%s\n", padSeq)
		}
	}
}
