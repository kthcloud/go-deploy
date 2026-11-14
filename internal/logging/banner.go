package logging

import (
	"fmt"
	"io"
)

const (
	logo = `   __   __  __       __             __
  / /__/ /_/ /  ____/ /__  __ _____/ /
 /  '_/ __/ _ \/ __/ / _ \/ // / _  / 
/_/\_\\__/_//_/\__/_/\___/\_,_/\_,_/  `
)

func Banner(w io.Writer, version string) error {
	_, err := fmt.Fprintf(w, "%s\nVersion:\t%s\n", logo, version)
	return err
}
