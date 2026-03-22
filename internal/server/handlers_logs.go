package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (s *Server) GetLogs(c echo.Context) error {
	logDir, err := os.UserConfigDir()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot find config dir"})
	}
	logPath := filepath.Join(logDir, "zoro", "zoro.log")

	data, err := os.ReadFile(logPath)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"logs": ""})
	}

	// Return last N lines
	lines := splitTail(string(data), parseIntOr(c.QueryParam("lines"), 100))
	return c.JSON(http.StatusOK, map[string]string{"logs": lines})
}

func splitTail(s string, n int) string {
	lines := make([]string, 0, n)
	start := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\n' {
			if i+1 < start {
				lines = append(lines, s[i+1:start])
			}
			start = i
			if len(lines) >= n {
				break
			}
		}
	}
	if start > 0 && len(lines) < n {
		lines = append(lines, s[:start])
	}
	// Reverse
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

func parseIntOr(s string, fallback int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return fallback
}
