package main

import (
	"errors"
	"testing"
)

func Test_getFilename(t *testing.T) {
	tests := []struct {
		name    string
		link    string
		want    string
		wantErr error
	}{
		{
			name:    "escaped link",
			link:    "https://raw.githubusercontent.com/rdmyldz/i2t/master/tesseract/testdata/bar%C4%B1%C5%9F.png",
			want:    "barış.png",
			wantErr: nil,
		},
		{
			name:    "unescaped link",
			link:    "https://raw.githubusercontent.com/rdmyldz/i2t/master/tesseract/testdata/a.png",
			want:    "a.png",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFilename(tt.link)
			if tt.wantErr == nil && err != nil {
				t.Errorf("expected nil error, got: %v", err)
			}
			if got != tt.want {
				t.Errorf("getFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDirName(t *testing.T) {

	tests := []struct {
		name string
		link string
		want string
	}{
		{
			name: "valid link",
			link: "https://api.github.com/repos/rdmyldz/i2t/contents/tesseract/testdata",
			want: "testdata",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDirName(tt.link); got != tt.want {
				t.Errorf("getDirName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getApiLink(t *testing.T) {
	tests := []struct {
		name     string
		htmlLink string
		want     string
		wantErr  error
	}{
		{
			name:     "valid html link",
			htmlLink: "https://github.com/tesseract-ocr/tesseract/tree/4.0/m4",
			want:     "https://api.github.com/repos/tesseract-ocr/tesseract/contents/m4?ref=4.0",
			wantErr:  nil,
		},
		{
			name:     "invalid html link",
			htmlLink: "https://api.github.com/repos/tesseract-ocr/tesseract/contents/m4?ref=4.0",
			want:     "",
			wantErr:  ErrInvalidURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getApiLink(tt.htmlLink)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("expected nil error, got: %v", err)
			}

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Errorf("getApiLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
