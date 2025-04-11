package types

import "time"

type Document struct {
	Name     string    `json:"name"`
	SizeMB   float64   `json:"size_mb"`
	Modified time.Time `json:"modified"`
}

type PDFDocument struct {
	Document
}

type DOCXDocument struct {
	Document
	Pages int `json:"pages"`
}
