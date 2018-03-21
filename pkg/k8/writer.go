package k8

import "k8s.io/client-go/kubernetes"

type K8Request struct {
	Type RequestType
	// Encapsulates a K8 resource type
	Object interface{}
}

type Writer struct {
	Events chan *K8Request
	Client kubernetes.Interface
}

// Can probably genericize writer to make this work
func (w *Writer) NewWriter(clientset kubernetes.Interface) *Writer {
	return &Writer{
		Events: make(chan *EndpointsRequest),
		Client: clientset,
	}
}
