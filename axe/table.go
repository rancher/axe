package axe

import (
	"fmt"
	"github.com/rancher/axe/axe/status"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rancher/axe/axe/action"
	"github.com/rancher/axe/axe/datafeeder"
	"github.com/rivo/tview"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
)

const (
	errorDelayTime = 1
)

type tableView struct {
	*tview.Table

	client       *kubernetes.Clientset
	app          *AppView
	data         []interface{}
	dataSource   datafeeder.DataSource
	lock         sync.Mutex
	sync         <-chan struct{}
	actions      []action.Action
	resourceKind ResourceKind
}

type EventHandler func(h status.GenericDrawer) func(event *tcell.EventKey) *tcell.EventKey

func NewTableView(app *AppView, kind string, h EventHandler) *tableView {
	v := ViewMap[kind]
	t := &tableView{
		Table: tview.NewTable(),
	}
	t.init(app, v.Kind, v.Feeder, v.Actions, h)
	if err := t.refresh(); err != nil {
		return t.UpdateStatus(err.Error(), true).(*tableView)
	}
	return t
}

func NewTableViewWithArgs(app *AppView, kind ResourceKind, feeder datafeeder.DataSource, actions []action.Action, h EventHandler) *tableView {
	t := &tableView{
		Table: tview.NewTable(),
	}
	t.init(app, kind, feeder, actions, h)
	if err := t.refresh(); err != nil {
		return t.UpdateStatus(err.Error(), true).(*tableView)
	}
	return t
}

func (t *tableView) init(app *AppView, resource ResourceKind, dataFeeder datafeeder.DataSource, actions []action.Action, h EventHandler) {
	{
		t.app = app
		t.resourceKind = resource
		t.dataSource = dataFeeder
		t.sync = app.syncs[resource.Kind]
		t.actions = actions
		t.client = app.clientset
	}
	{
		t.Table.SetBorder(true)
		t.Table.SetBackgroundColor(tcell.ColorBlack)
		t.Table.SetBorderAttributes(tcell.AttrBold)
		t.Table.SetSelectable(true, false)
		t.Table.SetTitle(t.resourceKind.Title)
	}
	if p, ok := t.app.pageRows[t.resourceKind.Kind]; ok {
		t.Table.Select(p.row, p.column)
	}

	actionMap := map[rune]action.Action{}
	for _, a := range t.actions {
		actionMap[a.Shortcut] = a
	}

	// todo: this needs to be changed to rowID to track selection if refresh happens
	t.Table.SetSelectionChangedFunc(func(row, column int) {
		t.app.pageRows[t.resourceKind.Kind] = position{
			row:    row,
			column: column,
		}
	})

	t.SetInputCapture(h(t))
}

func (t *tableView) run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-t.sync:
				if err := t.refresh(); err != nil {
					t.UpdateStatus(err.Error(), true)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (t *tableView) GetSelectionName() string {
	row, _ := t.Table.GetSelection()
	cell := t.Table.GetCell(row, 0)

	return strings.SplitN(cell.Text, " ", 2)[0]
}

func (t *tableView) SwitchToRootPage() {
	t.app.SwitchPage(t.app.currentPage, t.app.CurrentPage())
	t.app.Application.Draw()
}

func (t *tableView) refresh() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err := t.dataSource.Refresh(); err != nil {
		return err
	}

	t.drawTable()
	return nil
}

func (t *tableView) drawTable() {
	t.Clear()

	header := t.dataSource.Header()
	data := t.dataSource.Data()

	for col, name := range header {
		t.addHeaderCell(col, name)
	}

	for r, row := range data {
		if len(row) > 0 && row[0] == "" {
			continue
		}
		for col, value := range row {
			t.addBodyCell(r, col, value)
		}
	}
	t.GetApplication().Draw()
}

func (t *tableView) addHeaderCell(col int, name string) {
	c := tview.NewTableCell(fmt.Sprintf("[white]%s", name)).SetSelectable(false)
	{
		c.SetExpansion(1)
		c.SetTextColor(tcell.ColorAntiqueWhite)
		c.SetAttributes(tcell.AttrBold)
	}
	t.SetCell(0, col, c)
}

func (t *tableView) addBodyCell(row, col int, value string) {
	c := tview.NewTableCell(fmt.Sprintf("%s", value))
	{
		c.SetExpansion(1)
		c.SetTextColor(tcell.ColorAntiqueWhite)
	}
	t.SetCell(row+1, col, c)
}

func (t *tableView) InsertDialog(name string, page tview.Primitive, dialog tview.Primitive) {
	newpage := tview.NewPages()
	newpage.AddPage(name, t.app.CurrentPage(), true, true).
		AddPage("dialog", center(dialog, 40, 15), true, true)
	t.app.SwitchPage(t.app.currentPage, newpage)
	t.app.Application.SetFocus(dialog)
}

func (t *tableView) UpdateStatus(status string, isError bool) tview.Primitive {
	statusBar := tview.NewTextView()
	statusBar.SetBorder(true)
	statusBar.SetBorderAttributes(tcell.AttrBold)
	statusBar.SetBorderPadding(1, 1, 1, 1)
	if isError {
		statusBar.SetTitle("Error")
		statusBar.SetTitleColor(tcell.ColorRed)
		statusBar.SetTextColor(tcell.ColorRed)
		statusBar.SetBorderColor(tcell.ColorRed)
	} else {
		statusBar.SetTitle("Progress")
		statusBar.SetTitleColor(tcell.ColorYellow)
		statusBar.SetTextColor(tcell.ColorYellow)
		statusBar.SetBorderColor(tcell.ColorYellow)
	}
	statusBar.SetText(status)
	statusBar.SetTextAlign(tview.AlignCenter)
	newpage := tview.NewPages().
		AddPage("status", t.app.tableViews[t.app.currentPage], true, true).
		AddPage("dialog", center(statusBar, 100, 5), true, true)
	t.app.SwitchPage(t.app.currentPage, newpage)

	if isError {
		go func() {
			time.Sleep(time.Second * errorDelayTime)
			t.SwitchToRootPage()
		}()
	}
	return t
}

func (t *tableView) GetClientSet() *kubernetes.Clientset {
	return t.client
}

func (t *tableView) GetResourceKind() string {
	return t.resourceKind.Kind
}

func (t *tableView) GetCurrentPage() string {
	return t.app.currentPage
}

func (t *tableView) GetAction() []action.Action {
	return t.actions
}

func (t *tableView) GetApplication() *tview.Application {
	return t.app.Application
}

func (t *tableView) GetCurrentPrimitive() tview.Primitive {
	if t.app.drawQueue.Empty() {
		return t.app.tableViews[RootPage]
	}
	return t.app.drawQueue.Last()
}

func (t *tableView) SwitchPage(page string, draw tview.Primitive) {
	t.app.SwitchPage(page, draw)
}

func (t *tableView) GetTable() *tview.Table {
	return t.Table
}
