//go:build !js

package host

import "image"

// HardwareCursor: Ebitengine has no custom cursor image on the desktop, so there the game
// draws its pointer into the picture.
const HardwareCursor = false

func setCursor(image.Image, int, int) {}
