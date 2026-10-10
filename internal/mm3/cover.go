package mm3

import "image"

// Cover is the options menu's background (control.raw) in the VGA driver's game palette:
// goplay's floppy label. GameDir must be set.
func Cover() image.Image {
	if ccMain == nil {
		openCCArchive()
	}
	pal := ccLoad("vga")[offPalette:][:768]
	img := image.NewPaletted(image.Rect(0, 0, 320, 200), nil)
	for i := 0; i < 256; i++ {
		r, g, b := pal[3*i]<<2|pal[3*i]>>4, pal[3*i+1]<<2|pal[3*i+1]>>4, pal[3*i+2]<<2|pal[3*i+2]>>4
		img.Palette = append(img.Palette, rgb{r, g, b})
	}
	copy(img.Pix, ccLoad("control.raw"))
	return img
}

type rgb struct{ r, g, b uint8 }

func (c rgb) RGBA() (r, g, b, a uint32) {
	return uint32(c.r) * 0x101, uint32(c.g) * 0x101, uint32(c.b) * 0x101, 0xffff
}
