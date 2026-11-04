package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daregod/mp1-enc/internal/mp1"
	"github.com/daregod/mp1-enc/internal/wav"
)

var version = "0.1.0"

func main() {
	bitrate := flag.Int("b", 12, "bitrate index 1..14 (Layer I table)")
	crc := flag.Bool("crc", false, "enable CRC16 error protection")
	copyright := flag.Bool("c", false, "set copyright flag")
	original := flag.Bool("o", true, "set original flag")
	force := flag.Bool("f", false, "overwrite an existing output file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] input.wav [output.mp1]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("mp1enc", version)
		return
	}
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}
	if *bitrate < 1 || *bitrate > 14 {
		fmt.Fprintln(os.Stderr, "error: bitrate index must be in 1..14")
		os.Exit(2)
	}

	input := flag.Arg(0)
	output := flag.Arg(1)
	if output == "" {
		output = strings.TrimSuffix(input, filepath.Ext(input)) + ".mp1"
	}
	if _, err := os.Stat(output); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "error: %s already exists (use -f to overwrite)\n", output)
		os.Exit(1)
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

	tmp, err := os.CreateTemp(filepath.Dir(output), ".mp1enc-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	frames, err := mp1.Encode(tmp, in, cfg)
	if err != nil {
		_ = tmp.Close()
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if err := tmp.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if *force {
		_ = os.Remove(output)
	}
	if err := os.Rename(tmpName, output); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("%s -> %s: %d frames, %d kbit/s, %d Hz, %d ch\n",
		input, output, frames, mp1.Layer1Bitrates[cfg.BitrateIndex], in.SampleRate, in.Channels)
}
