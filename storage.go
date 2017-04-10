package main

import (
	"encoding/base64"
	"errors"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/russross/blackfriday"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
)

var StorageFileNotExistError = errors.New("File does not exist")

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
	case ".md":
		return "markdown"
	case ".markdown":
		return "markdown"
	default:
		return "text"
	}
}

func (f *HakoFile) String() string {
	return string(f.Contents)
}

func (f *HakoFile) Markdown() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon(f.Contents)))
}

func storageGet(f *HakoFile) error {
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(f.Owner))
	filePath := filepath.Join(userPrefix, filepath.Clean(f.Path))
	objHandle := bucket.Object(filePath)
	rc, err := objHandle.NewReader(ctx)
	if err == storage.ErrObjectNotExist {
		return StorageFileNotExistError
	} else if err != nil {
		return err
	}

	// Fetch file contents
	defer rc.Close()
	contents, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}

	objAttrs, err := objHandle.Attrs(ctx)
	if err != nil {
		return err
	}

	f.Path = strings.Join(strings.Split(objAttrs.Name, "/")[1:], "/")
	f.Size = objAttrs.Size
	f.Created = objAttrs.Created
	f.Updated = objAttrs.Updated
	f.Contents = contents
	return nil
}

func storageList(folder *HakoFile) ([]*HakoFile, error) {
	if !folder.IsFolder() {
		return nil, errors.New("Can only list files of folders")
	}
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(folder.Owner))
	folderFullPath := filepath.Join(userPrefix, filepath.Clean(folder.Path)) + "/"
	it := bucket.Objects(ctx, &storage.Query{Prefix: folderFullPath, Delimiter: "/"})

	files := []*HakoFile{}
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		// Handle synthetic folders
		if objAttrs.Prefix != "" {
			continue
			/*
				files = append(files, &HakoFile{
					Owner: folder.Owner,
					Path:  strings.Join(strings.Split(objAttrs.Prefix, "/")[1:], "/"),
				})
			*/
		}
		// Handle normal files
		files = append(files, &HakoFile{
			Owner:   folder.Owner,
			Path:    strings.Join(strings.Split(objAttrs.Name, "/")[1:], "/"),
			Size:    objAttrs.Size,
			Created: objAttrs.Created,
			Updated: objAttrs.Updated,
		})
	}
	return files, nil
}

func storagePut(f *HakoFile) error {
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(f.Owner))
	filePath := filepath.Join(userPrefix, filepath.Clean(f.Path))
	objHandle := bucket.Object(filePath)
	wc := objHandle.NewWriter(ctx)

	if _, err := wc.Write(f.Contents); err != nil {
		return err
	}
	return wc.Close()
}

func storageDel(f *HakoFile) error {
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(f.Owner))
	filePath := filepath.Join(userPrefix, filepath.Clean(f.Path))
	objHandle := bucket.Object(filePath)
	return objHandle.Delete(ctx)
}
