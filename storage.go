package main

import (
	"encoding/base64"
	"errors"
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
	if f.Path == "." {
		return "Home"
	}
	if f.Ext() == ".folder" {
		return filepath.Base(f.Path[:len(f.Path)-len(".folder")]) + "/"
	}
	return filepath.Base(f.Path)
}

func (f *HakoFile) ParentFolder() string {
	if filepath.Dir(f.Path) == "." {
		return ""
	}
	return filepath.Dir(f.Path) + ".folder"
}

func (f *HakoFile) Folder() string {
	if f.Ext() == ".folder" {
		return f.Path
	}
	return filepath.Dir(f.Path) + ".folder"
}

func (f *HakoFile) FolderName() string {
	if f.Ext() == ".folder" {
		return filepath.Base(f.Path[:len(f.Path)-len(".folder")])
	}
	folderName := filepath.Base(filepath.Dir(f.Path))
	if folderName == "." {
		return "Home"
	}
	return folderName
}

func (f *HakoFile) Ext() string {
	return strings.ToLower(filepath.Ext(f.Path))
}

func (f *HakoFile) Type() string {
	if f.Path == "." {
		return "folder"
	}
	switch f.Ext() {
	case ".folder":
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

func (f *HakoFile) Markdown() string {
	return string(blackfriday.MarkdownCommon(f.Contents))
}

func storageGet(f *HakoFile) error {
	if f.Type() == "folder" || f.Path == "." {
		return nil
	}

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

func storageList(email, folderPath string) ([]*HakoFile, error) {
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(email))
	if !strings.HasSuffix(folderPath, ".folder") {
		return nil, errors.New("Can't list files for folder not ending in '.folder'")
	}
	folderPath = folderPath[:len(folderPath)-len(".folder")]
	folderFullPath := filepath.Join(userPrefix, filepath.Clean(folderPath)) + "/"
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
			files = append(files, &HakoFile{
				Owner: email,
				Path:  strings.Join(strings.Split(objAttrs.Prefix, "/")[1:], "/") + ".folder",
			})
			continue
		}
		// Handle normal files
		files = append(files, &HakoFile{
			Owner:   email,
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
