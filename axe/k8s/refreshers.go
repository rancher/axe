// +build k8s

package k8s

import (
	"bytes"
	"os"
	"sort"
	"strings"

	"github.com/rancher/norman/types/convert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Wrapper struct {
	Group, Version, Name string
}

func (w Wrapper) RefreshResource(b *bytes.Buffer) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	restClient := clientset.RESTClient()
	apiPrefix := "apis"
	if w.Group == "" {
		apiPrefix = "api"
	}
	if w.Version == "" {
		w.Version = "v1"
	}
	req := restClient.Get().Prefix(apiPrefix, w.Group, w.Version).Resource(w.Name).Param("includeObject", "Object")
	header := "application/json;as=Table;g=meta.k8s.io;v=v1beta1, application/json"
	req.SetHeader("Accept", header)
	table := &v1beta1.Table{}
	if err := req.Do().Into(table); err != nil {
		return err
	}
	// insert namespace
	table.ColumnDefinitions = append([]v1beta1.TableColumnDefinition{
		{
			Name: "NAMESPACE",
		},
	}, table.ColumnDefinitions...)

	for i, header := range table.ColumnDefinitions {
		b.Write([]byte(strings.ToUpper(header.Name)))
		if i == len(table.ColumnDefinitions)-1 {
			b.Write([]byte("\n"))
		} else {
			b.Write([]byte("\t"))
		}
	}

	for _, row := range table.Rows {
		converted, err := runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
		if err != nil {
			return err
		}
		row.Object.Object = converted
		namespace := ""
		object, ok := row.Object.Object.(metav1.Object)
		if ok {
			namespace = object.GetNamespace()
		}
		row.Cells = append([]interface{}{namespace}, row.Cells...)
		for i, column := range row.Cells {
			b.Write([]byte(convert.ToString(column)))
			if i == len(row.Cells)-1 {
				b.Write([]byte("\n"))
			} else {
				b.Write([]byte("\t"))
			}
		}
	}
	return nil
}

func RefreshResourceKind(b *bytes.Buffer) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return err
	}
	clientset := kubernetes.NewForConfigOrDie(restConfig)

	Header := []string{
		"NAME",
		"GROUPVERSION",
	}
	list, err := clientset.Discovery().ServerPreferredResources()
	if err != nil {
		return err
	}
	for i, header := range Header {
		b.Write([]byte(header))
		if i == len(Header)-1 {
			b.Write([]byte("\n"))
		} else {
			b.Write([]byte("\t"))
		}
	}
	var resources []struct {
		Name         string
		GroupVersion string
	}


	for _, l := range list {
		for _, r := range l.APIResources {
			resources = append(resources, struct {
				Name         string
				GroupVersion string
			}{Name: r.Name, GroupVersion: l.GroupVersion})
		}
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	for _, r := range resources {
		b.Write([]byte(r.Name))
		b.Write([]byte("\t"))
		b.Write([]byte(r.GroupVersion))
		b.Write([]byte("\n"))
	}
	return nil
}
