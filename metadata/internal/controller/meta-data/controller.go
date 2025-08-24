package metadata

import (
	"context"
	"time"

	"gopkg.in/vansante/go-ffprobe.v2"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}
func (c *Controller) Get(ctx context.Context, path string) (*ffprobe.ProbeData, error) {
	ctx, cancelFn := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFn()

	data, err := ffprobe.ProbeURL(ctx, path)
	if err != nil {
		return nil, err
	}
	return data, nil
}
