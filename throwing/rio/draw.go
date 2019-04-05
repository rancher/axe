package rio

import (
	"github.com/gdamore/tcell"
	"github.com/rancher/axe/throwing"
	"github.com/rancher/axe/throwing/datafeeder"
	"github.com/rancher/axe/throwing/types"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

var (
	defaultBackGroundColor = tcell.ColorBlack

	colorStyles []string

	RootPage = serviceKind

	Shortcuts = [][]string{
		{"Key i", "inspect"},
		{"Key e", "Edit"},
		{"Key l", "logs"},
		{"Key x", "execute"},
		{"Key n", "Create"},
		{"Key d", "Delete"},
		{"Key r", "Refresh"},
		{"Key /", "Search"},
		{"Key p", "View Pods"},
	}

	Footers = []types.ResourceView{
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

	tableEventHandler = func(t *throwing.TableView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				actionView(t)
			case tcell.KeyRune:
				switch event.Rune() {
				case 'i':
					inspect("yaml", defaultStyle, false, t)
				case 'l':
					logs("", t)
				case 'x':
					execute("", t)
				case 'd':
					rm(t)
				case 'r':
					t.Refresh()
				case '/':
					t.ShowSearch()
				case 'm':
					t.ShowMenu()
				default:
					t.Navigate(event.Rune())
				case 'p':
					viewPods(t)
				}
			}
			return event
		}
	}

	Route = types.ResourceKind{
		Title: "Routes",
		Kind:  "route",
	}

	Service = types.ResourceKind{
		Title: "Services",
		Kind:  "service",
	}

	ExternalService = types.ResourceKind{
		Kind:  "ExternalService",
		Title: "externalservice",
	}

	DefaultAction = []types.Action{
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

	execAndlog = []types.Action{
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

	ViewMap = map[string]types.View{
		serviceKind: {
			Actions: append(
				DefaultAction,
				append(
					execAndlog,
					types.Action{
						Name:        "pods",
						Shortcut:    'p',
						Description: "view pods of a service",
					})...,
			),
			Kind:   Service,
			Feeder: datafeeder.NewDataFeeder(ServiceRefresher),
		},
		routeKind: {
			Actions: DefaultAction,
			Kind:    Route,
			Feeder:  datafeeder.NewDataFeeder(RouteRefresher),
		},
		externalServiceKind: {
			Actions: DefaultAction,
			Kind:    Route,
			Feeder:  datafeeder.NewDataFeeder(RouteRefresher),
		},
	}

	drawer = types.Drawer{
		RootPage:  RootPage,
		Shortcuts: Shortcuts,
		ViewMap:   ViewMap,
		PageNav:   PageNav,
		Footers:   Footers,
	}
)

func Start(c *cli.Context) error {
	kubeconfig := c.String("kubeconfig")

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	clientset := kubernetes.NewForConfigOrDie(restConfig)

	app := throwing.NewAppView(clientset, drawer, tableEventHandler)
	if err := app.Init(); err != nil {
		return err
	}
	return app.Run()
}
