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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

const (
	maxUploadSize = 20 << 20
)

var (
	tempDir   string
	mutex     sync.Mutex
	s3Client  *s3.S3
	s3Bucket  string
	awsRegion string
)

func init() {
	var err error
	tempDir, err = os.MkdirTemp("", "image-conversion")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}

	awsAccessKey := os.Getenv("AWS_ACCESS_KEY")
	awsSecretKey := os.Getenv("AWS_SECRET_KEY")
	s3Bucket = os.Getenv("AWS_S3_BUCKET")
	awsRegion = os.Getenv("AWS_REGION")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	s3Client = s3.New(sess)
}

func main() {
	http.HandleFunc("/convert", handleConvert)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server is running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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

	s3URL, err := uploadToS3(outputPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error uploading to S3: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(s3URL))
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

func uploadToS3(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	id := uuid.New().String()

	ext := filepath.Ext(filePath)

	fileName := id + ext

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3Bucket, awsRegion, fileName), nil
}
