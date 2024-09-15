//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	video   js.Value
	stream  js.Value
	canvas  js.Value
	ctx     js.Value
	bgCache graycache
)

const (
	ScreenWidth  = 320
	ScreenHeight = 240
	buttonWidth  = 100
	buttonHeight = 50
)

func init() {
	doc := js.Global().Get("document")
	bgCache = make(graycache, ScreenWidth*ScreenHeight)
	video = doc.Call("createElement", "video")
	canvas = doc.Call("createElement", "canvas")
	video.Set("autoplay", true)
	video.Set("muted", true)
	video.Set("videoWidth", ScreenWidth)
	video.Set("videoHeight", ScreenHeight)
	mediaDevices := js.Global().Get("navigator").Get("mediaDevices")
	promise := mediaDevices.Call("getUserMedia", map[string]interface{}{
		"video": true,
		"audio": false,
	})
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		stream = args[0]
		video.Set("srcObject", stream)
		video.Call("play")
		canvas.Set("width", ScreenWidth)
		canvas.Set("height", ScreenHeight)
		ctx = canvas.Call("getContext", "2d")
		return nil
	}))
}

func fetchVideoFrame() []byte {
	ctx.Call("drawImage", video, 0, 0, ScreenWidth, ScreenHeight)
	data := ctx.Call("getImageData", 0, 0, ScreenWidth, ScreenHeight).Get("data")
	jsBin := js.Global().Get("Uint8Array").New(data)
	goBin := make([]byte, data.Get("length").Int())
	_ = js.CopyBytesToGo(goBin, jsBin)
	return goBin
}

func generateCurrentCache() graycache {
	return newGrayCacheFromData(fetchVideoFrame(), ScreenWidth, ScreenHeight)
}

type Game struct {
	drawImg *ebiten.Image
	button  *ebiten.Image
}

func buttonImage() *ebiten.Image {
	img := ebiten.NewImage(buttonWidth, buttonHeight)
	img.Fill(color.White)
	// write text ib button
	// drawText(img, "Capture", 10, 20, 20, color.Black)
	return img
}

func newGame() *Game {
	return &Game{
		drawImg: ebiten.NewImage(ScreenWidth, ScreenHeight),
		button:  buttonImage(),
	}
}

func (g *Game) Update() error {
	if !ctx.Truthy() {
		return nil
	}
	crCache := generateCurrentCache()
	diff := cacheDiffImg(bgCache, crCache, ScreenWidth, ScreenHeight)
	g.drawImg = ebiten.NewImageFromImage(diff)
	// g.drawImg = ebiten.NewImageFromImage(newImage(fetchVideoFrame(), ScreenWidth, ScreenHeight))

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		sx, sy := ScreenWidth/2-buttonWidth/2, ScreenHeight+buttonHeight
		if x >= sx && x <= sx+buttonWidth && y >= sy && y <= sy+buttonHeight {
			bgCache = generateCurrentCache()
		}
	}
	return nil
}

func newImage(data []byte, w, h int) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			// p := x*h + y // shadow clone
			p := y*w + x
			r := uint8(data[p*4])
			g := uint8(data[p*4+1])
			b := uint8(data[p*4+2])
			a := uint8(data[p*4+2])
			m.Set(x, y, color.RGBA{r, g, b, a})
		}
	}
	return m
}

type graycache []int

func newGrayCache(rgba *image.RGBA, w, h int) graycache {
	gc := make(graycache, w*h)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			r, g, b, _ := rgba.At(x, y).RGBA()
			gc[y*w+x] = grayscale(r, g, b, 0)
		}
	}
	return gc
}

func newGrayCacheFromData(data []byte, w, h int) graycache {
	gc := make(graycache, w*h)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			p := y*w + x
			r := uint8(data[p*4])
			g := uint8(data[p*4+1])
			b := uint8(data[p*4+2])
			gc[y*w+x] = grayscale(color.RGBA{r, g, b, 0}.RGBA())
		}
	}
	return gc
}

func grayscale(r, g, b, _ uint32) int {
	return int(0.299*float64(r/257) + 0.587*float64(g/257) + 0.114*float64(b/257))
}

func cacheDiffImg(bg, cr graycache, w, h int) *image.RGBA {
	diff := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			t := thresholding(cr[y*w+x]-bg[y*w+x], 30)
			diff.Set(x, y, color.RGBA{t, t, t, 255})
		}
	}
	return diff
}

func thresholding(diff, t int) uint8 {
	if diff > t {
		return 255
	}
	return 0
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.drawImg, nil)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ScreenWidth/2-buttonWidth/2, ScreenHeight+buttonHeight)
	screen.DrawImage(g.button, op)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%f", ebiten.ActualFPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth * 2, ScreenHeight * 2
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Lifting Gopher")
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
