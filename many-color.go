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

var sizeRegex = regexp.MustCompile("^([1-9][0-9]*)x([1-9][0-9]*)$")
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
		fmt.Printf("error: %s\n\n", err)
		flag.Usage()
		os.Exit(1)
	}
	defer Close(input)

	s, err := parseSize(size)
	if err != nil {
		fmt.Printf("error: %s\n\n", err)
		flag.Usage()
		os.Exit(1)
	}

	var count int
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		hex := strings.TrimLeft(scanner.Text(), "#")
		if err := generateImage(hex, s); err != nil {
			fmt.Printf("Skipping %q due to error %s\n", hex, err)
			continue
		}
		count++
	}
	fmt.Printf("Generated %d images. Done!\n", count)
}

func generateImage(hex string, s Size) error {
	c, err := parseHex(hex)
	if err != nil {
		return err
	}
	img := image.NewRGBA(image.Rect(0, 0, s.Width, s.Height))
	for x := 0; x < s.Width; x++ {
		for y := 0; y < s.Height; y++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.OpenFile(fmt.Sprintf("%s.png", c.Name), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer Close(f)
	return png.Encode(f, img)
}

func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		fmt.Printf("fatal: %s", err)
		os.Exit(1)
	}
}

type Hex struct {
	color.Color
	Name string
}

func parseHex(str string) (Hex, error) {
	var hexParts []string
	for _, r := range colorRegex {
		hexParts = r.FindStringSubmatch(str)
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
		Name:  strings.Join(strs, ""),
	}, nil
}

type Size struct {
	Width, Height int
}

func parseSize(str string) (Size, error) {
	s := sizeRegex.FindStringSubmatch(str)
	if len(s) != 3 {
		return Size{}, errors.New("parse: invalid size")
	}
	var w, h int
	var err error
	if w, err = strconv.Atoi(s[1]); err != nil {
		return Size{}, err
	}
	if h, err = strconv.Atoi(s[2]); err != nil {
		return Size{}, err
	}
	return Size{w, h}, nil
}

func getInput(file string) (io.ReadCloser, error) {
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		return nil, errors.New("Pipe input is invalid")
	}

	return ioutil.NopCloser(bufio.NewReader(os.Stdin)), nil
}
