package main

import "fmt"
import "os"
import "log"
import "strconv"
import (
	"image"
	_ "image/jpeg"
	_ "image/png"
)
import "github.com/45Tatami/gobraille/converter"


func main() {
	var width int = 2
	var args []string = os.Args[1:]

	if len(args) > 1{
		imageIn := loadImage(args[0])
		w, err := strconv.Atoi(args[1])

		if err != nil {
			log.Fatal(err)
		}
		if w < 1 {
			log.Fatal("Width out of bounds")
		}
		width = w
		dyn := converter.GetConverter(converter.PTP_AVERAGED, imageIn, width)
		fmt.Println(dyn.Convert())
	} else { // Interactive mode
		fmt.Println("Not implemented yet")
	}

}

func loadImage(path string) image.Image {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		f.Close()
	}(file)

	// Load image
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	return img
}
