package main

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
)

var StorageFileNotExistError = errors.New("File does not exist")

type HakoFile struct {
	Owner    string
	Name     string
	Size     int64
	Created  time.Time
	Updated  time.Time
	Contents []byte
}

func (f *HakoFile) BaseName() string {
	return filepath.Base(f.Name)
}

func (f *HakoFile) Folder() string {
	if f.Ext() == ".folder" {
		return f.Name[:len(f.Name)-len(".folder")]
	}
	return filepath.Dir(f.Name)
}

func (f *HakoFile) Ext() string {
	return strings.ToLower(filepath.Ext(f.Name))
}

func (f *HakoFile) Type() string {
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
	default:
		return "text"
	}
}

func storageGet(f *HakoFile) error {
	if f.Type() == "folder" {
		return nil
	}

	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(f.Owner))
	folderPath := filepath.Join(userPrefix, filepath.Clean(f.Name))
	objHandle := bucket.Object(filepath.Join(userPrefix, folderPath))
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

	f.Name = objAttrs.Name
	f.Size = objAttrs.Size
	f.Created = objAttrs.Created
	f.Updated = objAttrs.Updated
	f.Contents = contents
	return nil
}

func storageList(email, path string) ([]*HakoFile, error) {
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(email))
	folderPath := filepath.Join(userPrefix, filepath.Clean(path))
	it := bucket.Objects(ctx, &storage.Query{Prefix: folderPath, Delimiter: "/"})

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
				Name:  objAttrs.Prefix + ".folder",
			})
			continue
		}
		// Handle normal files
		files = append(files, &HakoFile{
			Owner:   email,
			Name:    objAttrs.Name,
			Size:    objAttrs.Size,
			Created: objAttrs.Created,
			Updated: objAttrs.Updated,
		})
	}
	return files, nil
}
