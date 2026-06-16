package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Result struct {
	Filename  string    `json:"filename"`
	FilePath  string    `json:"file_path"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

func Run() (*Result, error) {
	savePath := "./backups"
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("backup_%s_%s.sql", os.Getenv("DB_NAME"), timestamp)
	filePath := filepath.Join(savePath, filename)

	cmd := exec.Command(
		"mysqldump",
		"-u", os.Getenv("DB_USER"),
		fmt.Sprintf("-p%s", os.Getenv("DB_PASSWORD")),
		"-h", os.Getenv("DB_HOST"),
		"-P", os.Getenv("DB_PORT"),
		"--single-transaction",
		"--routines",
		"--triggers",
		os.Getenv("DB_NAME"),
	)

	outFile, err := os.Create(filePath)
	// creates a file at the path specified by filePath
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile

	if err := cmd.Run(); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("mysqldump failed: %w", err)
	}

	info, _ := os.Stat(filePath)
	return &Result{
		Filename:  filename,
		FilePath:  filePath,
		SizeBytes: info.Size(),
		CreatedAt: time.Now(),
	}, nil
}

func DeleteOldBackups(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	entries, err := os.ReadDir("./backups")
	if err != nil {
		return err
	}
	for _, e := range entries {
		info, _ := e.Info()
		if !e.IsDir() && info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join("./backups", e.Name()))
		}
	}
	return nil
}
