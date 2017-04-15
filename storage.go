package main

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
)

var StorageFileNotExistError = errors.New("File does not exist")

func storageGet(f *HakoFile, fetchContents bool) error {
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(f.Owner))
	filePath := filepath.Join(userPrefix, filepath.Clean(f.Path))
	objHandle := bucket.Object(filePath)

	if fetchContents {
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
		f.Contents = contents
	}

	objAttrs, err := objHandle.Attrs(ctx)
	if err != nil {
		return err
	}

	f.Path = strings.Join(strings.Split(objAttrs.Name, "/")[1:], "/")
	f.Size = objAttrs.Size
	f.Created = objAttrs.Created
	f.Updated = objAttrs.Updated
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
