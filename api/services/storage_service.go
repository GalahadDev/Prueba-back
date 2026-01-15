package services

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"bitacora-medica-backend/api/config"

	"github.com/google/uuid"
)

type StorageService struct {
	Config *config.Config
}

func NewStorageService(cfg *config.Config) *StorageService {
	return &StorageService{Config: cfg}
}

// UploadConsentPDF sube el archivo a Supabase y retorna la URL pública
func (s *StorageService) UploadConsentPDF(file *multipart.FileHeader) (string, error) {
	// 1. Validar extensión
	ext := filepath.Ext(file.Filename)
	if ext != ".pdf" {
		return "", fmt.Errorf("only PDF files are allowed")
	}

	// 2. Abrir el archivo
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Leer contenido
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// 3. Generar nombre único: {uuid}_{timestamp}.pdf
	fileName := fmt.Sprintf("%s_%d.pdf", uuid.New().String(), time.Now().Unix())

	// 4. Subir a Supabase Storage vía REST API
	// Endpoint: POST /storage/v1/object/{bucket}/{path}
	bucketName := "patients-consent"
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.Config.SupabaseURL, bucketName, fileName)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(fileBytes))
	if err != nil {
		return "", err
	}

	// Headers requeridos por Supabase
	req.Header.Set("Authorization", "Bearer "+s.Config.SupabaseKey)
	req.Header.Set("Content-Type", "application/pdf")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to request Supabase Storage", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Supabase Storage Error", "status", resp.StatusCode, "body", string(body))
		return "", fmt.Errorf("failed to upload file, status: %d", resp.StatusCode)
	}

	// 5. Construir URL pública (Asumiendo bucket público)
	// Formato: {supabaseUrl}/storage/v1/object/public/{bucket}/{fileName}
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.Config.SupabaseURL, bucketName, fileName)

	return publicURL, nil
}

// Agrega este método a la struct StorageService
func (s *StorageService) UploadImage(file *multipart.FileHeader) (string, error) {
	// 1. Validar extensión (Solo imágenes)
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("only JPG/PNG images are allowed")
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// 2. Generar nombre único
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	// 3. Subir a bucket "session-evidence"
	bucketName := "session-evidence"
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.Config.SupabaseURL, bucketName, fileName)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(fileBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+s.Config.SupabaseKey)
	// Detectar content-type correcto
	contentType := http.DetectContentType(fileBytes)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Supabase request failed", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	// 4. Retornar URL Pública
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.Config.SupabaseURL, bucketName, fileName)
	return publicURL, nil
}
