package playground

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

func WriteTextFile() error {
	targetDir := os.Getenv("FILE_OUTPUT")
	if targetDir == "" {
		return fmt.Errorf("FILE_OUTPUT environment variable is not set")
	}

	baseFilename := "example.txt"
	content := fmt.Sprintf("dummy data (%s)", time.Now().Format(time.RFC3339))

	filename := baseFilename
	for i := 1; ; i++ {
		filePath := filepath.Join(targetDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		filename = fmt.Sprintf("example_%d.txt", i)
	}

	if err := os.WriteFile(filepath.Join(targetDir, filename), []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	fmt.Printf("Creating file: %s\n", filename)

	return nil
}

func WaitForInterruptSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	log.Println("waiting for interrupt signal")
	<-sigChan
	log.Println("interrupt signal received")
}
