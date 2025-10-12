package playground

import (
	"GoChatServer/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func WriteTextFile() error {
	var targetDir string
	if utils.IsWindows() {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting user home directory: %v", err)
		}
		targetDir = filepath.Join(userHome, "Desktop")
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting working directory: %v", err)
		}
		targetDir = cwd
	}

	baseFilename := "example.txt"
	content := fmt.Sprintf("dummy data (%s)", time.Now().Format(time.RFC3339))

	filename := baseFilename
	for i := 1; ; i++ {
		filePath := filepath.Join(targetDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		filename = fmt.Sprintf("example(%d).txt", i)
	}

	if err := os.WriteFile(filepath.Join(targetDir, filename), []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}
