package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"etc1tool/etc1"
)

var (
	exeName = "etc1tool"
)

func usage(message string) {
	if message != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n", message)
		fmt.Fprintf(os.Stderr, "usage:\n")
	}

	fmt.Fprintf(os.Stderr, "%s infile [--help | --encode | --encodeNoHeader | --decode] [--showDifference difffile] [-o outfile]\n", exeName)
	fmt.Fprintf(os.Stderr, "\tDefault is --encode\n")
	fmt.Fprintf(os.Stderr, "\t\t--help           print this usage information.\n")
	fmt.Fprintf(os.Stderr, "\t\t--encode         create an ETC1 file from a PNG file.\n")
	fmt.Fprintf(os.Stderr, "\t\t--encodeETC1S    create an ETC1S file from a PNG file.\n")
	fmt.Fprintf(os.Stderr, "\t\t--encodeNoHeader create a raw ETC1 data file (without a header) from a PNG file.\n")
	fmt.Fprintf(os.Stderr, "\t\t--decode         create a PNG file from an ETC1 file.\n")
	fmt.Fprintf(os.Stderr, "\t\t--showDifference difffile    Write difference between original and encoded\n")
	fmt.Fprintf(os.Stderr, "\t\t                             image to difffile. (Only valid when encoding).\n")
	fmt.Fprintf(os.Stderr, "\tIf outfile is not specified, an outfile path is constructed from infile,\n")
	fmt.Fprintf(os.Stderr, "\twith the apropriate suffix (.pkm or .png).\n")
	os.Exit(1)
}

func changeExtension(path, extension string) (string, error) {
	ext := filepath.Ext(path)
	if ext != "" {
		path = path[:len(path)-len(ext)]
	}
	return path + extension, nil
}

func readPNGFile(input string) ([]byte, uint32, uint32, error) {
	file, err := os.Open(input)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("could not open input file %s: %v", input, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("could not decode PNG file: %v", err)
	}

	bounds := img.Bounds()
	width := uint32(bounds.Dx())
	height := uint32(bounds.Dy())
	stride := width * 3

	data := make([]byte, stride*height)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			offset := (y*int(stride) + x*3)
			data[offset] = byte(r >> 8)
			data[offset+1] = byte(g >> 8)
			data[offset+2] = byte(b >> 8)
		}
	}

	return data, width, height, nil
}

func readPKMFile(input string) ([]byte, uint32, uint32, error) {
	file, err := os.Open(input)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("could not open input file %s: %v", input, err)
	}
	defer file.Close()

	header := make([]byte, etc1.PKMHeaderSize)
	_, err = file.Read(header)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("could not read header from input file %s: %v", input, err)
	}

	if !etc1.PKMIisValid(header) {
		return nil, 0, 0, fmt.Errorf("bad PKM header for input file %s", input)
	}

	width := etc1.PKMGetWidth(header)
	height := etc1.PKMGetHeight(header)
	encodedSize := etc1.GetEncodedDataSize(width, height)

	encodedData := make([]byte, encodedSize)
	_, err = file.Read(encodedData)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("could not read encoded data from input file %s: %v", input, err)
	}

	stride := width * 3
	imageData := make([]byte, stride*height)
	result := etc1.DecodeImage(encodedData, imageData, width, height, 3, stride)
	if result != 0 {
		return nil, 0, 0, fmt.Errorf("could not decode image: error code %d", result)
	}

	return imageData, width, height, nil
}

func writePNGFile(output string, width, height uint32, data []byte) error {
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	stride := width * 3

	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			offset := y*int(stride) + x*3
			r := data[offset]
			g := data[offset+1]
			b := data[offset+2]
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("could not open output file %s: %v", output, err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("could not encode PNG: %v", err)
	}

	return nil
}

func encode(input, output string, emitHeader, bETC1S bool, diffFile string) error {
	sourceImage, width, height, err := readPNGFile(input)
	if err != nil {
		return err
	}

	encodedSize := etc1.GetEncodedDataSize(width, height)
	encodedData := make([]byte, encodedSize)

	result := etc1.EncodeImage(sourceImage, width, height, 3, width*3, encodedData, bETC1S)
	if result != 0 {
		return fmt.Errorf("could not encode image: error code %d", result)
	}

	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("could not open output file %s: %v", output, err)
	}
	defer file.Close()

	if emitHeader {
		header := make([]byte, etc1.PKMHeaderSize)
		etc1.PKMFormatHeader(header, width, height)
		_, err = file.Write(header)
		if err != nil {
			return fmt.Errorf("could not write header to output file %s: %v", output, err)
		}
	}

	_, err = file.Write(encodedData)
	if err != nil {
		return fmt.Errorf("could not write encoded data to output file %s: %v", output, err)
	}

	if diffFile != "" {
		diffImage, outWidth, outHeight, err := readPKMFile(output)
		if err != nil {
			return err
		}

		if outWidth != width || outHeight != height {
			return fmt.Errorf("output file has incorrect bounds: %d, %d != %d, %d", outWidth, outHeight, width, height)
		}

		src := sourceImage
		dest := diffImage
		size := width * height * 3

		for i := uint32(0); i < size; i++ {
			diff := int(src[i]) - int(dest[i])
			diff *= diff
			diff <<= 3
			if diff < 0 {
				diff = 0
			} else if diff > 255 {
				diff = 255
			}
			dest[i] = byte(diff)
		}

		err = writePNGFile(diffFile, outWidth, outHeight, diffImage)
		if err != nil {
			return err
		}
	}

	return nil
}

func decode(input, output string) error {
	imageData, width, height, err := readPKMFile(input)
	if err != nil {
		return err
	}

	err = writePNGFile(output, width, height, imageData)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) > 0 {
		exeName = filepath.Base(os.Args[0])
	}

	var input, output, diffFile string
	var encodeDecodeSeen, shouldEncode, encodeETC1S, encodeHeader, showDifference bool

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			if arg == "-o" {
				if output != "" {
					usage("Only one -o flag allowed.")
				}
				if i+1 >= len(os.Args) {
					usage("Expected outfile after -o")
				}
				i++
				output = os.Args[i]
			} else if strings.HasPrefix(arg, "--") {
				switch arg {
				case "--encode":
					if encodeDecodeSeen {
						usage("At most one occurrence of --encode --encodeETC1S --encodeNoHeader or --decode is allowed.")
					}
					encodeDecodeSeen = true
					shouldEncode = true
					encodeHeader = true
				case "--encodeETC1S":
					if encodeDecodeSeen {
						usage("At most one occurrence of --encode --encodeETC1S --encodeNoHeader or --decode is allowed.")
					}
					encodeDecodeSeen = true
					shouldEncode = true
					encodeETC1S = true
					encodeHeader = true
				case "--encodeNoHeader":
					if encodeDecodeSeen {
						usage("At most one occurrence of --encode --encodeETC1S --encodeNoHeader or --decode is allowed.")
					}
					encodeDecodeSeen = true
					shouldEncode = true
					encodeHeader = false
				case "--decode":
					if encodeDecodeSeen {
						usage("At most one occurrence of --encode --encodeETC1S --encodeNoHeader or --decode is allowed.")
					}
					encodeDecodeSeen = true
				case "--showDifference":
					if showDifference {
						usage("Only one --showDifference option allowed.")
					}
					showDifference = true
					if i+1 >= len(os.Args) {
						usage("Expected difffile after --showDifference")
					}
					i++
					diffFile = os.Args[i]
				case "--help":
					usage("")
				default:
					usage(fmt.Sprintf("Unknown flag %s", arg))
				}
			} else {
				usage(fmt.Sprintf("Unknown flag %s", arg))
			}
		} else {
			if input != "" {
				usage(fmt.Sprintf("Only one input file allowed. Already have %s, now see %s", input, arg))
			}
			input = arg
		}
	}

	if !encodeDecodeSeen {
		shouldEncode = true
		encodeHeader = true
	}

	if !shouldEncode && showDifference {
		usage("--showDifference is only valid when encoding.")
	}

	if input == "" {
		usage("Expected an input file.")
	}

	if output == "" {
		var defaultExtension string
		if shouldEncode {
			defaultExtension = ".pkm"
		} else {
			defaultExtension = ".png"
		}
		var err error
		output, err = changeExtension(input, defaultExtension)
		if err != nil {
			usage(fmt.Sprintf("Could not change extension of input file name: %s", input))
		}
	}

	var err error
	if shouldEncode {
		err = encode(input, output, encodeHeader, encodeETC1S, diffFile)
	} else {
		err = decode(input, output)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
