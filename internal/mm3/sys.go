package mm3

import (
	"fmt"
	"os"
)

// Port of sys.asm's own routines (its video-driver thunks are vga.go).

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "mm3:", msg)
	os.Exit(1)
}

// fatalDiskError: the original prints its disk error message and exits to DOS.
func fatalDiskError() { fatal("disk error (MM3.CC or MM3.CUR missing or damaged)") }
