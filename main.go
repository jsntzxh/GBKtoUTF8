package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	if runtime.GOOS != "windows" {
		fmt.Fprintf(os.Stderr, "%s\n", T(currentLang, "err.windows_only"))
		fmt.Fprintf(os.Stderr, "%s\n", T(currentLang, "err.windows_only_hint"))
		os.Exit(1)
	}

	runUI()
}
