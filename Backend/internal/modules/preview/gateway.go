package preview

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Gateway struct {
	service   *Service
	transport *http.Transport
}

func NewGateway(service *Service) *Gateway {
	return &Gateway{
		service: service,
		transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}
}

func (gateway *Gateway) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	projectID, ok := gateway.service.projectFromHost(request.Host)
	if !ok {
		http.NotFound(writer, request)
		return
	}
	target, err := gateway.service.ResolveTarget(request.Context(), projectID)
	if err != nil {
		http.Error(writer, "Project preview is unavailable", http.StatusServiceUnavailable)
		return
	}
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = gateway.transport
	proxy.FlushInterval = -1
	proxy.ErrorHandler = func(writer http.ResponseWriter, _ *http.Request, _ error) {
		http.Error(writer, "Project application is not responding", http.StatusBadGateway)
	}
	proxy.ServeHTTP(writer, request)
}
