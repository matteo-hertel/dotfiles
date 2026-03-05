package palette

import (
	"fmt"
	"math"
)

// Generate creates a Theme from a name, background, foreground, and accent hex color.
// It produces a full 16-color ANSI palette derived from the given colors.
func Generate(name, bg, fg, accent string) (*Theme, error) {
	bgR, bgG, bgB, err := ParseHex(bg)
	if err != nil {
		return nil, fmt.Errorf("invalid background: %w", err)
	}
	fgR, fgG, fgB, err := ParseHex(fg)
	if err != nil {
		return nil, fmt.Errorf("invalid foreground: %w", err)
	}
	_, _, _, err = ParseHex(accent)
	if err != nil {
		return nil, fmt.Errorf("invalid accent: %w", err)
	}

	_, _, bgL := rgbToHSL(bgR, bgG, bgB)
	acR, acG, acB, _ := ParseHex(accent)
	acH, acS, _ := rgbToHSL(acR, acG, acB)

	dark := bgL < 0.5

	// Base saturation from accent, clamped to a reasonable range
	baseSat := acS
	if baseSat < 0.4 {
		baseSat = 0.4
	}
	if baseSat > 0.9 {
		baseSat = 0.9
	}

	// Six ANSI hues: red, green, yellow, blue, magenta, cyan
	// Blue uses accent hue if it's in a blue-ish range, otherwise default 220
	blueHue := acH
	if blueHue < 180 || blueHue > 270 {
		blueHue = 220
	}
	hues := [6]float64{0, 120, 60, blueHue, 300, 180}

	var normalL, brightL float64
	if dark {
		normalL = 0.55
		brightL = 0.70
	} else {
		normalL = 0.40
		brightL = 0.30
	}

	// color0: slightly lighter than bg (dark) or slightly darker (light)
	var c0L float64
	if dark {
		c0L = bgL + 0.08
	} else {
		c0L = bgL - 0.08
	}
	c0L = clamp(c0L)

	_, bgSatHSL, _ := rgbToHSL(bgR, bgG, bgB)
	color0 := hslToHex(0, bgSatHSL, c0L)

	// colors 1-6: the six hues at normal lightness
	var normalColors [6]string
	for i, h := range hues {
		normalColors[i] = hslToHex(h, baseSat, normalL)
	}

	// color7: near-foreground shade
	_, fgS, fgL := rgbToHSL(fgR, fgG, fgB)
	var c7L float64
	if dark {
		c7L = fgL - 0.10
	} else {
		c7L = fgL + 0.10
	}
	c7L = clamp(c7L)
	color7 := hslToHex(0, fgS*0.3, c7L)

	// color8: brighter than color0
	var c8L float64
	if dark {
		c8L = c0L + 0.10
	} else {
		c8L = c0L - 0.10
	}
	c8L = clamp(c8L)
	color8 := hslToHex(0, bgSatHSL, c8L)

	// colors 9-14: the six hues at bright lightness
	var brightColors [6]string
	for i, h := range hues {
		brightColors[i] = hslToHex(h, baseSat, brightL)
	}

	// color15: foreground color
	color15 := ToHex(fgR, fgG, fgB)

	theme := &Theme{
		Name:       name,
		Background: bg,
		Foreground: fg,
		Cursor:     fg,
		Colors: [16]string{
			color0,
			normalColors[0], normalColors[1], normalColors[2],
			normalColors[3], normalColors[4], normalColors[5],
			color7,
			color8,
			brightColors[0], brightColors[1], brightColors[2],
			brightColors[3], brightColors[4], brightColors[5],
			color15,
		},
	}

	return theme, nil
}

// rgbToHSL converts RGB (0-255) to HSL (h: 0-360, s: 0-1, l: 0-1).
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	l = (max + min) / 2.0

	if max == min {
		h = 0
		s = 0
		return
	}

	d := max - min
	if l > 0.5 {
		s = d / (2.0 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	case bf:
		h = (rf-gf)/d + 4
	}
	h *= 60

	return
}

// hslToRGB converts HSL (h: 0-360, s: 0-1, l: 0-1) to RGB (0-255).
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(math.Round(l * 255))
		return v, v, v
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	hNorm := h / 360.0

	r := hueToRGB(p, q, hNorm+1.0/3.0)
	g := hueToRGB(p, q, hNorm)
	b := hueToRGB(p, q, hNorm-1.0/3.0)

	return uint8(math.Round(r * 255)), uint8(math.Round(g * 255)), uint8(math.Round(b * 255))
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// hslToHex converts HSL values to a hex color string.
func hslToHex(h, s, l float64) string {
	r, g, b := hslToRGB(h, s, l)
	return ToHex(r, g, b)
}

// clamp restricts a value to the range [0, 1].
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
