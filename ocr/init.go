package ocr

import (
	"github.com/otiai10/gosseract/v2"
)

var client *gosseract.Client

func init() {
	client = gosseract.NewClient()
}
func Text(image string) (string, error) {
	// TODO: 支持多个 client 工作，减少由于多个请求带来的阻塞
	err := client.SetImage(image)
	if err != nil {
		return "", err
	}
	return client.Text()
}
