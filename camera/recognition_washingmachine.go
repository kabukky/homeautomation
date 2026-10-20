package camera

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"sort"
)

// ---------------------------------------------------------------------------
// Calibration — measured once from a frame of YOUR camera.
// Run with -debug and check that every marker sits on its segment.
// ---------------------------------------------------------------------------

// Origin of each digit cell: where the left vertical segments and the top
// horizontal segment would meet. Order: hours, minutes tens, minutes ones.
var digitOrigins = [3][2]float64{
	{637, 159.5}, // hours
	{660, 155.5}, // minutes, tens
	{678, 151.5}, // minutes, ones
}

// across: from the left vertical segments to the right ones.
// down:   from the top horizontal segment to the bottom one.
// The -1.5 compensates for the camera being slightly rotated.
var (
	across = [2]float64{12, -1.5}
	down   = [2]float64{0, 22}
)

// Area around the digits, used to measure the background level and as the
// crop for the debug image.
var displayRegion = image.Rect(620, 140, 700, 190)

const (
	maxShift    = 4    // px; camera drift the alignment search can absorb
	threshFrac  = 0.3  // lit if brighter than bg + 30% of (brightest segment - bg)
	minContrast = 40.0 // below this, the display is off or the image is unusable
	minMargin   = 0.15 // reject readings where lit and unlit segments are too close
)

// ---------------------------------------------------------------------------

// Segment positions in cell coordinates (0..1 across, 0..1 down).
// Bit order: a=top, b=upper right, c=lower right, d=bottom,
// e=lower left, f=upper left, g=middle.
var segPos = [7][2]float64{
	{0.5, 0}, {1, 0.25}, {1, 0.75}, {0.5, 1}, {0, 0.75}, {0, 0.25}, {0.5, 0.5},
}

var patterns = map[uint8]int{
	0b0111111: 0, 0b0000110: 1, 0b1011011: 2, 0b1001111: 3, 0b1100110: 4,
	0b1101101: 5, 0b1111101: 6, 0b1111100: 6, 0b0000111: 7, 0b0100111: 7,
	0b1111111: 8, 0b1101111: 9, 0b1100111: 9,
}

type Reading struct {
	Hours, Minutes int
	Levels         [3][7]float64 // measured brightness per segment
	Lit            [3]uint8      // segment bit pattern per digit
	Background     float64
	Threshold      float64
	Margin         float64 // (dimmest lit - brightest unlit) / (peak - background)
	Score          float64 // mean lit level - mean unlit level
	ShiftX, ShiftY int     // camera drift found by the alignment search
}

func (r Reading) TotalMinutes() int { return r.Hours*60 + r.Minutes }
func (r Reading) String() string    { return fmt.Sprintf("%d:%02d", r.Hours, r.Minutes) }

// red returns the red channel (0..255). The LEDs are orange-red, so this
// channel has by far the best contrast.
func red(img image.Image, x, y int) float64 {
	r, _, _, _ := img.At(x, y).RGBA()
	return float64(r >> 8)
}

// segmentLevel returns the 3x3 average around (x, y), which suppresses
// sensor noise.
func segmentLevel(img image.Image, x, y float64) float64 {
	cx, cy := int(math.Round(x)), int(math.Round(y))
	sum := 0.0
	for ky := -1; ky <= 1; ky++ {
		for kx := -1; kx <= 1; kx++ {
			sum += red(img, cx+kx, cy+ky)
		}
	}
	return sum / 9
}

func segmentPoint(digit, seg int) (float64, float64) {
	o, p := digitOrigins[digit], segPos[seg]
	return o[0] + p[0]*across[0] + p[1]*down[0],
		o[1] + p[0]*across[1] + p[1]*down[1]
}

func median(img image.Image, r image.Rectangle) float64 {
	var v []float64
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			v = append(v, red(img, x, y))
		}
	}
	sort.Float64s(v)
	return v[len(v)/2]
}

// ReadWashingMachine decodes the washing machine display. The camera may drift by a few pixels, so it
// tries every offset up to maxShift and keeps the valid reading whose
// sample points sit best on the segments (highest lit-vs-unlit contrast). It returns an error instead of
// guessing when no offset gives a clean, valid reading.
func ReadWashingMachine(img image.Image) (Reading, error) {
	if !displayRegion.Inset(-maxShift - 2).In(img.Bounds()) {
		return Reading{}, errors.New("image smaller than calibrated display region")
	}
	bg := median(img, displayRegion)

	var best Reading
	found := false
	for dy := -maxShift; dy <= maxShift; dy++ {
		for dx := -maxShift; dx <= maxShift; dx++ {
			rd, ok := decodeAt(img, bg, dx, dy)
			if ok && (!found || rd.Score > best.Score) {
				best, found = rd, true
			}
		}
	}
	if !found {
		return decodeAtFailed(img, bg)
	}
	if best.Margin < minMargin {
		return best, fmt.Errorf("reading %s is ambiguous (margin %.2f)", best, best.Margin)
	}
	return best, nil
}

// decodeAt samples all segments with the calibration shifted by (dx, dy).
func decodeAt(img image.Image, bg float64, dx, dy int) (Reading, bool) {
	rd := Reading{Background: bg, ShiftX: dx, ShiftY: dy}
	peak := 0.0
	for d := range digitOrigins {
		for s := range segPos {
			x, y := segmentPoint(d, s)
			rd.Levels[d][s] = segmentLevel(img, x+float64(dx), y+float64(dy))
			peak = math.Max(peak, rd.Levels[d][s])
		}
	}
	if peak-bg < minContrast {
		return rd, false
	}

	// Relative threshold: survives exposure changes and the fact that the
	// horizontal segments appear dimmer than the vertical ones.
	rd.Threshold = bg + threshFrac*(peak-bg)

	minLit, maxUnlit := math.Inf(1), bg
	var sumLit, sumUnlit float64
	var nLit int
	var val [3]int
	for d := range digitOrigins {
		for s := range segPos {
			if v := rd.Levels[d][s]; v > rd.Threshold {
				rd.Lit[d] |= 1 << s
				minLit = math.Min(minLit, v)
				sumLit += v
				nLit++
			} else {
				maxUnlit = math.Max(maxUnlit, v)
				sumUnlit += v
			}
		}
		n, ok := patterns[rd.Lit[d]]
		if !ok {
			return rd, false
		}
		val[d] = n
	}
	rd.Hours, rd.Minutes = val[0], val[1]*10+val[2]
	rd.Margin = (minLit - maxUnlit) / (peak - bg)
	rd.Score = sumLit/float64(nLit) - sumUnlit/float64(21-nLit)
	return rd, rd.Minutes <= 59 && nLit < 21
}

// decodeAtFailed explains why the unshifted calibration didn't decode.
func decodeAtFailed(img image.Image, bg float64) (Reading, error) {
	rd, _ := decodeAt(img, bg, 0, 0)
	if rd.Threshold == 0 {
		return rd, errors.New("display not visible (off, or image too dark)")
	}
	return rd, errors.New("no valid reading at any offset (check calibration with -debug)")
}

// writeDebug saves an enlarged crop of the display with a box around every
// 3x3 sample area: green = lit, blue = off. Use it to check the calibration.
func writeDebug(img image.Image, rd Reading, path string) error {
	const k = 10
	r := displayRegion
	out := image.NewRGBA(image.Rect(0, 0, r.Dx()*k, r.Dy()*k))
	for y := 0; y < r.Dy()*k; y++ {
		for x := 0; x < r.Dx()*k; x++ {
			out.Set(x, y, img.At(r.Min.X+x/k, r.Min.Y+y/k))
		}
	}
	for d := range digitOrigins {
		for s := range segPos {
			c := color.RGBA{60, 120, 255, 255}
			if rd.Lit[d]&(1<<s) != 0 {
				c = color.RGBA{0, 255, 80, 255}
			}
			x, y := segmentPoint(d, s)
			x += float64(rd.ShiftX)
			y += float64(rd.ShiftY)
			x0 := int((math.Round(x) - float64(r.Min.X) - 1) * k)
			y0 := int((math.Round(y) - float64(r.Min.Y) - 1) * k)
			size := 3 * k
			for i := 0; i < size; i++ {
				out.Set(x0+i, y0, c)
				out.Set(x0+i, y0+size-1, c)
				out.Set(x0, y0+i, c)
				out.Set(x0+size-1, y0+i, c)
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, out)
}
