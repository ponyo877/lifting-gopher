//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"syscall/js"

	"github.com/ponyo877/lifting-gopher/img"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	video            js.Value
	stream           js.Value
	canvas           js.Value
	ctx              js.Value
	bgCache          graycache
	gopher           *ebiten.Image
	arcadeFaceSource *text.GoTextFaceSource
)

const (
	ScreenWidth  = 320
	ScreenHeight = 240
	buttonWidth  = 100
	buttonHeight = 50
)

//go:embed img/*
var files embed.FS

func init() {
	img, _, err := image.Decode(bytes.NewReader(img.Gopher))
	if err != nil {
		log.Fatal(err)
	}
	gopher = ebiten.NewImageFromImage(img)
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}
	arcadeFaceSource = s
	bgCache = make(graycache, ScreenWidth*ScreenHeight)

	doc := js.Global().Get("document")
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

type Game struct {
	drawImg *ebiten.Image
	button  *ebiten.Image
	paths   []vector.Path
	y       float64
	gv      float64
}

func buttonImage() *ebiten.Image {
	fsize := 10.0
	img := ebiten.NewImage(buttonWidth, buttonHeight)
	img.Fill(color.White)
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(img.Bounds().Dx())/2, float64(img.Bounds().Dy())/2-fsize/2)
	op.ColorScale.ScaleWithColor(color.Black)
	op.LineSpacing = fsize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(img, "BG CAPTURE", &text.GoTextFace{
		Source: arcadeFaceSource,
		Size:   fsize,
	}, op)
	return img
}

func newGame() *Game {
	return &Game{
		drawImg: ebiten.NewImage(ScreenWidth, ScreenHeight),
		button:  buttonImage(),
		paths:   []vector.Path{},
		y:       0,
		gv:      0.25,
	}
}

func (g *Game) Update() error {
	if !ctx.Truthy() {
		return nil
	}
	g.y += g.gv
	g.gv += 0.01
	goBin := fetchVideoFrame()
	crCache := newGrayCacheFromData(goBin, ScreenWidth, ScreenHeight)
	mp := cacheDiffBitmap(bgCache, crCache, ScreenWidth, ScreenHeight)
	if mp[Point{ScreenWidth / 2, int(g.y)}] {
		g.gv = -1
	} else if g.y > ScreenHeight-float64(gopher.Bounds().Dy())/2 {
		g.gv = 0
	}

	g.drawImg = ebiten.NewImageFromImage(newImage(goBin, ScreenWidth, ScreenHeight))

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		sx, sy := ScreenWidth/2-buttonWidth/2, ScreenHeight+buttonHeight
		if x >= sx && x <= sx+buttonWidth && y >= sy && y <= sy+buttonHeight {
			bgCache = newGrayCacheFromData(goBin, ScreenWidth, ScreenHeight)
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.drawImg, nil)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ScreenWidth/2-buttonWidth/2, ScreenHeight+buttonHeight)
	screen.DrawImage(g.button, op)

	opg := &ebiten.DrawImageOptions{}
	w, h := gopher.Bounds().Dx(), gopher.Bounds().Dy()
	opg.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
	opg.GeoM.Translate(ScreenWidth/2-buttonWidth/2, float64(g.y))
	screen.DrawImage(gopher, opg)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%f", ebiten.ActualFPS()))
	ebitenutil.DebugPrint(screen, "\nThe Go gopher was designed by Renée French.")
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

type Point struct {
	x, y int
}

func newImage(data []byte, w, h int) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
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

func cacheDiffBitmap(bg, cr graycache, w, h int) map[Point]bool {
	mp := make(map[Point]bool)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if isOverThreshold(cr[y*w+x]-bg[y*w+x], 30) {
				mp[Point{x, y}] = true
			}
		}
	}
	return mp
}

func isOverThreshold(diff, t int) bool {
	return diff > t
}
