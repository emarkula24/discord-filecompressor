package model

import v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"

type Metadata struct {
	Filename       string `json:"filename"`
	NbStreams      int    `json:"nb_streams"`
	NbPrograms     int    `json:"nb_programs"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	StartTime      string `json:"start_time"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
	ProbeScore     int    `json:"probe_score"`
	Tags           Tags   `json:"tags"`
}

type Tags struct {
	CompatibleBrands string `json:"compatible_brands"`
	MajorBrand       string `json:"major_brand"`
	MinorVersion     string `json:"minor_version"`
}

type UploadURL struct {
	JobID        int64                    // corresponds to JobId in proto
	PresignedURL *v4.PresignedHTTPRequest // corresponds to PresignedUrl in proto
	ObjectKey    string                   // corresponds to ObjectKey in proto
}

type CompressionEvent struct {
	JobID     int64    `json:"job_id"`
	ObjectKey string   `json:"object_key"`
	Metadata  Metadata `json:"metadata"`
}
