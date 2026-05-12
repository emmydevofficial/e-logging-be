package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func transcribeAudio(file multipart.File, filename string) (string, error) {
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return "", err
	}

	w.WriteField("model", "whisper-1")
	w.Close()

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	transcription, ok := result["text"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return transcription, nil
}

func paraphraseText(rawText string) (string, error) {
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	url := "https://api.openai.com/v1/chat/completions"

	prompt := "You are an expert Transmission Station Operator. Your task is to paraphrase the following dictated text into a professional, concise station log entry. Use standard transmission station terminology such as 'tripping report', 'ramping up generator', 'ramping down generator', etc. Fix any grammar errors and eliminate filler words. Do not add conversational pleasantries; return ONLY the logged event."

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "gpt-4o-mini", // or gpt-3.5-turbo
		"messages": []map[string]string{
			{"role": "system", "content": prompt},
			{"role": "user", "content": rawText},
		},
		"temperature": 0.3,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error during paraphrase: %s", resp.Status)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from OpenAI")
}