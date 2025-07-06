package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/alaingilbert/ogame/pkg/gameforge"
)

// TBotSolver service
func TbotSolver(apiKey string) gameforge.CaptchaSolver {
	return func(ctx context.Context, question, icons []byte) (int64, error) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("question", "question.png")
		if err != nil {
			return -1, err
		}
		if _, err = io.Copy(part, bytes.NewReader(question)); err != nil {
			return -1, err
		}
		part1, err := writer.CreateFormFile("icons", "icons.png")
		if err != nil {
			return -1, err
		}
		if _, err := io.Copy(part1, bytes.NewReader(icons)); err != nil {
			return -1, err
		}
		if err := writer.Close(); err != nil {
			return -1, err
		}

		req, err := http.NewRequest(http.MethodPost, "https://solver.obot.de/api/v1/captcha/solve", body)
		if err != nil {
			return -1, err
		}
		req.Header.Add("Content-Type", writer.FormDataContentType())
		req.Header.Set("OBOT_API_KEY", apiKey)
		req = req.WithContext(ctx)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return -1, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			by, err := io.ReadAll(resp.Body)
			if err != nil {
				return -1, err
			}
			return -1, errors.New("failed to auto solve captcha: " + string(by))
		}
		by, err := io.ReadAll(resp.Body)
		if err != nil {
			return -1, err
		}
		var answerJson struct {
			Answer int64 `json:"answer"`
		}
		if err := json.Unmarshal(by, &answerJson); err != nil {
			return -1, err
		}
		return answerJson.Answer, nil
	}
}
