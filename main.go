package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	maxUploadSize = 10 << 20
	port          = ":8080"
)

var (
	tempDir string
	mutex   sync.Mutex
)

func main() {
	var err error
	tempDir, err = os.MkdirTemp("", "image-conversion")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	http.HandleFunc("/convert", handleConvert)
	fmt.Printf("Server is running on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	outputFormat := r.URL.Query().Get("format")
	if outputFormat == "" {
		outputFormat = "avif"
	}

	inputPath, outputPath, err := saveAndConvert(file, header.Filename, outputFormat)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing image: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", outputFormat))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(outputPath)))
	http.ServeFile(w, r, outputPath)
}

func saveAndConvert(file io.Reader, filename, outputFormat string) (string, string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	inputPath := filepath.Join(tempDir, filename)
	outputPath := filepath.Join(tempDir, strings.TrimSuffix(filename, filepath.Ext(filename))+"."+outputFormat)

	outFile, err := os.Create(inputPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create input file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		return "", "", fmt.Errorf("failed to save input file: %w", err)
	}

	if err := convertImage(inputPath, outputPath, outputFormat); err != nil {
		return "", "", fmt.Errorf("failed to convert image: %w", err)
	}

	return inputPath, outputPath, nil
}

func convertImage(inputPath, outputPath, outputFormat string) error {
	quality := "75"
	args := []string{
		inputPath,
		"-quality", quality,
		"-strip",
		"-auto-orient",
	}

	switch outputFormat {
	case "avif":
		args = append(args, "-define", "heic:speed=8", "-define", "heic:preserve-orientation=true")
	case "webp":
		args = append(args, "-define", "webp:lossless=false", "-define", "webp:method=6")
	case "jpg", "jpeg":
	case "png":
		args = append(args, "-define", "png:compression-level=9", "-define", "png:compression-strategy=2")
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	args = append(args, outputPath)

	cmd := exec.Command("convert", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("conversion failed: %w, output: %s", err, string(output))
	}
	return nil
}
