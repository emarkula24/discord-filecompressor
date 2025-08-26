package http

import (
	"context"
	"encoding/json"
	"ffmpeg/wrapper/conversion/pkg/model"
	"ffmpeg/wrapper/video/internal/gateway"
	"fmt"
	"net/http"
)

// Gateway defines an HTTP gateway for a rating service.

type Gateway struct {
	addr string
}

// New creates a new HTTP gateway for a rating service.

func New(addr string) *Gateway {

	return &Gateway{addr}

}

func (g *Gateway) GetCompressedVideo(ctx context.Context, duration model.Duration, videoLink model.VideoLink) (string, error) {

	req, err := http.NewRequest(http.MethodGet, g.addr+"/compress", nil)
	if err != nil {
		return "", err
	}

	req = req.WithContext(ctx)

	values := req.URL.Query()
	values.Add("duration", fmt.Sprintf("%v", duration))
	values.Add("videolink", string(videoLink))
	req.URL.RawQuery = values.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("non-2xx response: %v", resp)
	}

	var v string
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}
	return v, nil

}
