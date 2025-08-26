package http

import (
	"context"
	"encoding/json"
	"ffmpeg/wrapper/metadata/pkg/model"
	"ffmpeg/wrapper/video/internal/gateway"
	"fmt"
	"net/http"
)

type Gateway struct {
	addr string
}

func New(addr string) *Gateway {
	return &Gateway{addr}
}

func (g *Gateway) Get(ctx context.Context, path string) (*model.Metadata, error) {
	req, err := http.NewRequest(http.MethodGet, g.addr+"/metadata", nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("path", path)
	req.URL.RawQuery = values.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non-2xx response: %v", resp)
	}
	var v *model.Metadata
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v, nil

}
