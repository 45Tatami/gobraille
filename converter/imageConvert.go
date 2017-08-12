package converter

import (
	"image"
	"image/color"
//	"fmt"
	"math"
)

const FULL, HALF int = 0xffff, FULL/2
var defaultColors = []int{0x28FF, 0x28FE, 0x28B7, 0x28AB, 0x2868, 0x286A, 0x2848, 0x2840, 0x2800}

type ConverterType string
const PTP, PTP_AVERAGED  ConverterType  = "point_to_point", "ptp_averaged"

type ImageConverter interface {
	Convert() string
	SetPicture(image.Image)
	SetScale(int)
}

type converter struct {
	img image.Image
	Scale int
	lum []int
	convType ConverterType
	silent bool
}

func GetConverter(cnvType ConverterType, img image.Image, scale int) ImageConverter {
	return &converter{img, scale, nil, cnvType, false}
}

func (this converter) SetPicture(img image.Image) {
	this.img = img
	this.lum = nil
}

func (this converter) SetScale(scale int) {
	this.Scale = scale
	this.lum = nil
}

// Calculates and returns the brightness of a 2*4 (*scale) block starting at the given coordinates
func (this converter) blockLumosity(x,y int) int {
	var lumSum, count int
	var height, width int = 4*this.Scale, 2*this.Scale

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
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

func (this converter) Convert() string {
	var convertedImage string

	for y := 0; y < this.img.Bounds().Max.Y; y += 4*this.Scale   {
		for x := 0; x < this.img.Bounds().Max.X; x += 2*this.Scale {
			char := this.convertBlockPointForPoint(x,y)
			if this.convType == PTP_AVERAGED {
				char = this.averageOutBlock(x, y, char)
			}
			convertedImage += string(char)
		}
		convertedImage += "\n"
	}

	return convertedImage
}

func (this converter) averageOutBlock(x, y int, sum uint) uint {
	// TODO check on squares, not "blocks"
	// TODO check Neighbours?
	var pc, dots uint

	dots = sum - 0x2800

	for i := dots; i > 0; i = i >> 1 {
		pc += (i & 0x1)
	}

	var pointDifference int = int(math.Floor(float64(this.blockLumosity(x,y))/float64(FULL/8) + 0.5)) - int(8 - pc)

	var i uint
	for i = 0; i < 8; i++ {
		switch {
		case pointDifference == 0:
				break
		case pointDifference > 0:
			if (dots & (0x1 << i)) == 0 {continue}
			sum -= 0x1 << i
			pointDifference--
		case pointDifference < 0:
			if (dots & (0x1 << i)) == 1 {continue}
			sum += 0x1 << i
			pointDifference++
		}
	}

	return sum
}


func (this converter) convertBlockPointForPoint(x, y int) uint {
	var char uint = 0x2800 // Base for utf-8 braille

	for sx := 0; sx < 2; sx++ {
		for sy := 0; sy < 3; sy++ {
			char = char | (this.isAbove(x,y,sx,sy) << uint8(sy + 3*sx))
		}
		char = char | (this.isAbove(x,y,sx,3)  << uint8(6+sx))
	}
	return char
}

func (this converter) isAbove(x, y, dx, dy int) uint {
	var sum int
	for iy := 0; iy < this.Scale; iy++ {
		for ix := 0; ix < this.Scale; ix++ {
			sum += grayscaleValue(this.img.At(x+dx*this.Scale+ix, y+dy*this.Scale+iy))
		}
	}

	average := sum / (this.Scale*this.Scale)

	if average < FULL/2 {
		return 1
	} else {
		return 0
	}
}

func (this converter) calcLocalAverage(x, y int) int {
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
