// +build k8s

package axe

import (
	"github.com/gdamore/tcell"
	"github.com/rancher/axe/axe/action"
	"github.com/rancher/axe/axe/datafeeder"
	"github.com/rancher/axe/axe/k8s"
	"github.com/rancher/axe/axe/status"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	k8sKind = "kubernetes"
)

var (
	RootPage = k8sKind

	k8sResourceKind = ResourceKind{
		Title: "K8s",
		Kind:  "k8s",
	}

	PageNav = map[rune]string{
		'1': k8sKind,
	}

	Footers = []resourceView{
		{
			Title: "Kubernetes",
			Kind:  k8sKind,
			Index: 1,
		},
	}

	Shortcuts = [][]string{
		{"Key g", "Get"},
		{"Key e", "Edit"},
		{"Key d", "Delete"},
	}

	ViewMap = map[string]View{
		k8sKind: {
			Actions: []action.Action{
				{
					Name:        "get",
					Shortcut:    'g',
					Description: "get a resource",
				},
				{
					Name:        "edit",
					Shortcut:    'e',
					Description: "edit a resource",
				},
				{
					Name:        "delete",
					Shortcut:    'd',
					Description: "delete a resource",
				},
			},
			Kind:   k8sResourceKind,
			Feeder: datafeeder.NewDataFeeder(k8s.RefreshResourceKind),
		},
	}

	// Root page event handler, only have menu and navigate back
	RootEventHandler = func(app *AppView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyRune:
				if event.Rune() == 'm' {
					if !app.showMenu {
						newpage := tview.NewPages().AddPage("menu", app.CurrentPage(), true, true).
							AddPage("menu-decor", center(app.menuView, 60, 15), true, true)
						app.content.SwitchPage(app.currentPage, newpage)
						app.showMenu = true
					} else {
						app.SwitchPage(app.currentPage, app.CurrentPage())
						app.showMenu = false
					}
				}
			case tcell.KeyEscape:
				app.LastPage()
			}
			return event
		}
	}

	tableEventHandler = func(h status.GenericDrawer) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				if err := ResourceView(h); err != nil {
					h.UpdateStatus(err.Error(), true)
				}
			}
			return event
		}
	}

	itemEventHandler = func(h status.GenericDrawer) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			// todo:
			}
			return event
		}
	}
)

func ResourceView(h status.GenericDrawer) error {
	viewResource(h)
	return nil
}

func viewResource(h status.GenericDrawer) {
	table := h.GetTable()
	row, _ := table.GetSelection()
	kind := table.GetCell(row, 0).Text
	groupVersion := table.GetCell(row, 1).Text

	var apiResource metav1.APIResource
	group, version := kv.Split(groupVersion, "/")
	if version == "" {
		version = group
		group = ""
	}
	apiResource.Group = group
	apiResource.Version = version
	apiResource.Name = kind

	rkind := ResourceKind{
		Title: apiResource.Name,
		Kind:  kind,
	}

	w := k8s.Wrapper{
		Group:   apiResource.Group,
		Version: apiResource.Version,
		Name:    apiResource.Name,
	}

	feeder := datafeeder.NewDataFeeder(w.RefreshResource)
	newTable := NewTableViewWithArgs(h.(*tableView).app, rkind, feeder, nil, itemEventHandler)
	newTable.app.tableViews[w.Name] = newTable
	h.SwitchPage(w.Name, newTable)
}
