// +build k8s

package axe

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rancher/axe/axe/action"
	"github.com/rancher/axe/axe/datafeeder"
	"github.com/rancher/axe/axe/k8s"
	"github.com/rancher/axe/axe/status"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
	"strings"
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
						app.SwitchPage(app.currentPage, newpage)
						app.showMenu = true
					} else {
						app.SwitchPage(app.currentPage, app.CurrentPage())
						app.showMenu = false
					}
				}
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
			case
			}
			return event
		}
	}

	itemEventHandler = func(h status.GenericDrawer) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'g':
				Get(h)
			case 'e':
				Edit(h)
			case 'd':
				Delete(h)
			}
			return event
		}
	}
)

func getNamespaceAndName(h status.GenericDrawer) (string, string) {
	table := h.GetTable()
	namespaced := false
	if strings.Contains(table.GetCell(0, 0).Text, "NAMESPACE") {
		namespaced = true
	}

	row, _ := table.GetSelection()
	var namespace, name string
	if namespaced {
		namespace, name = table.GetCell(row, 0).Text, table.GetCell(row, 1).Text
	} else {
		name = table.GetCell(row, 0).Text
	}
	return namespace, name
}

func Get(h status.GenericDrawer) {
	namespace, name := getNamespaceAndName(h)
	out := &strings.Builder{}
	errB := &strings.Builder{}
	var args []string
	if namespace != "" {
		args = []string{"get", h.GetResourceKind(), "-n", namespace, name, "-o", "yaml"}
	} else {
		args = []string{"get", h.GetResourceKind(), name, "-o", "yaml"}
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout, cmd.Stderr = out, errB
	if err := cmd.Run(); err != nil {
		h.UpdateStatus(errB.String(), true)
		return
	}
	box := tview.NewTextView()
	box.SetDynamicColors(true).SetBackgroundColor(tcell.ColorBlack)
	box.SetText(out.String())
	newpage := tview.NewPages().AddPage("get", box, true, true)
	h.SwitchPage(h.GetCurrentPage(), newpage)
}

func Edit(h status.GenericDrawer) {
	namespace, name := getNamespaceAndName(h)
	errb := &strings.Builder{}
	var args []string
	if namespace != "" {
		args = []string{"edit", h.GetResourceKind(),"-n", namespace, name}
	} else {
		args = []string{"edit", h.GetResourceKind(), name}
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, errb

	h.GetApplication().Suspend(func() {
		clearScreen()
		if err := cmd.Run(); err != nil {
			h.UpdateStatus(errb.String(), true)
		}
		return
	})
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func Delete(h status.GenericDrawer) {
	namespace, name := getNamespaceAndName(h)
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to delete %s %s?", h.GetResourceKind(), name)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				var args []string
				if namespace != "" {
					args = []string{"delete", h.GetResourceKind(), "-n", namespace, name}
				} else {
					args = []string{"delete", h.GetResourceKind(), name}
				}
				cmd := exec.Command("rio", args...)
				errB := &strings.Builder{}
				cmd.Stderr = errB
				if err := cmd.Run(); err != nil {
					h.UpdateStatus(errB.String(), true)
					return
				}
				h.SwitchToRootPage()
			} else if buttonLabel == "Cancel" {
				h.BackPage()
			}
		})
	h.InsertDialog("delete", h.GetCurrentPrimitive(), modal)
}

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
	h.SwitchPage(h.GetCurrentPage(), newTable)
}
