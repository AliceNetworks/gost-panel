package api

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

// AgentVersion represents the agent version info
type AgentVersion struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Checksum  string `json:"checksum,omitempty"`
}

// CurrentAgentVersion is the current agent version
// This should be updated when releasing new agent versions
var CurrentAgentVersion = "1.3.0"
var AgentBuildTime = "2026-02-06"

// agentGetVersion returns the current agent version
func (s *Server) agentGetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, AgentVersion{
		Version:   CurrentAgentVersion,
		BuildTime: AgentBuildTime,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	})
}

// agentCheckUpdate checks if an update is available
func (s *Server) agentCheckUpdate(c *gin.Context) {
	clientVersion := c.Query("version")
	clientOS := c.DefaultQuery("os", runtime.GOOS)
	clientArch := c.DefaultQuery("arch", runtime.GOARCH)

	// Check if update is available
	needsUpdate := compareVersions(clientVersion, CurrentAgentVersion) < 0

	response := gin.H{
		"current_version": clientVersion,
		"latest_version":  CurrentAgentVersion,
		"needs_update":    needsUpdate,
		"build_time":      AgentBuildTime,
	}

	if needsUpdate {
		// Check if binary exists for this OS/arch
		binaryPath := getAgentBinaryPath(clientOS, clientArch)
		if _, err := os.Stat(binaryPath); err == nil {
			checksum, _ := getFileChecksum(binaryPath)
			response["download_url"] = fmt.Sprintf("/agent/download/%s/%s", clientOS, clientArch)
			response["checksum"] = checksum
		}
	}

	c.JSON(http.StatusOK, response)
}

// agentDownload serves the agent binary for download
func (s *Server) agentDownload(c *gin.Context) {
	osName := c.Param("os")
	archName := c.Param("arch")

	// Validate OS and arch
	validOS := map[string]bool{"linux": true, "darwin": true, "windows": true}
	validArch := map[string]bool{"amd64": true, "arm64": true, "386": true, "arm": true}

	if !validOS[osName] || !validArch[archName] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid os or arch"})
		return
	}

	binaryPath := getAgentBinaryPath(osName, archName)

	// Check if file exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "agent binary not found",
			"message": fmt.Sprintf("Binary for %s/%s is not available. Please build it first.", osName, archName),
		})
		return
	}

	// Get file info
	fileInfo, err := os.Stat(binaryPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set headers
	fileName := "gost-agent"
	if osName == "windows" {
		fileName += ".exe"
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("X-Agent-Version", CurrentAgentVersion)

	// Add checksum header
	if checksum, err := getFileChecksum(binaryPath); err == nil {
		c.Header("X-Checksum-SHA256", checksum)
	}

	c.File(binaryPath)
}

// getAgentBinaryPath returns the path to the agent binary
func getAgentBinaryPath(osName, archName string) string {
	basePath := "/root/gost-panel/dist/agents"
	fileName := fmt.Sprintf("gost-agent-%s-%s", osName, archName)
	if osName == "windows" {
		fileName += ".exe"
	}
	return filepath.Join(basePath, fileName)
}

// getFileChecksum calculates SHA256 checksum of a file
func getFileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// compareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// clientHeartbeat handles client heartbeat requests
func (s *Server) clientHeartbeat(c *gin.Context) {
	token := c.Param("token")

	// Update client status
	err := s.svc.UpdateClientHeartbeat(token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
