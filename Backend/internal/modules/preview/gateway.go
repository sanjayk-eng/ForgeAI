package preview

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

type Gateway struct {
	service *Service
}

func NewGateway(service *Service) *Gateway {
	return &Gateway{service: service}
}

func (gateway *Gateway) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := gateway.service.projectFromHost(request.Host)
	if !ok {
		http.NotFound(writer, request)
		return
	}
	containerID, containerPort, err := gateway.service.ResolveTarget(request.Context(), projectID)
	if err != nil {
		http.Error(writer, "Project preview is unavailable", http.StatusServiceUnavailable)
		return
	}
	targetURL := &url.URL{Scheme: "http", Host: net.JoinHostPort("preview.internal", strconv.Itoa(containerPort))}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	transport := gateway.service.previewTransport(containerID, containerPort)
	defer transport.CloseIdleConnections()
	proxy.Transport = transport
	proxy.FlushInterval = -1
	proxy.ErrorHandler = func(writer http.ResponseWriter, _ *http.Request, _ error) {
		http.Error(writer, "Project application is not responding", http.StatusBadGateway)
	}
	proxy.ServeHTTP(writer, request)
}
