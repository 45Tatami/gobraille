package converter

import "testing"
import "fmt"
import "image"
import _ "image/jpeg"
import "os"

func BenchmarkImage(t *testing.B) {
	file, _ := os.Open("../sample.jpg")
	img, _, _ := image.Decode(file)

	fmt.Println(GetConverter(EXACT, img, 32).Convert())
}
