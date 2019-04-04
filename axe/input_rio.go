// +build rio

package axe

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
	"github.com/gdamore/tcell"
	"github.com/rancher/axe/axe/action"
	"github.com/rancher/axe/axe/datafeeder"
	"github.com/rancher/axe/axe/rio"
	"github.com/rancher/axe/axe/status"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	serviceKind         = "services"
	routeKind           = "routes"
	podKind             = "pods"
	externalServiceKind = "externalservices"

	stackLabel   = "rio.cattle.io/stack"
	serviceLabel = "rio.cattle.io/service"

	defaultStyle = "native"
)

func init() {
	for style := range styles.Registry {
		colorStyles = append(colorStyles, style)
	}
	sort.Strings(colorStyles)
}

var (
	defaultBackGroundColor = tcell.ColorBlack

	colorStyles []string

	RootPage = serviceKind

	Shortcuts = [][]string{
		{"Key i", "Inspect"},
		{"Key e", "Edit"},
		{"Key l", "Logs"},
		{"Key x", "Exec"},
		{"Key n", "Create"},
		{"Key d", "Delete"},
		{"Key r", "Refresh"},
		{"Key /", "Search"},
		{"Key p", "View Pods"},
	}

	Footers = []resourceView{
		{
			Title: "Services",
			Kind:  serviceKind,
			Index: 1,
		},
		{
			Title: "Routes",
			Kind:  routeKind,
			Index: 2,
		},
	}

	PageNav = map[rune]string{
		'1': serviceKind,
		'2': routeKind,
	}

	tableEventHandler = func(h status.GenericDrawer) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				ActionView(h)
			case tcell.KeyRune:
				switch event.Rune() {
				case 'i':
					Inspect("yaml", defaultStyle, false, h)
				case 'l':
					Logs("", h)
				case 'x':
					Exec("", h)
				case 'd':
					Rm(h)
				case 'r':
					h.Refresh()
				case '/':
					h.ShowSearch()
				case 'm':
					h.ShowMenu()
				default:
					h.Navigate(event.Rune())
				case 'p':
					ViewPods(h)
				}
			}
			return event
		}
	}



	Route = ResourceKind{
		Title: "Routes",
		Kind:  "route",
	}

	Service = ResourceKind{
		Title: "Services",
		Kind:  "service",
	}

	ExternalService = ResourceKind{
		Kind:  "ExternalService",
		Title: "externalservice",
	}

	DefaultAction = []action.Action{
		{
			Name:        "inspect",
			Shortcut:    'i',
			Description: "inspect a resource",
		},
		{
			Name:        "edit",
			Shortcut:    'e',
			Description: "edit a resource",
		},
		{
			Name:        "create",
			Shortcut:    'c',
			Description: "create a resource",
		},
		{
			Name:        "delete",
			Shortcut:    'd',
			Description: "delete a resource",
		},
	}

	execAndlog = []action.Action{
		{
			Name:        "exec",
			Shortcut:    'x',
			Description: "exec into a container or service",
		},
		{
			Name:        "log",
			Shortcut:    'l',
			Description: "view logs of a service",
		},
	}

	ViewMap = map[string]View{
		serviceKind: {
			Actions: append(
				DefaultAction,
				append(
					execAndlog,
					action.Action{
						Name:        "pods",
						Shortcut:    'p',
						Description: "view pods of a service",
					})...,
			),
			Kind:   Service,
			Feeder: datafeeder.NewDataFeeder(rio.ServiceRefresher),
		},
		routeKind: {
			Actions: DefaultAction,
			Kind:    Route,
			Feeder:  datafeeder.NewDataFeeder(rio.RouteRefresher),
		},
		externalServiceKind: {
			Actions: DefaultAction,
			Kind:    Route,
			Feeder:  datafeeder.NewDataFeeder(rio.RouteRefresher),
		},
	}
)

func ActionView(h status.GenericDrawer) {
	list := newSelectionList()
	list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
		switch s {
		case "inspect":
			formatList := newSelectionList()
			formatList.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				if s2 == "json format" || s2 == "yaml format" {
					Inspect(s, defaultStyle, false, h)
				} else {
					colorList := newSelectionList()
					for _, s := range colorStyles {
						colorList.AddItem(s, "", ' ', nil)
					}
					colorList.SetSelectedFunc(func(i int, style string, s2 string, r rune) {
						Inspect(s, style, true, h)
					})

					h.InsertDialog("inspect-color", h.GetCurrentPrimitive(), colorList)
				}
			})
			formatList.AddItem("yaml", "yaml format", 'y', nil)
			formatList.AddItem("yaml", "yaml with different color styles", 'u', nil)
			formatList.AddItem("json", "json format", 'j', nil)
			formatList.AddItem("json", "json with different color styles", 'k', nil)
			h.InsertDialog("inspect-format", h.GetCurrentPrimitive(), formatList)
		case "log":
			list := newContainerSelectionList(h)
			list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				Logs(s, h)
			})
			h.InsertDialog("logs-containers", h.GetCurrentPrimitive(), list)
		case "exec":
			list := newContainerSelectionList(h)
			list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				Exec(s, h)
			})
			h.InsertDialog("exec-containers", h.GetCurrentPrimitive(), list)
		case "pods":
			ViewPods(h)
		case "delete":
			Rm(h)
		}
	})

	for _, a := range h.GetAction() {
		list.AddItem(a.Name, a.Description, a.Shortcut, nil)
	}
	h.InsertDialog("option", h.GetCurrentPrimitive(), list)
}

func newSelectionList() *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(tcell.ColorGray)
	list.SetMainTextColor(tcell.ColorBlack)
	list.SetSecondaryTextColor(tcell.ColorPurple)
	list.SetShortcutColor(defaultBackGroundColor)
	return list
}

func newContainerSelectionList(h status.GenericDrawer) *tview.List {
	list := newSelectionList()
	containers, err := listRioContainer(h.GetSelectionName(), h.GetClientSet())
	if err != nil {
		h.UpdateStatus(err.Error(), true)
		return nil
	}
	for _, c := range containers {
		list.AddItem(c, "", ' ', nil)
	}
	return list
}

func listRioContainer(name string, clientset *kubernetes.Clientset) ([]string, error) {
	stackName, serviceName := kv.Split(name, "/")
	if serviceName == "" {
		serviceName = stackName
		stackName = "default"
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", stackLabel, stackName, serviceLabel, serviceName),
	})
	if err != nil {
		return nil, err
	}
	var containers []string
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			containers = append(containers, container.Name)
		}
		for _, container := range pod.Spec.InitContainers {
			containers = append(containers, container.Name)
		}
	}

	return containers, nil

}

/*
	general rio operation(inspect, edit, exec, logs, create)
*/
func Inspect(format, style string, colorful bool, h status.GenericDrawer) {
	name := h.GetSelectionName()
	outBuffer := &strings.Builder{}
	errBuffer := &strings.Builder{}
	args := []string{"inspect", "-t", h.GetResourceKind(), name}
	cmd := exec.Command("rio", args...)
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer
	if err := cmd.Run(); err != nil {
		h.UpdateStatus(errBuffer.String(), true)
		return
	}

	inspectBox := tview.NewTextView()
	if colorful {
		inspectBox.SetDynamicColors(true).SetBackgroundColor(tcell.Color(styles.Registry[style].Get(chroma.Background).Background))
		writer := tview.ANSIWriter(inspectBox)
		if err := quick.Highlight(writer, outBuffer.String(), format, "terminal256", style); err != nil {
			h.UpdateStatus(err.Error(), true)
			return
		}
	} else {
		inspectBox.SetDynamicColors(true).SetBackgroundColor(defaultBackGroundColor)
		inspectBox.SetText(outBuffer.String())
	}

	newpage := tview.NewPages().AddPage("inspect", inspectBox, true, true)
	h.SwitchPage(h.GetCurrentPage(), newpage)
}

func Logs(container string, h status.GenericDrawer) {
	name := h.GetSelectionName()
	args := []string{"logs", "-f"}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, name)
	cmd := exec.Command("rio", args...)

	logbox := tview.NewTextView()
	{
		logbox.SetTitle(fmt.Sprintf("Logs - (%s)", name))
		logbox.SetBorder(true)
		logbox.SetTitleColor(tcell.ColorPurple)
		logbox.SetDynamicColors(true)
		logbox.SetBackgroundColor(tcell.ColorBlack)
		logbox.SetChangedFunc(func() {
			logbox.ScrollToEnd()
			h.GetApplication().Draw()
		})
		logbox.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				cmd.Process.Kill()
			}
		})
	}

	cmd.Stdout = tview.ANSIWriter(logbox)
	go func() {
		if err := cmd.Run(); err != nil {
			return
		}
	}()

	newpage := tview.NewPages().AddPage("logs", logbox, true, true)
	h.SwitchPage(h.GetCurrentPage(), newpage)
}

func Exec(container string, h status.GenericDrawer) {
	name := h.GetSelectionName()
	shellArgs := []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c /bin/bash /dev/null || exec /bin/bash) || exec /bin/sh"}
	args := []string{"exec", "-it"}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, name)
	cmd := exec.Command("rio", append(args, shellArgs...)...)
	errorBuffer := &strings.Builder{}
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, errorBuffer, os.Stdin

	h.GetApplication().Suspend(func() {
		clearScreen()
		if err := cmd.Run(); err != nil {
			h.UpdateStatus(errorBuffer.String(), true)
		}
		return
	})
}

func Rm(h status.GenericDrawer) {
	name := h.GetSelectionName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to delete %s %s?", h.GetResourceKind(), name)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				args := []string{"rm", "-t", h.GetResourceKind(), name}
				cmd := exec.Command("rio", args...)
				if err := cmd.Run(); err != nil {
					h.UpdateStatus(err.Error(), true)
					return
				}
				h.SwitchToRootPage()
			} else if buttonLabel == "Cancel" {
				h.BackPage()
			}
		})
	h.InsertDialog("delete", h.GetCurrentPrimitive(), modal)
}

func ViewPods(h status.GenericDrawer) {
	name := h.GetSelectionName()
	app := h.(*tableView).app

	rkind := ResourceKind{
		Kind:  podKind,
		Title: "Pods",
	}

	var podEventHandler = func(h status.GenericDrawer) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				ActionView(h)
			case tcell.KeyRune:
				switch event.Rune() {
				case 'i':
					Inspect("yaml", defaultStyle, false, h)
				case 'l':
					Logs("", h)
				case 'x':
					Exec("", h)
				case 'r':
					h.Refresh()
				case '/':
					h.ShowSearch()
				case 'm':
					h.ShowMenu()
				default:
					h.Navigate(event.Rune())
				}
			}
			return event
		}
	}

	feeder := datafeeder.NewDataFeeder(rio.PodRefresher(name))
	table := NewTableViewWithArgs(app, rkind, feeder, DefaultAction, podEventHandler)
	h.SwitchPage(h.GetCurrentPage(), table)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
