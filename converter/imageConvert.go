package converter

import (
	"image"
	"image/color"
	"fmt"
	"math"
)

const STATIC, DYNAMIC, EXACT string = "static", "dynamic", "exact"
const FULL, HALF int = 0xffff, FULL/2

func GetConverter(cnvType string, img image.Image, width int) converter {
	switch cnvType {
	case DYNAMIC:
		return dynamicConverter{staticConverter: staticConverter{genericConverter: genericConverter{img: img, width: width, height: width*2}, colors: defaultColors}}
	case EXACT:
		slice := make([][][2]int, img.Bounds().Max.Y/(width*2))
		for i := range slice {
			slice[i] = make([][2]int, img.Bounds().Max.X/width)
		}
		return exactConverter{dynamicConverter: dynamicConverter{staticConverter: staticConverter{genericConverter: genericConverter{img:img, width: 2, height: 4}, colors: defaultColors}}, diff: slice, scale: width/2}
	default:
		return staticConverter{genericConverter: genericConverter{img: img, width: width, height: width*2}, colors: defaultColors}
	}
}



var defaultColors = []int{0x28FF, 0x28FE, 0x28B7, 0x28AB, 0x2868, 0x286A, 0x2848, 0x2840, 0x2800}

type converter interface {
	Convert() string
}

type genericConverter struct {
	img image.Image
	width, height int
}

type staticConverter struct {
	genericConverter
	colors []int
}

type dynamicConverter struct {
	staticConverter
	avgLumosity int
}

type exactConverter struct {
	dynamicConverter
	diff [][][2]int
	scale int
}


func (this staticConverter) Convert() string {
	var convertedImage string

	for y := 0; y < this.img.Bounds().Max.Y; y += this.height {
		for x := 0; x < this.img.Bounds().Max.X; x += this.width {
			convertedImage += this.convertBlock(x,y)
		}
		convertedImage += "\n"
	}

	return convertedImage
}

func (this staticConverter) convertBlock(x, y int) string {
	return this.brailleChar(this.blockLumosity(x, y))
}

func (this staticConverter) brailleChar(grayscale int) string {
	return string(this.colors[grayscale / (0xffff/len(this.colors) + 1)])
}

func (this staticConverter) blockLumosity(x,y int) int {
	var lumSum, count int

	for dy := 0; dy < this.height; dy++ {
		for dx := 0; dx < this.width; dx++ {
			if y+dy >= this.img.Bounds().Max.Y {
				break
			}
			if x+dx >= this.img.Bounds().Max.X {
				continue
			}
			pix := this.img.At(x+dx, y+dy)
			lumSum += grayscaleValue(pix)
			count++
		}
	}

	var lumen int
	if count > 0 {
		lumen = (lumSum/count)
	} else {
		lumen = 0xffff
	}

	return lumen
}

func (this dynamicConverter) Convert() string {
	this.setAvgLum()
	var lumSum, count int
	for y := 0; y < this.img.Bounds().Max.Y; y++ {
		for x := 0; x < this.img.Bounds().Max.X; x++ {
			lumSum += grayscaleValue(this.img.At(x,y))
			count++
		}
	}
	this.avgLumosity = lumSum/count

	var convertedImage string

	for y := 0; y < this.img.Bounds().Max.Y; y += this.height {
		for x := 0; x < this.img.Bounds().Max.X; x += this.width {
			convertedImage += this.convertBlock(x,y)
		}
		convertedImage += "\n"
	}

	return convertedImage
}

func (this *dynamicConverter) setAvgLum() {
	var lumSum, count int
	for y := 0; y < this.img.Bounds().Max.Y; y++ {
		for x := 0; x < this.img.Bounds().Max.X; x++ {
			lumSum += grayscaleValue(this.img.At(x,y))
			count++
		}
	}
	this.avgLumosity = lumSum/count
}

func (this dynamicConverter) convertBlock(x, y int) string {
	lumosity := this.staticConverter.blockLumosity(x, y)
	return this.brailleChar(lumosity)
}

func (this dynamicConverter) brailleChar(lumosity int) string {
	upperBound := this.avgLumosity*2
	lowerBound := this.avgLumosity - (0xFFFF - this.avgLumosity)

	if upperBound > 0xFFFF {upperBound = 0xFFFF }
	if lowerBound < 0 { lowerBound = 0 }

	var index int
	switch {
	case lumosity > upperBound :
		index = len(this.colors) -1
	case lumosity < lowerBound:
		index = 0
//		fmt.Printf("Index: %d Lum: %X Up: %X Low: %X\n", index, lumosity, upperBound, lowerBound)
	default:
		var steps int = int((upperBound - lowerBound)/len(this.colors) + 1)
		index = (lumosity - lowerBound)/steps
//		fmt.Printf("Index: %d Lum: %X Up: %X Low: %X\n", index, lumosity, upperBound, lowerBound)
		if index >= len(this.colors) {
			fmt.Printf("Index: %d Lum: %X Up: %X Low: %X", index, lumosity, upperBound, lowerBound)
		}
	}

	return string(this.colors[index])
}

func (this exactConverter) Convert() string {
	this.setAvgLum()
	this.avgLumosity = (HALF + this.avgLumosity)/2
	var convertedImage string

	if this.width < 2 {
		this.width = 2
		this.height = 4
	}

	for y := 0; y < this.img.Bounds().Max.Y; y += this.height*this.scale {
		for x := 0; x < this.img.Bounds().Max.X; x += this.width*this.scale {
			convertedImage += this.convertBlock(x,y)
		}
		convertedImage += "\n"
	//	fmt.Println()
	}

	return convertedImage
}

func (this exactConverter) convertBlock(x, y int) string {
	var char uint = 0x2800
	var sum uint = 0x0

	for sx := 0; sx < 2; sx++ {
		for sy := 0; sy < 3; sy++ {
			sum = sum | (this.isAbove(x,y,sx,sy) << uint8(sy + 3*sx))
		}
	}
	sum = sum | (this.isAbove(x,y,0,3)  << 6)
	sum = sum | (this.isAbove(x,y,1,3)  << 7)

	var pc uint

	for i := sum; i > 0; i = i >> 1 {
		pc += (i & 0x1)
	}

	// TODO check on squares, not "blocks"
	// TODO check Neighbours?
	var maxP int = 8
	var wc int = maxP - int(pc)
	var pointDifference int = int(math.Floor(float64(this.blockLumosity(x,y))/float64(FULL/maxP) + 0.5)) - wc

	var i uint
	for i = 0; i < 8; i++ {
		switch {
		case pointDifference == 0:
				break
		case pointDifference > 0:
			if ((sum >> i) & 0x1) == 0 {continue}
			sum -= 0x1 << i
			pointDifference--
		case pointDifference < 0:
			if ((sum >> i) & 0x1) == 1 {continue}
			sum += 0x1 << i
			pointDifference++
		}
	}
	if pointDifference != 0 {
		fmt.Printf("Point difference of %d in %d,%d\n", pointDifference, x, y)
	}
//	fmt.Printf("%d ", pointDifference)

	return string(char + sum)
}

func (this exactConverter) isAbove(x, y, dx, dy int) uint {
	var sum int
	for iy := 0; iy < this.scale; iy++ {
		for ix := 0; ix < this.scale; ix++ {
			sum += grayscaleValue(this.img.At(x+dx*this.scale+ix, y+dy*this.scale+iy))
		}
	}

	average := sum / (this.scale*this.scale)

//	if average < this.calcLocalAverage(x, y) {
	if average < 0xFFFF/2 {
		return 1
	} else {
		return 0
	}
}

func (this staticConverter) calcLocalAverage(x, y int) int {
	var sum, count int
	dr := 1
	for dy := -dr; dy <= dr; dy++ {
		for dx := dr; dx <= dr; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			sum += this.blockLumosity(x+dx, y+dy)
			count++
		}
	}
	return (sum/count)
	// TODO Performance, check in array or something
}


func grayscaleValue(pixel color.Color) int {
	r, g, b, _ := pixel.RGBA()
	return int(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
}
