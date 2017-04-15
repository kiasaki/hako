package main

import (
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/russross/blackfriday"
)

type HakoFile struct {
	Owner    string
	Path     string
	Size     int64
	Created  time.Time
	Updated  time.Time
	Contents []byte
}

func (f *HakoFile) Name() string {
	if f.Path == "." || f.Path == "" {
		return "Home"
	}
	if f.IsFolder() {
		return filepath.Base(f.Path) + "/"
	}
	return filepath.Base(f.Path)
}

func (f *HakoFile) ParentPath() string {
	return filepath.Dir(f.Path)
}

func (f *HakoFile) Ext() string {
	return strings.ToLower(filepath.Ext(f.Path))
}

func (f *HakoFile) IsFolder() bool {
	return f.Type() == "folder"
}

func (f *HakoFile) Type() string {
	switch f.Ext() {
	case "":
		return "folder"
	case ".":
		return "folder"
	case ".png":
		return "image"
	case ".jpg":
		return "image"
	case ".jpeg":
		return "image"
	case ".svg":
		return "image"
	case ".gif":
		return "image"
	case ".pdf":
		return "binary"
	case ".zip":
		return "binary"
	case ".tar":
		return "binary"
	case ".gz":
		return "binary"
	case ".dmg":
		return "binary"
	case ".iso":
		return "binary"
	case ".mp3":
		return "audio"
	case ".ogg":
		return "audio"
	case ".flac":
		return "audio"
	case ".wav":
		return "audio"
	case ".aiff":
		return "audio"
	case ".webm":
		return "audio"
	case ".vault":
		return "vault"
	case ".md":
		return "markdown"
	case ".markdown":
		return "markdown"
	default:
		return "text"
	}
}

func (f *HakoFile) ViewNeedsContents() bool {
	typ := f.Type()
	return typ == "vault" || typ == "markdown" || typ == "text"
}

func (f *HakoFile) String() string {
	return string(f.Contents)
}

func (f *HakoFile) Markdown() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon(f.Contents)))
}

func (f *HakoFile) UpdatedString() string {
	loc, _ := time.LoadLocation("EST")
	return f.Updated.In(loc).Format("2006-01-02 15:04")
}

func (f *HakoFile) SizeString() string {
	return bytesString(uint64(f.Size))
}

func (f *HakoFile) SignedURL() string {
	url, err := fileSignedURL(f)
	if err != nil {
		panic(err)
	}
	return url
}
