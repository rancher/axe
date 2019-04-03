package status

import (
	"github.com/rancher/axe/axe/action"
	"github.com/rivo/tview"
	"k8s.io/client-go/kubernetes"
)

type GenericDrawer interface {
	UpdateStatus(status string, isError bool) tview.Primitive
	InsertDialog(name string, page tview.Primitive, dialog tview.Primitive)
	GetCurrentPrimitive() tview.Primitive
	GetSelectionName() string
	GetClientSet() *kubernetes.Clientset
	GetResourceKind() string
	GetCurrentPage() string
	GetApplication() *tview.Application
	GetAction() []action.Action
	GetTable() *tview.Table
	SwitchPage(page string, draw tview.Primitive)
	SwitchToRootPage()
}
