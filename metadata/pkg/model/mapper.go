package model

import (
	"ffmpeg/wrapper/src/gen"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

func MetadataToProto(m *Metadata) *gen.Metadata {
	return &gen.Metadata{
		Filename:       m.Filename,
		NbStreams:      int64(m.NbStreams),
		NbPrograms:     int64(m.NbPrograms),
		FormatName:     m.FormatName,
		FormatLongName: m.FormatLongName,
		StartTime:      m.StartTime,
		Duration:       m.Duration,
		Size:           m.Size,
		BitRate:        m.BitRate,
		ProbeScore:     int64(m.ProbeScore),
		Tags:           TagsToProto(m.Tags),
	}
}

func TagsToProto(t Tags) *gen.Tags {
	return &gen.Tags{
		CompatibleBrands: t.CompatibleBrands,
		MajorBrand:       t.MajorBrand,
		MinorVersion:     t.MinorVersion,
	}
}

func MetadataFromProto(m *gen.Metadata) *Metadata {
	return &Metadata{
		Filename:       m.Filename,
		NbStreams:      int(m.NbStreams),
		NbPrograms:     int(m.NbPrograms),
		FormatName:     m.FormatName,
		FormatLongName: m.FormatLongName,
		StartTime:      m.StartTime,
		Duration:       m.Duration,
		Size:           m.Size,
		BitRate:        m.BitRate,
		ProbeScore:     int(m.ProbeScore),
		Tags:           TagsFromProto(m.Tags),
	}
}

func TagsFromProto(t *gen.Tags) Tags {
	return Tags{
		CompatibleBrands: t.CompatibleBrands,
		MajorBrand:       t.MajorBrand,
		MinorVersion:     t.MinorVersion,
	}
}

func PresignedToProto(req *v4.PresignedHTTPRequest) *gen.PresignedRequest {
	headers := make(map[string]string)
	for k, v := range req.SignedHeader {
		if len(v) > 0 {
			headers[k] = v[0] // use first value for simplicity
		}
	}

	return &gen.PresignedRequest{
		Method:  req.Method,
		Url:     req.URL,
		Headers: headers,
	}
}
