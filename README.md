# mp1enc

Go port of the MPEG-1 Audio Layer I encoder
[uzura1](https://github.com/cure-honey/uzura1) by cure-honey (original
Fortran 2003 implementation).

The port keeps the same architecture: polyphase analysis filter bank,
psychoacoustic model (FFT, absolute threshold, spreading), Layer I bit
allocation and quantization, and the MPEG frame/bitstream writer.

## Build

```bash
go build -o mp1enc ./cmd/mp1enc
go build -o mp1info ./cmd/mp1info    # frame/header inspector
# or
make go-build
```

## Usage

```bash
./mp1enc -b 12 input.wav output.mp1
```

Options:

* `-b 1..14` — Layer I bitrate index (default 12 = 384 kbit/s at 44.1 kHz);
* `-crc` — enable CRC16 error protection;
* `-c` — set copyright bit; `-o` — original bit (default on).

Supported input: 16-bit PCM WAV, mono or stereo, 32 / 44.1 / 48 kHz.

## Validation

The output was checked with ffmpeg (`codec_name=mp1`): 384 kbit/s stereo,
256 kbit/s, mono, 48 and 32 kHz all decode. A 440 Hz sine at 384 kbit/s
decodes at about 51 dB SNR, on par with `mp2enc -l 1` (50.9 dB) at the same
bitrate; frame count and file size match `mp2enc` exactly.

`go test ./...` covers the WAV reader, the FFT against a naive DFT, and the
frame header/size of the encoder.

## Layout

```
cmd/mp1enc      command line encoder
cmd/mp1info     MP1 frame/header/padding inspector
internal/mp1    filter bank, psychoacoustics, allocation, bitstream
internal/wav    16-bit PCM WAV reader
```

## Upstream

* Original Fortran encoder: <https://github.com/cure-honey/uzura1>
* The Fortran sources are not part of this repository; it contains only the
  Go port.