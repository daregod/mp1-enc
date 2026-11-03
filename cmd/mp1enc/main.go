package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mp1enc/internal/mp1"
	"mp1enc/internal/wav"
)

func main() {
	bitrate := flag.Int("b", 12, "bitrate index 1..14 (Layer I table)")
	crc := flag.Bool("crc", false, "enable CRC16 error protection")
	copyright := flag.Bool("c", false, "set copyright flag")
	original := flag.Bool("o", true, "set original flag")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] input.wav [output.mp1]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}
	input := flag.Arg(0)
	output := flag.Arg(1)
	if output == "" {
		output = strings.TrimSuffix(input, filepath.Ext(input)) + ".mp1"
	}

	in, err := wav.Read(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	cfg := mp1.DefaultConfig()
	cfg.BitrateIndex = *bitrate
	cfg.CRC = *crc
	cfg.Copyright = *copyright
	cfg.Original = *original

	out, err := os.Create(output)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer out.Close()

	frames, err := mp1.Encode(out, in, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("%s -> %s: %d frames, %d kbit/s, %d Hz, %d ch\n",
		input, output, frames, mp1.Layer1Bitrates[cfg.BitrateIndex], in.SampleRate, in.Channels)
}
