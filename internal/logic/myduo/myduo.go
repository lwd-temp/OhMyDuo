package myduo

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"oh-my-duo/internal/consts"
	"oh-my-duo/internal/service"
	"oh-my-duo/internal/utils"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type sMyDuo struct {
	ImgSize         image.Rectangle
	FontUnicode     *truetype.Font
	FontAsciiBold   *truetype.Font
	BackgroundImage image.Image
	RoundMaskImage  image.Image
	BoxHeadImage    image.Image
	BoxFootImage    image.Image
}

func prepareFont(path string) *truetype.Font {
	tt, err := freetype.ParseFont(utils.GetResource("resource/public/resource/font/" + path))
	if err != nil {
		panic(err)
	}
	return tt
}

func prepareImage(path string) image.Image {
	im, _, err := image.Decode(
		bytes.NewReader(utils.GetResource(
			"resource/public/resource/image/" + path)),
	)
	if err != nil {
		panic(err)
	}
	return im
}

func init() {
	service.RegisterMyDuo(New())
}
func New() *sMyDuo {
	return &sMyDuo{
		ImgSize:         image.Rect(0, 0, 793, 793),
		FontUnicode:     prepareFont("OPPOSans-R.ttf"),
		FontAsciiBold:   prepareFont("DIN Next Rounded LT W05 Bold.ttf"),
		BackgroundImage: prepareImage("background-image.png"),
		RoundMaskImage:  prepareImage("round-mask.png"),
		BoxHeadImage:    prepareImage("box-head.png"),
		BoxFootImage:    prepareImage("box-foot.png"),
	}
}

func (sv *sMyDuo) Draw(ctx context.Context, elem consts.MyDuoElements) []byte {
	img := image.NewRGBA(sv.ImgSize)
	sv.drawBackground(img)
	sv.drawCharacter(img, elem.Character)
	// box
	sv.drawBox(img, 41, 97, 751, 422,
		color.RGBA{229, 229, 229, 255},
		color.RGBA{255, 255, 253, 255})
	// text
	textImg := sv.drawText(elem.OriginText)
	draw.Draw(img, img.Bounds(), textImg, textImg.Bounds().Min, draw.Over)
	// encode to bytes
	buff := new(bytes.Buffer)
	png.Encode(buff, img)
	return buff.Bytes()
}

func (sv *sMyDuo) drawBackground(img *image.RGBA) {
	draw.Draw(img, img.Bounds(), sv.BackgroundImage, sv.BackgroundImage.Bounds().Min, draw.Src)
}

func (sv *sMyDuo) drawCharacter(img *image.RGBA, character consts.MyDuoCharacters) {
	c := prepareImage("characters/" + character.ToString() + ".png")
	draw.Draw(img, img.Bounds(), c, c.Bounds().Min, draw.Over)
}

func (sv *sMyDuo) drawTextOnImg(img *image.RGBA, f font.Face, s string, x fixed.Int26_6, y fixed.Int26_6) {
	// max width = 635 px
	// text size = 64 px
	// line spacing = 18 px
	bg := image.NewUniform(color.RGBA{75, 75, 75, 255})
	point := fixed.Point26_6{X: x, Y: y}
	drawer := &font.Drawer{
		Dst:  img,
		Src:  bg,
		Face: f,
		Dot:  point,
	}
	drawer.DrawString(s)
}

func (sv *sMyDuo) drawText(s string) *image.RGBA {
	// max width = 635 px
	// text size = 64 px
	// line spacing = 18 px
	img := image.NewRGBA(image.Rect(0, 0, 635, 95))
	faceAsciiBold := truetype.NewFace(sv.FontAsciiBold, &truetype.Options{Size: 45})
	faceUnicode := truetype.NewFace(sv.FontUnicode, &truetype.Options{Size: 45})
	pieces := utils.SplitText(s)
	lines := 0
	cursor := fixed.Int26_6(0 * 64)
	max_width := fixed.Int26_6(635 * 64)
	for _, v := range pieces {
		var w fixed.Int26_6 = 0
		var f font.Face
		if v.Unicode {
			f = faceUnicode
		} else {
			f = faceAsciiBold
		}
		w = font.MeasureString(f, v.Text)
		// todo: split long words
		if cursor+w > max_width {
			// next line
			lines++
			cursor = 0
		}
		sv.drawTextOnImg(img, f, v.Text, cursor, fixed.Int26_6(40*64+(lines*50*64)))
		cursor += w
	}
	if lines == 0 {
		crop := image.NewRGBA(image.Rect(0, 0, 635, 53))
		draw.Draw(crop, crop.Bounds(), img, img.Bounds().Min, draw.Over)
		return crop
	}
	return img
}

func (sv *sMyDuo) drawBox(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA, fill color.RGBA) {
	r := 42
	draw.Draw(img, img.Bounds().Add(image.Point{x1, y1}), sv.BoxHeadImage, sv.BoxHeadImage.Bounds().Min, draw.Over)
	draw.Draw(img, img.Bounds().Add(image.Point{x1, y2 - r}), sv.BoxFootImage, sv.BoxFootImage.Bounds().Min, draw.Over)
	draw.Draw(img,
		image.Rectangle{Min: image.Point{x1 + 5, y1 + r}, Max: image.Point{x2 - 5, y2 - r + 5}},
		image.NewUniform(fill), image.Point{0, 0}, draw.Over)
	for i := 0; i < 5; i++ {
		sv.drawLine(img, x1+i, y1+r, x1+i, y2-r, c)
		sv.drawLine(img, x2-i, y1+r, x2-i, y2-r, c)
		// sv.drawLine(img, x1+r, y1+i, x2-r, y1+i, c)
		// sv.drawLine(img, x1+r, y2-i, x2-r, y2-i, c)
	}
	// for i := y1 + r + 5; i < ; i++ {

	// }
	// sv.drawRing(img, x1+r, y1+r, r-5, r, c, math.Pi, math.Pi/2*3)
	// sv.drawRing(img, x1+r, y2-r, r-5, r, c, -math.Pi/2, math.Pi/2*3)
}

func (sv *sMyDuo) drawRing(img *image.RGBA, x, y, innerRadius, outerRadius int, c color.RGBA, start, end float64) {
	if end < start {
		end += 2 * math.Pi
	}
	for i := x - outerRadius; i <= x+outerRadius; i++ {
		for j := y - outerRadius; j <= y+outerRadius; j++ {
			dx := i - x
			dy := j - y
			distance := math.Sqrt(float64(dx*dx + dy*dy))
			angle := math.Atan2(float64(dy), float64(dx))
			if distance >= float64(innerRadius) && distance <= float64(outerRadius) {
				if angle < 0 {
					angle += 2 * math.Pi
				}
				if angle >= start && angle <= end {
					img.SetRGBA(i, j, c)
				}
			}
		}
	}

}

func (sv *sMyDuo) drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		img.Set(x1, y1, c)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}
