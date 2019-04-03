package rio

import (
	"sort"

	"github.com/alecthomas/chroma/styles"
	"github.com/gdamore/tcell"
)

var defaultBackGroundColor = tcell.ColorBlack

var colorStyles []string

func init() {
	for style := range styles.Registry {
		colorStyles = append(colorStyles, style)
	}
	sort.Strings(colorStyles)
}
