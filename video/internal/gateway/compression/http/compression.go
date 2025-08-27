package http

import (
	"context"
	"encoding/json"
	"ffmpeg/wrapper/compression/pkg/model"
	"ffmpeg/wrapper/pkg/discovery"
	"ffmpeg/wrapper/video/internal/gateway"
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

// Gateway defines an HTTP gateway for a rating service.

type Gateway struct {
	registry discovery.Registry
}

// New creates a new HTTP gateway for a rating service.

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func (g *Gateway) GetCompressedVideo(ctx context.Context, duration model.Duration, videoLink model.VideoLink) (string, error) {
	addrs, err := g.registry.ServiceAddresses(ctx, "compression")
	if err != nil {
		return "", err
	}
	url := "http://" + addrs[rand.Intn(len(addrs))] + "/compress"
	log.Printf("%s", "Calling metadata service. Request: GET "+url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
