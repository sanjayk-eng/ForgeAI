package preview

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var projectLabelPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

func (service *Service) previewURL(projectID string) (string, error) {
	label, err := service.projectLabel(projectID)
	if err != nil {
		return "", err
	}
	preview := *service.origin
	preview.Host = label + "." + service.origin.Host
	preview.Path = "/"
	preview.RawQuery = ""
	return preview.String(), nil
}

func (service *Service) projectLabel(projectID string) (string, error) {
	projectID = strings.ToLower(strings.TrimSpace(projectID))
	if len(projectID) == 0 || !projectLabelPattern.MatchString(projectID) {
		return "", fmt.Errorf("invalid project ID")
	}
	label := projectID + "--" + service.signature(projectID)
	if len(label) > 63 {
		return "", fmt.Errorf("project ID is too long for a preview hostname")
	}
	return label, nil
}

func (service *Service) projectFromHost(requestHost string) (string, bool) {
	hostname, requestPort, ok := splitRequestHost(requestHost)
	if !ok || !samePort(service.origin, requestPort) {
		return "", false
	}
	hostname = strings.ToLower(hostname)
	suffix := "." + strings.ToLower(service.origin.Hostname())
	if !strings.HasSuffix(hostname, suffix) {
		return "", false
	}
	label := strings.TrimSuffix(hostname, suffix)
	if strings.Contains(label, ".") {
		return "", false
	}
	separator := strings.LastIndex(label, "--")
	if separator <= 0 {
		return "", false
	}
	projectID := label[:separator]
	provided, err := hex.DecodeString(label[separator+2:])
	if err != nil || !projectLabelPattern.MatchString(projectID) || !hmac.Equal(provided, service.signatureBytes(projectID)) {
		return "", false
	}
	return projectID, true
}

func (service *Service) signature(projectID string) string {
	return hex.EncodeToString(service.signatureBytes(projectID))
}

func (service *Service) signatureBytes(projectID string) []byte {
	mac := hmac.New(sha256.New, service.key)
	_, _ = mac.Write([]byte(projectID))
	return mac.Sum(nil)[:12]
}

func splitRequestHost(requestHost string) (hostname, port string, ok bool) {
	if strings.Count(requestHost, ":") == 1 {
		hostname, port, err := net.SplitHostPort(requestHost)
		return hostname, port, err == nil
	}
	return requestHost, "", requestHost != ""
}

func samePort(origin *url.URL, requestPort string) bool {
	originPort := origin.Port()
	if originPort == "" {
		if origin.Scheme == "https" {
			originPort = "443"
		} else {
			originPort = "80"
		}
	}
	if requestPort == "" {
		requestPort = originPort
	}
	return requestPort == originPort
}
