package axe

import (
	"bytes"

	"github.com/rancher/axe/axe/action"
	"github.com/rancher/axe/axe/datafeeder"
	"github.com/rivo/tview"
)

var Kubeconfig string

type View struct {
	Actions   []action.Action
	Kind      ResourceKind
	Feeder    datafeeder.DataSource
}

type resourceView struct {
	tview.Primitive
	Title string
	Kind  string
	Index int
}

type ResourceKind struct {
	Title string
	Kind  string
}

type Refresher func(b *bytes.Buffer) error
