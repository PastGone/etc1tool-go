# etc1tool (Go Version)

A command line tool for converting between the PNG and
[ETC1](https://registry.khronos.org/DataFormat/specs/1.3/dataformat.1.3.html#ETC1)
image formats. ETC1 is a lossy, fixed-rate (4 bits per texel; 64 bits = 8 bytes
per 4×4 texel block) texture compression format.

## Features

- Encode PNG files to ETC1 format
- Encode PNG files to ETC1S format (a subset of ETC1)
- Decode ETC1 files to PNG format
- Support for raw ETC1 data files (without PKM header)
- Support for showing the difference between original and encoded images

## Installation

### Prerequisites
- Go 1.20 or later

### Build
```bash
go build -o etc1tool.exe
```

## Usage

```bash
etc1tool infile [--help | --encode | --encodeNoHeader | --decode] [--showDifference difffile] [-o outfile]
```

### Options
- `--help`: Print usage information
- `--encode`: Create an ETC1 file from a PNG file (default)
- `--encodeETC1S`: Create an ETC1S file from a PNG file
- `--encodeNoHeader`: Create a raw ETC1 data file (without a header) from a PNG file
- `--decode`: Create a PNG file from an ETC1 file
- `--showDifference difffile`: Write difference between original and encoded image to difffile (only valid when encoding)
- `-o outfile`: Specify output file path

### Examples

#### Encode a PNG file to ETC1 format:
```bash
etc1tool input.png --encode -o output.pkm
```

#### Encode a PNG file to ETC1S format:
```bash
etc1tool input.png --encodeETC1S -o output.pkm
```

#### Decode an ETC1 file to PNG format:
```bash
etc1tool input.pkm --decode -o output.png
```

## License

Apache 2.0
