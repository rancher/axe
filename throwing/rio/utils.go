package rio

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rancher/axe/throwing"
	"github.com/rancher/axe/throwing/datafeeder"
	"github.com/rancher/axe/throwing/types"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func actionView(t *throwing.TableView) {
	list := newSelectionList()
	list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
		switch s {
		case "inspect":
			formatList := newSelectionList()
			formatList.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				inspect(s, t)
			})
			formatList.AddItem("yaml", "yaml format", 'y', nil)
			formatList.AddItem("json", "json format", 'j', nil)
			t.InsertDialog("inspect-format", t.GetCurrentPrimitive(), formatList)
		case "log":
			list := newContainerSelectionList(t)
			list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				logs(s, t)
			})
			t.InsertDialog("logs-containers", t.GetCurrentPrimitive(), list)
		case "exec":
			list := newContainerSelectionList(t)
			list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				execute(s, t)
			})
			t.InsertDialog("exec-containers", t.GetCurrentPrimitive(), list)
		case "pods":
			viewPods(t)
		case "delete":
			rm(t)
		}
	})

	for _, a := range t.GetAction() {
		list.AddItem(a.Name, a.Description, a.Shortcut, nil)
	}
	t.InsertDialog("option", t.GetCurrentPrimitive(), list)
}

func newSelectionList() *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(tcell.ColorGray)
	list.SetMainTextColor(tcell.ColorBlack)
	list.SetSecondaryTextColor(tcell.ColorPurple)
	list.SetShortcutColor(defaultBackGroundColor)
	return list
}

func newContainerSelectionList(t *throwing.TableView) *tview.List {
	list := newSelectionList()
	containers, err := listRioContainer(t.GetSelectionName(), t.GetClientSet())
	if err != nil {
		t.UpdateStatus(err.Error(), true)
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
func inspect(format string, t *throwing.TableView) {
	name := t.GetSelectionName()
	outBuffer := &strings.Builder{}
	errBuffer := &strings.Builder{}
	args := []string{"inspect", "--format", format, "-t", t.GetResourceKind(), name}
	cmd := exec.Command("rio", args...)
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer
	if err := cmd.Run(); err != nil {
		t.UpdateStatus(errBuffer.String(), true)
		return
	}

	inspectBox := tview.NewTextView()
	inspectBox.SetDynamicColors(true).SetBackgroundColor(defaultBackGroundColor)
	inspectBox.SetText(outBuffer.String())

	newpage := tview.NewPages().AddPage("inspect", inspectBox, true, true)
	t.SwitchPage(t.GetCurrentPage(), newpage)
}

func logs(container string, t *throwing.TableView) {
	name := t.GetSelectionName()
	args := []string{"logs", "-f"}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, name)
	cmd := exec.Command("rio", args...)

	logbox := tview.NewTextView()
	{
		logbox.SetTitle(fmt.Sprintf("logs - (%s)", name))
		logbox.SetBorder(true)
		logbox.SetTitleColor(tcell.ColorPurple)
		logbox.SetDynamicColors(true)
		logbox.SetBackgroundColor(tcell.ColorBlack)
		logbox.SetChangedFunc(func() {
			logbox.ScrollToEnd()
			t.GetApplication().Draw()
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
	t.SwitchPage(t.GetCurrentPage(), newpage)
}

func execute(container string, t *throwing.TableView) {
	name := t.GetSelectionName()
	shellArgs := []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c /bin/bash /dev/null || exec /bin/bash) || exec /bin/sh"}
	args := []string{"exec", "-it"}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, name)
	cmd := exec.Command("rio", append(args, shellArgs...)...)
	errorBuffer := &strings.Builder{}
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, errorBuffer, os.Stdin

	t.GetApplication().Suspend(func() {
		clearScreen()
		if err := cmd.Run(); err != nil {
			t.UpdateStatus(errorBuffer.String(), true)
		}
		return
	})
}

func rm(t *throwing.TableView) {
	name := t.GetSelectionName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to delete %s %s?", t.GetResourceKind(), name)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				args := []string{"rm", "-t", t.GetResourceKind(), name}
				cmd := exec.Command("rio", args...)
				if err := cmd.Run(); err != nil {
					t.UpdateStatus(err.Error(), true)
					return
				}
				t.SwitchToRootPage()
			} else if buttonLabel == "Cancel" {
				t.SwitchToRootPage()
			}
		})
	t.InsertDialog("delete", t.GetCurrentPrimitive(), modal)
}

func viewPods(t *throwing.TableView) {
	name := t.GetSelectionName()
	rkind := types.ResourceKind{
		Kind:  podKind,
		Title: "Pods",
	}

	var podEventHandler = func(t *throwing.TableView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				actionView(t)
			case tcell.KeyRune:
				switch event.Rune() {
				case 'i':
					inspect("yaml", t)
				case 'l':
					logs("", t)
				case 'x':
					execute("", t)
				case 'r':
					t.Refresh()
				case '/':
					t.ShowSearch()
				case 'm':
					t.ShowMenu()
				default:
					t.Navigate(event.Rune())
				}
			}
			return event
		}
	}

	feeder := datafeeder.NewDataFeeder(PodRefresher(name))
	table := t.NewNestTableView(rkind, feeder, DefaultAction, nil, podEventHandler)
	t.SwitchPage(t.GetCurrentPage(), table)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
