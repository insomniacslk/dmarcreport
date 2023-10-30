package main

import (
	"bytes"
	"fmt"

	report "github.com/oliverpool/go-dmarc-report"
)

type Report struct {
	Name string
	Data []byte
	Type string
}

func parseReport(r *Report) (*report.Aggregate, error) {
	var (
		agg *report.Aggregate
		err error
	)
	f := bytes.NewReader(r.Data)
	switch r.Type {
	case "application/zip":
		agg, err = report.DecodeZip(f, int64(len(r.Data)))
	case "application/x-gzip-compressed":
		agg, err = report.DecodeGzip(f)
	case "text/plain":
		agg, err = report.Decode(f)
	default:
		return nil, fmt.Errorf("unknown type '%s'", r.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	return agg, nil
}
