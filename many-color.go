package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var sizeRegex = regexp.MustCompile("([1-9][0-9]*)x([1-9][0-9]*)")
var colorRegex = []*regexp.Regexp{
	regexp.MustCompile("^([0-9a-f][0-9a-f])([0-9a-f][0-9a-f])([0-9a-f][0-9a-f])$"),
	regexp.MustCompile("^([0-9a-f])([0-9a-f])([0-9a-f])$"),
}

func main() {
	var size, file string
	flag.StringVar(&size, "size", "800x600", "The output image size.")
	flag.StringVar(&file, "file", "", "Input file. (optional)")
	flag.Parse()

	input, err := getInput(file)
	if err != nil {
		flag.Usage()
		exitf(err.Error())
	}

	s := sizeRegex.FindStringSubmatch(size)
	if len(s) != 3 {
		exitf("Failed to parse size %s\n", size)
	}
	var width, height int
	if width, err = strconv.Atoi(s[1]); err != nil {
		exitf("Failed to parse width %s", s[1])
	}
	if height, err = strconv.Atoi(s[2]); err != nil {
		exitf("Failed to parse height %s", s[1])
	}
	fmt.Printf("Width: %dpx\nHeight: %dpx\n", width, height)

	defer input.Close()
	generateImages(input, width, height)
}

func generateImages(input io.Reader, width, height int) {
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		generateImage(scanner.Text(), width, height)
	}
}

func generateImage(rawHex string, width, height int) {
	hex := strings.TrimLeft(rawHex, "#")
	c, err := parseHex(hex)
	if err != nil {
		fmt.Printf("Skipping %q due to error %s\n", hex, err)
		return
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, c)
		}
	}
	filename := fmt.Sprintf("%s.png", c.hex)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		fmt.Printf("Skipping %q due to error %s\n", hex, err)
		return
	}
	png.Encode(f, img)
	fmt.Printf("% 7s > %s\n", rawHex, filename)
}

type Hex struct {
	color.Color
	hex string
}

func parseHex(hex string) (Hex, error) {
	var hexParts []string
	for _, r := range colorRegex {
		hexParts = r.FindStringSubmatch(hex)
		if len(hexParts) == 4 {
			break
		}
	}
	if len(hexParts) != 4 {
		return Hex{}, errors.New("parse: not a hex color")
	}

	ints := make([]uint8, 3)
	strs := make([]string, 3)
	for i, str := range hexParts[1:] {
		if len(str) == 1 {
			str += str
		}
		v, _ := strconv.ParseInt(str, 16, 0)
		ints[i] = uint8(v)
		strs[i] = str
	}
	return Hex{
		Color: color.RGBA{ints[0], ints[1], ints[2], 255},
		hex:   strings.Join(strs, ""),
	}, nil
}

func getInput(file string) (io.ReadCloser, error) {
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Reading from file %s\n", file)
		return f, nil
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		return nil, errors.New("Pipe input is invalid")
	}

	fmt.Printf("Reading from stdin\n")
	return ioutil.NopCloser(bufio.NewReader(os.Stdin)), nil
}

func exitf(template string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("error: %s\n", template), args...)
	os.Exit(1)
}
