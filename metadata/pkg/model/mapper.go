package model

import (
	"ffmpeg/wrapper/src/gen"
)

func MetadataToProto(m *Metadata) *gen.Metadata {
	return &gen.Metadata{
		Filename:       m.Filename,
		NbStreams:      int32(m.NbStreams),
		NbPrograms:     int32(m.NbPrograms),
		FormatName:     m.FormatName,
		FormatLongName: m.FormatLongName,
		StartTime:      m.StartTime,
		Duration:       m.Duration,
		Size:           m.Size,
		BitRate:        m.BitRate,
		ProbeScore:     int32(m.ProbeScore),
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
