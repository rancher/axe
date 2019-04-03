// +build rio

package rio

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
	"github.com/gdamore/tcell"
	"github.com/rancher/axe/axe/status"
	"github.com/rancher/norman/pkg/kv"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)


