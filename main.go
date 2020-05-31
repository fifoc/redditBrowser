package main

import (
	"fmt"
	"github.com/nfnt/resize"
	image2 "image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
)

var error []byte
var cacheLock sync.Mutex
var cache map[string][]byte = make(map[string][]byte)

func splitColor(color uint32) (uint32, uint32, uint32) {
	return color >> 16 & 0xFF, color >> 8 & 0xFF, color & 0xFF
}

func colorDelta(c uint32, d uint32) int64 {
	r, g, b := splitColor(c)
	tR, tG, tB := splitColor(d)

	factorR := int64(int(tR) - int(r))
	factorG := int64(int(tG) - int(g))
	factorB := int64(int(tB) - int(b))
	delta := (factorR * factorR) + (factorG * factorG) + (factorB * factorB)

	return delta
}

var palette []uint32
var lumaCache map[uint32]uint64 = make(map[uint32]uint64)
var paletteSet map[uint32]bool

func CalculateColorLuma(col uint32) uint64 {
	rr, gg, bb := splitColor(col)
	r, g, b := float64(rr), float64(gg), float64(bb)

	r = math.Pow(r, 2)
	g = math.Pow(g, 2)
	b = math.Pow(b, 2)

	r = .299 * r
	g = .587 * g
	b = .114 * b

	importance := math.Sqrt(r + g + b)
	importance = importance * 1000

	return uint64(importance) + 1000
}

var FIF_altMode = true

func a(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		return
	}
	a := req.URL.Query()
	if _, ok := a["frog"]; ok == true {
		cacheLock.Lock()
		if entry, ok := cache[a["frog"][0]]; ok == true {
			cacheLock.Unlock()
			fmt.Println("Served from cache")
			w.Write(entry)
			return
		} else {
			cacheLock.Unlock()
			resp, err := http.Get("https://i.redd.it/" + a["frog"][0])
			if err != nil {
				w.Write(error)
				return
			}

			defer resp.Body.Close()

			image, _, err := image2.Decode(resp.Body)
			if err != nil {
				w.Write(error)
				return
			}

			newImage := resize.Thumbnail(320, 200, image, resize.Lanczos3)
			www := newImage.Bounds().Size()
			z := encodeFif((www.X / 2) * 2, (www.Y / 4) * 4, newImage, false)
			cacheLock.Lock()
			cache[a["frog"][0]] = z
			cacheLock.Unlock()
			if len(cache) > 100 {
				cache = make(map[string][]byte)
			}
			fmt.Println("Encoded fif!")
			w.Write(z)
			return
		}
		w.Write(error)
	} else {
		w.Write(error)
	}
}

func main() {
	fmt.Println("Generating OC palette...")
	palette, _ = generatePalette()

	fmt.Println("Calculating color importances...")
	for i := 0; i < len(palette); i++ {
		lumaCache[palette[i]] = CalculateColorLuma(palette[i])
	}

	http.HandleFunc("/getfif", a)
	error, _ = ioutil.ReadFile("error.fif")

	http.ListenAndServe(":8090", nil)
}
