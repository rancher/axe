package k8s

import (
	"os"

	"github.com/gdamore/tcell"
	"github.com/rancher/axe/throwing"
	"github.com/rancher/axe/throwing/datafeeder"
	"github.com/rancher/axe/throwing/types"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sKind = "kubernetes"

	RootPage = k8sKind

	k8sResourceKind = types.ResourceKind{
		Title: "K8s",
		Kind:  "k8s",
	}

	PageNav = map[rune]string{
		'1': k8sKind,
	}

	Footers = []types.ResourceView{
		{
			Title: "Kubernetes",
			Kind:  k8sKind,
			Index: 1,
		},
	}

	Shortcuts = [][]string{
		{"Key g", "get"},
		{"Key e", "edit"},
		{"Key d", "delete"},
	}

	ViewMap = map[string]types.View{
		k8sKind: {
			Actions: []types.Action{
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
			Feeder: datafeeder.NewDataFeeder(RefreshResourceKind),
		},
	}

	tableEventHandler = func(t *throwing.TableView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				if err := resourceView(t); err != nil {
					t.UpdateStatus(err.Error(), true)
				}
			case tcell.KeyRune:
				if event.Rune() == 'm' {
					t.ShowMenu()
				}
			}
			return event
		}
	}

	itemEventHandler = func(t *throwing.TableView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'g':
				get(t)
			case 'e':
				edit(t)
			case 'd':
				delete(t)
			case 'm':
				t.ShowMenu()
			}
			return event
		}
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
	os.Setenv("KUBECONFIG", kubeconfig)

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
