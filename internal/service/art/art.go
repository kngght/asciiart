package art

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	_ "golang.org/x/image/webp"
)

const chars = ` .-':_,^=;><+!rc*/z?sLTv)J7(|Fi{C}fI31tlu[neoZ5Yxjya]2ESwqkP6h9d4VpOGbUAKXHm8RD#$Bg0MNWQ%&@`

type art interface {
	AsciiArt()
}

func makeCharImg(img image.Image, wdt int) string {
	dst := toScale(img, wdt)
	var asciiArt strings.Builder
	asciiArt.Grow(dst.Bounds().Dx() * dst.Bounds().Dy() * 25)
	for y := dst.Bounds().Min.Y; y < dst.Bounds().Max.Y; y++ {
		for x := dst.Bounds().Min.X; x < dst.Bounds().Max.X; x++ {
			c := color.RGBAModel.Convert(dst.At(x, y))
			r, g, b, _ := c.RGBA()
			asciiArt.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r>>8, g>>8, b>>8, tochar(c)))
		}
		asciiArt.WriteString("\n")
	}
	return asciiArt.String()
}

func toScale(src image.Image, wdt int) image.Image {
	srcW := float64(src.Bounds().Dx())
	srcH := float64(src.Bounds().Dy())
	hght := int(math.Round((float64(wdt) * (float64(src.Bounds().Dy()) / float64(src.Bounds().Dx())) * 0.47)))
	dst := image.NewRGBA(image.Rectangle{image.Point{}, image.Point{wdt, hght}})
	for y := 0; y < hght; y++ {
		y0 := int(math.Round(float64(y) * srcH / float64(hght)))
		y1 := int(math.Round(float64(y+1) * srcH / float64(hght)))
		for x := 0; x < wdt; x++ {
			x0 := int(math.Round(float64(x) * srcW / float64(wdt)))
			x1 := int(math.Round(float64(x+1) * srcW / float64(wdt)))
			rect := image.Rect(
				src.Bounds().Min.X+x0, src.Bounds().Min.Y+y0,
				src.Bounds().Min.X+x1, src.Bounds().Min.Y+y1,
			)
			dst.Set(x, y, averageColor(src, rect))
		}
	}
	return dst
}

func averageColor(src image.Image, rect image.Rectangle) color.RGBA {
	rect = rect.Intersect(src.Bounds())
	if rect.Empty() {
		return color.RGBA{}
	}
	sum := uint64(rect.Dx() * rect.Dy())
	var sumR, sumG, sumB, sumA uint64
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r, g, b, a := color.RGBAModel.Convert(src.At(x, y)).(color.RGBA).RGBA()
			sumR += uint64(r)
			sumG += uint64(g)
			sumB += uint64(b)
			sumA += uint64(a)
		}
	}
	return color.RGBA{
		R: uint8(sumR / sum >> 8),
		G: uint8(sumG / sum >> 8),
		B: uint8(sumB / sum >> 8),
		A: uint8(sumA / sum >> 8),
	}
}

func tochar(c color.Color) string {
	r, g, b, _ := c.RGBA()
	y := (19595*r + 38470*g + 7471*b + 1<<15) >> 24
	i := int(float64(len(chars))*float64(y)) >> 8
	return string(chars[i])
}

func closeFile(f *os.File) {
	err := f.Close()
	mayBeErr(err)
}
func mayBeErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func giffing(g *gif.GIF, wdt int) *charGif {
	fnl := charGif{
		d: append(make([]int, 0, len(g.Delay)), g.Delay...),
		s: make([]string, len(g.Image)),
	}

	canH := g.Config.Height
	canW := g.Config.Width

	var bgColor color.Color = color.Transparent
	if len(g.Image) > 0 && int(g.BackgroundIndex) < len(g.Image[0].Palette) {
		bgColor = g.Image[0].Palette[g.BackgroundIndex]
	}

	canvas := image.NewRGBA(image.Rect(0, 0, canW, canH))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(bgColor), canvas.Rect.Min, draw.Src)

	for i, frame := range g.Image {
		disposal := byte(g.Disposal[i])

		var savedCarvas *image.RGBA
		if disposal == gif.DisposalPrevious {
			savedCarvas := image.NewRGBA(canvas.Rect)
			draw.Draw(savedCarvas, savedCarvas.Rect, canvas, canvas.Rect.Min, draw.Src)
		}
		draw.Draw(canvas, frame.Bounds(), frame, frame.Rect.Min, draw.Over)
		fnl.s[i] = makeCharImg(canvas, wdt)

		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, frame.Rect, image.NewUniform(bgColor), canvas.Rect.Min, draw.Src)
		case gif.DisposalPrevious:
			draw.Draw(canvas, frame.Rect, savedCarvas, canvas.Rect.Min, draw.Src)
		}
	}
	return &fnl
}

func MakeASCIIArt(path string, wdt int) {
	file, err := os.Open(path)
	mayBeErr(err)
	defer closeFile(file)
	_, format, err := image.DecodeConfig(file)
	mayBeErr(err)

	// Сброс позиции файла в начало
	_, err = file.Seek(0, io.SeekStart)
	mayBeErr(err)

	switch format {
	// Для формата gif
	case "gif":
		imgGif, err := gif.DecodeAll(file)
		mayBeErr(err)
		charGifPtr := giffing(imgGif, wdt)
		charGifPtr.asciiArt()
	// Для формата jpeg, png, webp
	default:
		img, _, err := image.Decode(file)
		mayBeErr(err)
		fmt.Print(makeCharImg(img, wdt))
	}

}

type charGif struct {
	d []int
	s []string
}

func (c charGif) asciiArt() {
	for {
		for i, f := range c.s {
			time.Sleep(time.Duration(c.d[i] * 10 * int(time.Millisecond)))
			fmt.Print("\x1b[H\x1b[2J")
			fmt.Print(f)
		}
	}
}
