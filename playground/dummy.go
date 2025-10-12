package playground

import (
	"GoChatServer/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func WriteTextFile() error {
	if !utils.IsWindows() {
		return fmt.Errorf("not supported on this platform")
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %v", err)
	}

	desktopPath := filepath.Join(userHome, "Desktop")
	baseFilename := "example.txt"
	content := fmt.Sprintf("dummy data (%s)", time.Now().Format(time.RFC3339))

	filename := baseFilename
	for i := 1; ; i++ {
		filePath := filepath.Join(desktopPath, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		filename = fmt.Sprintf("example(%d).txt", i)
	}

	err = os.WriteFile(filepath.Join(desktopPath, filename), []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}
