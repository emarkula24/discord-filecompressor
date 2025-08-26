package model

import "ffmpeg/wrapper/metadata/pkg/model"

type ConvertedVideo struct {
	Link        string         `json:"link"`
	OldMetadata model.Metadata `json:"metadata"`
}
