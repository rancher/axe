package rio

import (
	"fmt"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	b := &strings.Builder{}
	w := anotherWriter{
		Writer: b,
	}
	var data = `NAME                                IMAGE                                                                                                                                     CREATED        STATE     SCALE     ENDPOINT   EXTERNAL   DETAIL
build-controller/build-controller   gcr.io/knative-releases/github.com/knative/build/cmd/controller@sha256:6c9133810e75c057e6084f5dc65b6c55cb98e42692f45241f8d0023050f27ba9   18 hours ago   active    1                               
cert-manager/cert-manager           daishan1992/cert-manager:latest                                                                                                           18 hours ago   active    1                               
istio/istio-citadel                 istio/citadel:1.0.6                                                                                                                       18 hours ago   active    1                               
istio/istio-gateway                 istio/proxyv2:1.0.6                                                                                                                       18 hours ago   active    1                               
istio/istio-pilot                   istio/pilot:1.0.6                                                                                                                         18 hours ago   active    1 `

	_, _ = w.Write([]byte(data))
	fmt.Println(b.String())
}
