package rio

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"os/exec"
	"strings"
)

/*
	Refresher refreshes the data by invoking the defined functions. Right now refreshers are invoked by shell output,
	but it can customized by override buffer writer.
 */
var (
	RouteRefresher = func(b *bytes.Buffer) error {
		w := &anotherWriter{
			Writer: b,
		}
		data := &bytes.Buffer{}
		cmd := exec.Command("rio", "route")
		errBuffer := &strings.Builder{}
		cmd.Stdout = w
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		_, err := w.Write(data.Bytes())
		return err
	}

	ServiceRefresher = func(b *bytes.Buffer) error {
		w := &anotherWriter{
			Writer: b,
		}
		data := &bytes.Buffer{}
		cmd := exec.Command("rio", "ps")
		errBuffer := &strings.Builder{}
		cmd.Stdout = data
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		_, err := w.Write(data.Bytes())
		return err
	}

	PodRefresher = func(name string) func(b *bytes.Buffer) error {
		return func(b *bytes.Buffer) error {
			w := &anotherWriter{
				Writer: b,
			}
			data := &bytes.Buffer{}
			cmd := exec.Command("rio", "ps", "-c", name)
			errBuffer := &strings.Builder{}
			cmd.Stdout = data
			cmd.Stderr = errBuffer
			if err := cmd.Run(); err != nil {
				return errors.New(errBuffer.String())
			}
			_, err := w.Write(data.Bytes())
			return err
		}
	}
)

// todo: name something good
type anotherWriter struct {
	io.Writer
}

func (a anotherWriter) Write(p []byte) (n int, err error) {
	headerScanned := false
	writeTab := false
	var index []int
	offset := 0
	index = append(index, offset)
	for i, b := range p {
		if !headerScanned {
			if b == '\n' {
				headerScanned = true
				_, _ = a.Writer.Write(p[offset:i+1])
				offset = i+1
			} else if !writeTab && b != ' ' && b != '\t' {
				continue
			} else if b == ' ' {
				writeTab = true
				continue
			} else if writeTab && b != ' ' {
				_, _ = a.Writer.Write(p[offset:i])
				_, _ = a.Writer.Write([]byte("\t"))
				index = append(index, i)
				writeTab = false
				offset = i
			}
		} else {
			if b == '\n' || i == len(p)-1 {
				start := offset
				for k, j := range index {
					if k == 0 {
						continue
					}
					if k == len(index)-1 {
						_, _ = a.Writer.Write(p[start:i+1])
						continue
					}
					if start+j-index[k-1] > i {
						break
					}
					_, _ = a.Writer.Write(p[start:start+j-index[k-1]])
					_, _ = a.Writer.Write([]byte("\t"))
					start = start+j-index[k-1]
				}
				_, _ = a.Writer.Write([]byte("\n"))
				offset = i+1
			}
		}
	}
	return len(p), nil
}




