package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func ResourceIdentity(projectID string) (containerName, volumeName string, err error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return "", "", ErrInvalidSandbox
	}
	hash := sha256.Sum256([]byte(projectID))
	suffix := hex.EncodeToString(hash[:])[:16]
	return fmt.Sprintf("forgeai-sandbox-%s", suffix), fmt.Sprintf("forgeai-workspace-%s", suffix), nil
}
