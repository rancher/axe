// +build rio

package rio

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

/*
	Refresher refreshes the data by invoking the defined functions. Right now refreshers are invoked by shell output,
	but it can customized by override buffer writer.
 */
var (
	RouteRefresher = func(b *bytes.Buffer) error {
		cmd := exec.Command("rio", "route")
		errBuffer := &strings.Builder{}
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	ServiceRefresher = func(b *bytes.Buffer) error {
		cmd := exec.Command("rio", "ps")
		errBuffer := &strings.Builder{}
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}
)


