// +build rio

package axe

import (
	"github.com/gdamore/tcell"
	"github.com/rancher/axe/axe/action"
	"github.com/rancher/axe/axe/datafeeder"
	"github.com/rancher/axe/axe/rio"
	"github.com/rancher/axe/axe/status"
	"github.com/rivo/tview"
)

const (
	serviceKind         = "services"
	routeKind           = "routes"
	externalServiceKind = "externalservices"

	defaultStyle = "native"
)

var (
	RootPage = serviceKind

	Shortcuts = [][]string{
		{"Key i", "Inspect"},
		{"Key e", "Edit"},
		{"Key l", "Logs"},
		{"Key x", "Exec"},
		{"Key n", "Create"},
		{"Key d", "Delete"},
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

	// Root page handler to handler page nav and menu view
	RootEventHandler = func(app *AppView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyRune:
				if kind, ok := PageNav[event.Rune()]; ok {
					app.footerView.TextView.Highlight(kind).ScrollToHighlight()
					if app.tableViews[kind] == nil {
						app.tableViews[kind] = NewTableView(app, kind, tableEventHandler)
					}
					app.content.SwitchPage(kind, app.tableViews[kind])
				} else if event.Rune() == 'm' {
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
				rio.ActionView(h)
			case tcell.KeyRune:
				switch event.Rune() {
				case 'i':
					rio.Inspect("yaml", defaultStyle, false, h)
				case 'l':
					rio.Logs("", h)
				case 'x':
					rio.Exec("", h)
				case 'd':
					rio.Rm(h)
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

	ViewMap = map[string]View{
		serviceKind: {
			Actions: append(DefaultAction,
				action.Action{
					Name:        "exec",
					Shortcut:    'x',
					Description: "exec into a container or service",
				},
				action.Action{
					Name:        "log",
					Shortcut:    'l',
					Description: "view logs of a service",
				}),
			Kind:      Service,
			Feeder: datafeeder.NewDataFeeder(rio.ServiceRefresher),
		},
		routeKind: {
			Actions:   DefaultAction,
			Kind:      Route,
			Feeder: datafeeder.NewDataFeeder(rio.RouteRefresher),
		},
		externalServiceKind: {
			Actions:   DefaultAction,
			Kind:      Route,
			Feeder: datafeeder.NewDataFeeder(rio.RouteRefresher),
		},
	}
)
