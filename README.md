# mp1enc

Small, dependency-free MPEG-1 Audio Layer I encoder written in Go, ported
from the MIT-licensed Fortran 2003/08 encoder
[uzura1](https://github.com/cure-honey/uzura1).

The port keeps the original architecture: polyphase analysis filter bank,
psychoacoustic model, Layer I bit allocation, quantization, and the MPEG
frame/bitstream writer.

**Status:** 0.1.0 — usable for legacy/interoperability work and experiments.
This is not a replacement for ffmpeg or a high-end psychoacoustic encoder.

## Install

```bash
go install github.com/daregod/mp1-enc/cmd/mp1enc@latest
go install github.com/daregod/mp1-enc/cmd/mp1info@latest
```

Or build from a checkout:

```bash
make build
# or
go build -buildvcs=false -o mp1enc ./cmd/mp1enc
go build -buildvcs=false -o mp1info ./cmd/mp1info
```

## Usage

```bash
mp1enc -b 12 input.wav output.mp1
```

Options:

* `-b 1..14` — Layer I bitrate index (default 12 = 384 kbit/s at 44.1 kHz);
* `-crc` — enable CRC16 error protection;
* `-c` — set the copyright bit;
* `-o` — set the original bit (default on);
* `-f` — overwrite an existing output file;
* `-version` — print the version.

If `output.mp1` already exists, the encoder refuses to overwrite it unless
`-f` is given. Output is written through a temporary file and renamed only
after successful encoding.

`mp1info` prints frame/header information for an MP1 file.

## Limitations

* input must be 16-bit PCM WAV;
* mono or stereo only;
* 32, 44.1 or 48 kHz;
* no resampling, metadata, stdin/stdout streaming, or multi-channel audio;
* a final partial 384-sample frame is padded with silence;
* quality is the quality of the original simple encoder, not a modern
  psychoacoustic implementation.

## Validation

The encoder has been checked with ffmpeg (`codec_name=mp1`): 384 kbit/s
stereo, 256 kbit/s, mono, 48 kHz and 32 kHz decode successfully. A 440 Hz
sine at 384 kbit/s decodes at about 51 dB SNR, comparable to `mp2enc -l 1`
(50.9 dB) at the same bitrate; frame count and file size match `mp2enc`.

`go test ./...` covers the WAV reader, FFT against a naive DFT, frame
headers/sizes, CRC/header handling, partial-frame padding, and invalid input.

## Layout

```
cmd/mp1enc      command-line encoder
cmd/mp1info     MP1 frame/header inspector
internal/mp1    filter bank, psychoacoustics, allocation, bitstream
internal/wav    16-bit PCM WAV reader
```

## License

MIT. See [LICENSE](LICENSE).

The original Fortran encoder is Copyright (c) 2017 cure honey and is also
MIT-licensed. Its sources are not part of this repository; this repository
contains only the Go port and its attribution.
