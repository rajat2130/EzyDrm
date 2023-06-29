package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	ezyDRMAPIURL     = "https://api.ezydrm.com"
	ezyDRMUsername   = "your_username"
	ezyDRMPassword   = "your_password"
	ezyDRMContentKey = "your_content_key"
)

func main() {
	filePath := "path/to/your/file.mp4"
	outputPath := "path/to/your/encrypted/file.enc"

	// Authenticate with ezyDRM
	token, err := authenticate()
	if err != nil {
		log.Fatal("Failed to authenticate with ezyDRM:", err)
	}

	// Generate a license key
	licenseKey, err := generateLicenseKey(token)
	if err != nil {
		log.Fatal("Failed to generate license key:", err)
	}

	// Encrypt the file
	err = encryptFile(filePath, outputPath, licenseKey)
	if err != nil {
		log.Fatal("Failed to encrypt file:", err)
	}

	fmt.Println("File encrypted successfully!")
}

func authenticate() (string, error) {
	authURL := fmt.Sprintf("%s/auth/token", ezyDRMAPIURL)

	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: ezyDRMUsername,
		Password: ezyDRMPassword,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(authURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.Token, nil
}

func generateLicenseKey(token string) (string, error) {
	keyURL := fmt.Sprintf("%s/license/generate", ezyDRMAPIURL)

	req, err := http.NewRequest("POST", keyURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		LicenseKey string `json:"license_key"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.LicenseKey, nil
}

func encryptFile(filePath, outputPath, licenseKey string) error {
	encryptURL := fmt.Sprintf("%s/encrypt", ezyDRMAPIURL)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	// Add the license key
	err = writer.WriteField("license_key", licenseKey)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", encryptURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
