package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/pkg/browser"
)

var imgBBUploadUrl = "https://imgbb.com/json"
var googleSearchUrl = "https://lens.google.com/uploadbyurl"

type imgBBResponse struct {
	Image struct {
		URL       string `json:"url"`
		DeleteUrl string `json:"delete_url"`
	}
}

func UploadImage(img io.Reader) (string, error) {
	form, body, err := buildImgBBForm(img)
	if err != nil {
		return "", err
	}
	defer form.Close()

	resp, err := http.Post(imgBBUploadUrl, form.FormDataContentType(), body)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ImgBB returned %v status code", resp.StatusCode)
	}
	defer resp.Body.Close()
	var respData imgBBResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", err
	}
	return respData.Image.URL, nil
}

func GetSearchURL(imageURL string) string {
	timestamp := time.Now().UnixMilli()
	return fmt.Sprintf(
		"%v?hl=en&re=df&vpw=1920&vph=1080&st=%v&url=%v",
		googleSearchUrl,
		timestamp,
		imageURL,
	)
}

func OpenUrlInBrowser(url string) {
	browser.OpenURL(url)
}

func buildImgBBForm(img io.Reader) (*multipart.Writer, *bytes.Buffer, error) {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)
	fileWriter, err := form.CreateFormFile("source", "screenshot.png")
	if err != nil {
		return nil, nil, err
	}
	_, err = io.Copy(fileWriter, img)
	if err != nil {
		return nil, nil, err
	}

	if err := form.WriteField("upload-expiration", "PT5M"); err != nil {
		return nil, nil, err
	}

	if err := form.WriteField("type", "file"); err != nil {
		return nil, nil, err
	}

	if err := form.WriteField("action", "upload"); err != nil {
		return nil, nil, err
	}
	return form, body, nil
}
