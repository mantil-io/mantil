// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// terraform/modules/dynamodb.zip
// terraform/modules/funcs.zip
// terraform/templates/main.tf
// aws/project-policy.json
package assets

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}


type assetFile struct {
	*bytes.Reader
	name            string
	childInfos      []os.FileInfo
	childInfoOffset int
}

type assetOperator struct{}

// Open implement http.FileSystem interface
func (f *assetOperator) Open(name string) (http.File, error) {
	var err error
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	content, err := Asset(name)
	if err == nil {
		return &assetFile{name: name, Reader: bytes.NewReader(content)}, nil
	}
	children, err := AssetDir(name)
	if err == nil {
		childInfos := make([]os.FileInfo, 0, len(children))
		for _, child := range children {
			childPath := filepath.Join(name, child)
			info, errInfo := AssetInfo(filepath.Join(name, child))
			if errInfo == nil {
				childInfos = append(childInfos, info)
			} else {
				childInfos = append(childInfos, newDirFileInfo(childPath))
			}
		}
		return &assetFile{name: name, childInfos: childInfos}, nil
	} else {
		// If the error is not found, return an error that will
		// result in a 404 error. Otherwise the server returns
		// a 500 error for files not found.
		if strings.Contains(err.Error(), "not found") {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
}

// Close no need do anything
func (f *assetFile) Close() error {
	return nil
}

// Readdir read dir's children file info
func (f *assetFile) Readdir(count int) ([]os.FileInfo, error) {
	if len(f.childInfos) == 0 {
		return nil, os.ErrNotExist
	}
	if count <= 0 {
		return f.childInfos, nil
	}
	if f.childInfoOffset+count > len(f.childInfos) {
		count = len(f.childInfos) - f.childInfoOffset
	}
	offset := f.childInfoOffset
	f.childInfoOffset += count
	return f.childInfos[offset : offset+count], nil
}

// Stat read file info from asset item
func (f *assetFile) Stat() (os.FileInfo, error) {
	if len(f.childInfos) != 0 {
		return newDirFileInfo(f.name), nil
	}
	return AssetInfo(f.name)
}

// newDirFileInfo return default dir file info
func newDirFileInfo(name string) os.FileInfo {
	return &bindataFileInfo{
		name:    name,
		size:    0,
		mode:    os.FileMode(2147484068), // equal os.FileMode(0644)|os.ModeDir
		modTime: time.Time{}}
}

// AssetFile return a http.FileSystem instance that data backend by asset
func AssetFile() http.FileSystem {
	return &assetOperator{}
}

var _terraformModulesDynamodbZip = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x0a\xf0\x66\x66\x11\x61\x60\x60\xe0\x60\xb8\x9d\xf2\x35\xe8\x9b\x3e\xbb\xf8\x19\x06\x06\x86\x47\x8c\x0c\x0c\xec\x0c\x32\x0c\xb9\x89\x99\x79\x7a\x25\x69\xa1\x21\x9c\x0c\xcc\x76\x3f\xbf\x27\xd8\xff\xfc\x9e\x50\x5a\xc1\xcd\xc0\xc8\xf2\x95\x91\x81\x01\xac\xb1\x75\x82\x2f\x5f\x93\x81\x80\xdb\x77\xdb\xa3\x4b\xd8\x3f\xb0\xe9\xc5\x6c\x70\xbf\xd0\xba\x3a\xad\x81\x4d\x4a\x4b\x35\xf7\xa2\xc0\xe1\xbb\xb9\xdc\xa1\x1f\xd5\xcb\x2e\x1f\xfd\x9d\xf6\xec\xac\x47\xe8\x22\x93\x02\x65\x06\xd9\x79\x96\x6b\xee\x8b\xe9\xad\x30\x97\xb3\xb8\xb7\x78\xdb\x21\xae\xab\x1b\xa7\xf5\x4f\xbc\xbc\x6f\xe9\xfb\x9e\x9b\x1b\x9f\xcf\x7b\x2a\x27\xb9\xf4\xed\x66\xa5\x69\x33\xbf\x4d\x3d\x2b\x3a\x71\xe7\xa6\x4e\x81\xb3\xbd\x52\x25\x27\x1b\x1f\x5d\xdd\x90\x55\x22\xb6\x4a\x6b\xc6\x51\xa7\x22\xae\x92\x9f\xd7\x1f\x45\x4f\xa8\x9b\x9f\x71\x67\x41\xe6\xb3\x8b\x82\x6b\x6c\x62\x9d\x92\xc3\x3f\x0b\x74\x76\xae\x70\x91\xd1\x6d\x99\xb2\x68\xf5\x4c\x69\x99\x6f\x9e\xe5\xea\xd3\xd4\x24\x4f\x4f\xfa\x6b\x78\x92\xff\x4a\xfa\x85\x28\xdd\xae\x83\x9f\x16\x6c\x3a\xf1\x47\x8a\x7d\xc3\x9c\x67\xdc\x7c\xe6\x05\x33\x5f\xe6\x5f\x6f\x7f\xe4\x7e\xf8\xaa\xad\xf3\xfa\xce\x3b\x8c\xa0\xc0\xe0\x62\x00\x01\x50\x60\xc8\x88\xcd\x5e\xa5\xcb\xc0\xc0\x00\xc2\x5c\x0c\x32\x0c\xf9\xa5\x25\x05\xa5\x25\xc5\x04\xc2\x03\xa2\x4a\x41\xa9\x24\x31\x29\x27\x35\x3e\x2f\x31\x37\x55\x49\xa1\x9a\x4b\x41\xa1\x2c\x31\xa7\x34\x55\xc1\x56\x21\x27\x3f\x39\x31\x47\x0f\x24\xce\x55\xcb\x85\x1a\xfa\xca\x0c\x72\x9b\xdc\x18\x18\x18\xfc\x19\x18\x18\x78\x18\x64\x18\xca\x12\x8b\x32\x41\xa6\x10\xb2\x52\xdb\x5b\xe7\x94\x8f\xcf\x29\xcf\xd0\x80\x53\xde\x67\x7c\x43\x83\x56\x3c\x0b\x0a\xd0\x58\xa9\xe3\x11\xda\xc0\x14\xdd\xfa\x68\xd2\xa2\x33\x7b\x4a\x16\x14\x78\x65\x15\x16\x05\xa6\x39\x7a\xa6\xd5\xc9\x05\x44\x9c\xe4\x5f\x11\xa7\x68\xef\xbb\x50\x45\x39\x74\x21\x4b\x4a\x8b\x50\xb8\xae\x38\x43\x80\x37\x23\x93\x1c\x33\xae\xb4\x20\x01\x0e\x16\x46\x06\x06\x86\x25\x8d\x20\x16\x3c\x65\xb0\x42\x9c\x85\xe6\x24\x88\x61\xb8\xc2\x12\xd9\x30\x5e\x46\x06\x94\x90\xc5\x67\x1e\xae\xa0\x42\x36\xaf\x8e\x91\x01\x2d\xe0\x70\x99\xc8\xca\x06\xd2\xc5\xcc\xc0\xcc\xf0\x1e\xe4\x2a\x26\x10\x0f\x10\x00\x00\xff\xff\x46\x7c\xe0\x1a\x0f\x03\x00\x00")

func terraformModulesDynamodbZipBytes() ([]byte, error) {
	return bindataRead(
		_terraformModulesDynamodbZip,
		"terraform/modules/dynamodb.zip",
	)
}

func terraformModulesDynamodbZip() (*asset, error) {
	bytes, err := terraformModulesDynamodbZipBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "terraform/modules/dynamodb.zip", size: 783, mode: os.FileMode(420), modTime: time.Unix(1627387497, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _terraformModulesFuncsZip = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x98\x67\x54\x53\x59\xfb\xf6\x0f\x90\x84\x66\x22\x89\x14\xa9\x52\x82\x80\xd2\xab\x48\xc9\x48\x2f\x52\xa5\x17\x93\xd0\x3b\x48\x09\x48\x51\x10\x21\xa1\x37\x41\x41\xe9\xa0\x20\xa0\x22\x22\x42\xe8\x5d\x69\x82\x41\xa9\x62\x40\xe9\x5d\x3a\x08\xef\x9a\xd7\x35\xff\xf1\x71\xd6\xf3\xcc\xf9\x74\xce\x87\x7d\xad\xb3\xef\xbd\x7f\x7b\x5f\xd7\x6d\xa8\x4b\x05\x62\x04\x00\x80\x06\x90\x88\x5c\x30\x16\x79\x3d\x1f\xfc\x89\x12\x00\xbe\x9f\x02\x00\x08\xc0\x01\x60\x7d\x5c\x45\xfd\x9d\x4c\x4d\x68\x01\xaa\xeb\x36\xf3\x18\xb4\xcd\x3c\x26\x20\x88\x1e\xa0\x00\x6d\x53\x00\xc0\xff\x1f\xb7\x62\x66\xeb\x3e\x2e\xce\x18\x36\xf0\xfc\x4d\x90\x98\x88\xfb\xa3\xc4\x46\xe8\x3b\x41\x6b\x4b\x6a\x6b\xce\xac\x78\xcf\xbb\xdc\x8c\x87\xb2\x1b\x6e\xd6\x39\x48\xe7\xef\x3a\xc7\x87\xcf\x84\x9f\xa4\xcc\x90\xa0\xc5\xef\x25\x98\xb5\x54\xfc\x51\xfe\xab\xf2\x43\x31\xfc\x17\x8c\xba\x19\x5a\x66\x50\x4c\x33\x37\x7b\x67\x6d\x66\x0c\xae\x4b\xc6\xe8\x3c\x04\x27\xa9\xcc\x7e\x7f\x4d\x9d\xbc\xc1\x48\x66\x44\xc0\x37\xdb\xe5\x50\xc1\x65\xd7\xd7\xc0\x45\x8a\x03\x33\x1e\x30\x5a\x03\x67\x38\x7e\x89\x97\x3e\x40\x58\xf2\x3d\x65\x09\xe3\x62\xa2\xd2\x68\xb5\x39\x53\x7e\x83\xaf\xd0\xad\x74\x5c\x8d\x3b\x0f\xeb\xa6\x0f\xc5\x96\x95\x77\xdd\x6a\x85\x60\xa3\x46\xf7\x34\x61\x7a\x97\xbc\xdb\xa7\x90\x28\x88\x9a\x43\x48\x79\xbe\x24\xe7\x87\x0a\xdd\x6d\x54\xed\x58\x74\x6b\xc0\x90\x54\xcc\x87\x21\xf7\x9b\xbe\x6a\x47\x34\xd2\x36\x63\x80\xb6\xaa\x0d\xf0\x46\x4a\x6c\x89\x96\x4b\xa7\xa1\x96\xba\xdf\x03\x1a\xa7\x50\x0d\x5d\xa7\x8e\x53\xe0\xe0\x88\x17\x5e\x8d\x32\x02\x13\xf8\x0a\x9d\x9d\x8b\xb8\x92\xa2\xf3\x08\x59\xfe\x9c\x03\xa3\x76\xf8\xa0\x9f\x54\xde\xbb\x5b\xb6\xa3\x35\xd6\xbb\x7d\x1f\xa9\x16\x17\x8d\x56\x10\xc5\x92\x8b\x97\xaa\xbb\x8e\x63\x6e\xdb\xd9\xc9\x3c\x7c\x4f\xeb\x18\x67\x76\x26\xa0\xc4\x2d\x55\x95\xb5\x0c\x46\xbe\x46\xa2\x87\x4f\x8a\x1d\x5e\x77\x78\x33\x87\x41\x75\x8b\xda\xb8\x50\x40\x2c\x5c\xb2\x3f\x8d\x08\x5f\x69\xc3\x48\xef\xc4\xdf\xcb\xed\xef\x62\x69\x51\x90\x0a\xb8\x62\xc4\x7a\x32\x3b\xac\xed\xcd\x39\xa0\x3d\xbc\x09\xe9\x29\xf5\x96\x24\x43\x83\xbc\x23\x77\xef\x9f\xeb\xa4\x38\x54\xbc\xdc\x64\xd7\xc6\xbd\x77\x69\xf3\x8b\x4c\xd0\xfd\x96\x6f\xb5\x73\xd2\x6a\xae\xda\xa1\x2f\xc8\x3b\x0f\x1e\x42\x4c\x3b\xbf\x06\xe9\xb4\x3d\xe0\xac\xd7\xa7\x9d\xbe\x3a\x63\x63\x66\x9f\x0c\xdb\xa1\xf3\xa6\xc5\x0a\x55\xf3\xbc\xd6\xf8\x6a\xd3\xa1\x39\xd1\x31\x1b\xb1\x90\x31\xc2\xda\x50\xac\x4f\xfc\x08\xb1\x4a\xd7\x4f\x69\x59\xba\xda\x97\xf5\xa3\xfc\xe3\xbd\x82\x9e\x89\x5a\x6c\x02\x11\xb7\xe3\xd8\x1e\x6d\xde\xf2\xc9\xfb\xfc\xf3\xe4\x8d\x67\xaf\x2a\xec\xd3\x45\x3a\xa0\x92\xa2\x52\xf6\x99\x87\xf2\x3c\xfb\xef\xa4\x58\xb1\xd4\xfb\x7a\xec\x5b\x5e\x7d\x6b\xf7\x2f\x7a\xf7\xec\x1d\xc8\xec\x09\xe0\x50\x6e\x07\x06\x59\x53\xfa\xc5\xbb\xf9\xd3\x4d\xdb\x4a\xd5\xdb\xdb\x81\x84\x8d\xf8\x63\x4d\x0c\x97\xf3\xed\xbe\xd6\xa6\x77\x3f\xe6\x63\x24\x5a\x15\xe1\x8e\x3a\xa2\x24\x9e\x81\xf3\x6f\xc3\x20\xc5\xa6\xbd\xbc\x1d\xbd\xfc\xca\x51\xda\xf8\xc7\x8d\xa6\xcb\x25\x1d\x46\x0a\x4f\x8c\xc3\x0e\x68\x38\xab\xe3\xe7\xfc\x36\x19\x75\x64\xbf\x5c\x8d\x84\x7a\x58\x99\xd0\x1f\x2f\x7e\x21\x22\x58\x9d\xf2\x09\xeb\x98\xa1\xaa\x2d\x7f\x5c\x4e\x4c\x89\xe2\xd3\x33\xe3\x5e\xa7\xc5\x95\x14\x8b\xd6\x98\x1a\xd2\xf6\xd8\xda\x24\x91\x21\xd5\x44\x3c\xa2\x94\xe7\x21\x4f\xb6\x2d\x92\x32\x3d\xa3\x00\x79\xd6\x8a\xd9\xe2\xbe\x60\xdf\x5c\xf7\xa2\xc0\x3e\xff\x7d\x27\xcd\xd7\x96\xcb\xc7\x21\x59\x37\xa2\xca\x0f\xe8\x4b\x72\xf7\x6d\x1d\x6d\x4e\x26\xf0\x44\xad\x22\xc5\x7e\x1b\x41\x5c\x31\xad\xaf\x66\x19\xb4\xcd\x54\x8a\x5a\xb1\x99\xb3\x95\xae\x3d\xb8\x91\x68\xca\xb6\x76\xfa\x6f\x2c\x44\x6e\x7e\x33\xa6\x6c\x68\xcc\xdf\x07\x00\x60\x95\x02\x00\xa8\xff\xc4\x22\xc0\xdf\xe5\x2f\x2e\xa2\x41\x33\x18\x3c\x68\xe6\x77\x2e\xa2\x92\x5b\x74\xa8\x24\x18\xa2\xd7\x6e\xc6\xde\x68\xc0\x07\x33\xc5\x12\x08\x34\xd3\xb1\xc3\x69\x5a\x67\xb3\x2a\xd9\x5d\xd2\x02\x24\xc7\x5f\xc6\xae\xad\x3a\x10\x87\x8a\x19\x6d\xba\x1d\xe7\x7a\x4e\x7e\x64\xfa\x59\x08\x6c\xbd\xb8\x5a\x76\x16\x49\x81\x1b\x42\xc2\xd0\xed\x0c\xca\x98\x38\x8f\x9e\x73\x34\xac\x92\xb4\x82\xdc\x50\xbd\x7c\x02\xbb\x24\x17\x38\xbf\xd5\x16\x67\x1f\xbf\x24\x59\xe9\x7e\xcd\x6a\x4e\x87\x3a\xd6\x7c\xdd\xf3\x0f\x11\x5b\xc6\x07\xe7\x7c\x07\x66\x0d\x4d\x7d\xc0\x2c\x3e\xb4\x52\x60\x1d\x8e\xab\xc0\x77\xfb\x53\xd7\xf4\x39\xc5\x5c\xbd\x9b\x65\xdf\x30\xd1\xad\xbc\x08\xc4\x73\x09\x08\xe3\xd7\xd0\x53\x8c\x93\x73\xf3\x5a\xd5\x92\x23\x77\xce\x69\x76\x1d\xa0\x5a\x33\xce\xd4\x7f\xa6\xcd\x29\x00\x2d\x4c\x8e\x58\xcc\xc8\x16\x71\x36\x38\x1e\x2c\x33\xd6\x40\xf3\x05\x5e\x7f\xb2\xf7\xbb\x97\x6c\x57\xa9\x3d\xc5\xc5\x52\x5b\xba\xc6\xb5\x86\x76\xdb\xad\x7b\x13\x86\x33\x08\xf8\xf1\x42\x10\xd3\xe3\x5a\xc6\x86\xeb\x96\x38\x8d\x19\x6d\x54\x6a\x38\x22\x0f\x7a\x57\x05\x5a\x63\x0a\x7b\x8e\xc0\x7f\x57\xce\x21\xec\xc0\x18\x2f\x3e\x15\xe9\x44\x05\x00\xe1\xb0\x9f\x07\x8a\x2b\xd6\xf3\xaf\xc2\x2d\x13\x8f\x31\xab\xc4\xe3\xdf\x0b\x37\x6c\x6e\xeb\x3d\x26\xce\x18\x36\x67\x6e\x81\x35\x28\x89\x18\x0a\x33\x89\xab\xa3\xd6\x20\x96\x0e\x8d\x99\xfa\x5b\x9d\x7b\x4a\x57\xdb\x5f\x6c\xa1\x67\xe7\xa4\xf0\x02\xb6\x31\x15\xe8\x49\xad\xa5\x45\xe9\xbf\xc8\x95\xa5\xc3\xbf\xa1\xb4\x3b\xfb\xd9\x0f\x76\x23\x39\x4f\xed\x02\x2b\x43\x52\x3f\x29\xf9\x43\x90\xa0\x3a\x88\x6b\x83\x97\x6d\x0a\x4f\xe2\xfb\x0e\xe1\xa1\x4d\x88\x60\x4c\x32\x68\x0f\x91\x4a\x02\x15\x1a\x8d\xab\xb1\xdc\x8a\xd6\x7c\xf5\xf8\xf4\xd3\xbb\x39\x69\xba\xd7\x8a\x44\x57\xc2\x1f\xc5\x3e\xd8\x04\x51\x5c\xbb\x30\x9e\x4f\xb4\xc0\xc8\x3c\x35\x61\x8b\x5c\x71\xbb\x04\x87\x3b\x1b\x10\xe4\xe5\x37\x06\x1a\x6f\xb7\x94\x2c\x24\x32\xac\x09\x68\x75\xa8\x4f\x83\xdb\x15\x2e\x2f\x73\xed\x4f\xaa\x36\x85\x2c\x73\xbe\xf9\x52\xfd\x98\xcd\x4e\x4b\x3a\xda\x9b\xfd\x59\xfe\x4d\xfb\xd1\xac\x4f\xf6\x97\x75\x4d\xad\xc0\x4a\x21\xc4\x08\xf2\x17\x4c\x42\x6d\xaa\x22\x36\x54\x85\xef\x6e\xc5\x35\x3d\xcf\xe4\x6b\xfd\xcd\x0e\x44\xed\x31\xe9\xc7\xb7\xd5\xee\x38\xbc\xce\x24\x34\x30\xe1\x52\xca\x70\xb2\x6c\x14\x52\x93\x41\x0d\x82\x88\x2a\x7e\x17\x2c\x6a\x8d\x6d\xb5\xb7\xb7\xc6\x45\xda\x41\xfb\x5c\x8e\xa0\xde\x99\x68\x0f\x2d\x89\xe9\x61\x2b\x4f\xa3\x2e\x86\x98\x8d\xfa\xb1\x95\xf2\x00\xaa\x8a\x46\xe7\xc4\x1f\x6e\xee\xcb\x75\x3c\xa5\xb5\xfc\xa7\xe1\x46\x02\x4b\x30\xac\x7b\x91\xc8\x3b\x6b\xd5\xfd\x96\xf3\x98\xe4\x3e\x0f\xee\x0a\xdf\x20\xcf\xd4\x8c\xfb\x71\x22\x2f\xbf\x43\x07\xb3\x1c\x50\x74\x78\x71\xe3\xe8\xce\xf3\xa3\xea\x45\xde\xc9\xcc\x35\x71\xfa\x91\xa3\xc8\x57\xae\x42\xae\xa2\xf8\x52\xbe\x00\xb0\xe0\xd0\x98\xcd\xc0\xc3\xb2\x97\x51\x63\x08\x87\xb4\x09\x04\x46\x7d\xcd\x53\xc8\x69\x86\xc1\xa8\x58\x5e\x0a\x67\x6e\x40\x17\x42\x7f\x00\xef\x3a\x86\xa8\xcc\x73\xd9\xe7\x99\xa8\x0f\x57\x64\xb2\xf0\x76\x6c\xe4\x64\xfa\xf3\xd5\x42\x9e\xdf\xf3\x33\x97\x30\x42\x23\x79\x06\xa2\x2f\xa4\x39\xf1\xe4\x46\x11\xdc\x0a\x99\x23\x10\xd4\x9f\x9a\x8d\x69\x51\x48\x3d\x92\x9a\x56\x4f\x55\x86\xab\x3a\xcb\xac\xc5\x3d\xb3\xb8\x8b\x4b\x65\x0a\x21\xb3\xcf\xb2\xd5\x73\x66\xac\xb7\x71\x06\xe3\xbc\xeb\xc3\x75\xb8\xfb\x42\xc3\xe3\x79\xd5\xb9\xac\xb1\xc7\xb7\x2d\xd5\x5b\x05\x90\x36\x77\x77\x43\xe8\x5e\xde\x69\x22\x08\x2b\xa8\x95\xa6\x20\xc5\xba\xcc\x78\x16\x91\x16\x50\x5b\xc6\xbc\x3f\xc2\x38\x17\x74\x5b\xdd\xe7\xbd\x10\x72\x4a\x2d\x9f\x92\x16\xee\xa6\x80\xb8\x4d\xf6\xd0\x03\x77\x43\xdb\xd5\xa1\xbe\x6a\xac\x16\x5f\x01\x0d\x6e\x8a\x9a\xe9\xf0\xf8\xea\x9c\x92\xbc\x0f\x2d\x15\x99\x4a\xb1\x42\x77\x7c\x98\xbf\x43\xde\xef\xa7\xdf\x91\x54\x2d\xe9\x27\x7a\xed\xee\xed\xcd\xe0\x77\x39\x91\x0a\x69\xaa\x49\x2d\x5d\x9e\xa1\xef\x54\x69\xc2\xf6\x12\xd0\x01\x54\xdb\x09\x74\xb7\x04\xff\xd8\x8f\x79\x7f\xfc\x40\x1f\x0e\x96\x9c\x7d\xe7\x53\x11\x2a\x12\x4a\x92\xd9\x54\x35\xbc\xba\x79\xf8\xa3\xe9\xf5\xa5\xce\x1c\x59\x89\x6e\xf3\x15\x06\x8c\x66\x4b\xe9\x0c\xb5\x68\x76\xf0\x9c\xee\x86\xac\x1c\x3d\xfd\xaa\x6e\xb3\xbe\x6c\xbe\x69\xf7\x35\x02\x06\x5d\xe4\xfe\x45\x31\xc4\xb1\xac\xfe\x86\x41\xa1\x73\x67\xb1\xaf\xd8\xf2\xd2\x4b\x55\xb3\x59\x4c\x8a\x56\x1b\xeb\xb0\x47\x3d\x16\x16\x42\x92\x47\xde\xff\x76\x35\xdc\xa9\xff\x9b\xff\xc4\xd3\xaa\xf9\x76\x17\x07\xb3\x81\x10\x16\xdd\x9a\xf8\xbb\x48\x92\xd7\xa4\x1a\xfd\x90\xdc\xb0\x97\xef\x30\x87\x60\xd9\xce\xed\x50\x53\xe7\x6c\x2d\xc5\x92\x5e\xd8\x54\x7b\x9b\xb9\xd0\xd0\xdb\xc7\x5e\xea\xdd\xd5\x9d\x55\x6f\x10\x7e\xe9\x6e\xa0\x47\x7a\x95\xf0\xea\xdc\x22\x93\xf3\x11\xa3\xbb\x9e\xc1\x3d\x47\x4c\x50\xf4\xc1\xc6\xfa\x00\xa4\x5b\x89\x37\x9a\x94\xf5\x7e\xb3\x3a\xd8\xea\x01\x2c\x23\xb1\x9b\x8b\x4a\x3f\x71\x6c\x0a\xe6\xfe\x2d\xfa\x92\x06\x6c\x81\x79\xdc\x7d\xe0\x95\x84\x97\xc2\x21\xfd\xdf\x94\x1a\x84\xee\x18\x83\xd8\x25\xa5\xe3\xc0\x00\xc0\x07\xfb\x79\xbe\x79\x62\x5d\xbd\xfe\xc2\x74\xd6\x6f\x1f\x33\xe7\xb7\xff\x3b\xa6\xe9\xe6\x9e\xde\xe3\xb2\x0c\xc7\xcf\xd0\x37\xd2\xe8\xdb\x45\x39\xe4\x3d\x5f\xe0\xe8\x93\xa1\xa3\x59\xcc\x14\x97\xfd\x5d\x0e\x4b\x97\xb8\x0d\x11\x22\x7e\x8f\xb5\x22\x8b\x1e\xbd\x89\x38\x39\x5c\xd5\xd2\xc9\xc8\x2a\xd2\x83\x22\x09\x6f\x89\x73\x9f\x73\xed\x8b\x44\x33\x5f\xf9\x3c\x04\x04\x3f\xfa\x5c\xc2\x6f\xbd\xad\x20\xb3\xbf\xd5\x51\x9e\x94\x4c\x33\xad\xc3\xaa\xf4\xf1\xf4\x19\x1f\x05\x3c\xea\xdf\xe9\x32\xf0\x5a\xb8\xe3\xa9\x76\x5d\xbc\xe7\x3b\xd4\xaf\x2c\x2c\x8b\xbb\x15\x24\x1c\x68\xb8\xc1\xef\x5e\x5c\xd5\x6c\x02\x66\x6f\xb7\x8b\x6e\x7b\x69\x5a\x7c\xc6\x8a\xc2\xaf\x9f\x50\x10\x30\x51\xa0\x12\xe8\x36\xb8\x2a\x22\x5c\x1d\xb1\xcd\x91\xc1\xcd\x98\x1c\xd5\xb6\xc2\x5a\x4e\x3e\x92\x33\x01\xd6\x84\xc2\x96\x5d\x1b\x67\x15\xea\x6f\x38\xc9\xd9\x3a\x91\x2f\x06\x3f\xe8\x2f\x08\x77\xb1\xe8\xd3\x94\xfc\xf1\xd6\xb0\x5c\x2a\xf5\x70\x7d\x40\x30\x54\x5b\xfe\xed\xbc\x42\x47\xdd\xe7\x27\xeb\xef\x9d\x2b\xf3\x80\xfb\xd3\x4e\x0c\x31\x5d\xb5\xcc\x5e\x60\x9e\x80\xc1\x44\xd7\xf4\x33\x9f\x5b\x9b\xe1\x7b\xb4\xdf\xb3\x5b\xd7\xf2\xac\x84\x42\x50\x85\x82\xc1\x96\x4b\xce\xa3\x7a\x97\xf9\x14\x94\xcb\xf4\x77\x2e\xb5\xaa\x54\xf0\xed\xd0\x49\x89\xd4\xde\x3c\x85\x63\xcc\x16\xb0\xa5\x89\xfe\x74\xe4\x94\x75\xeb\xf8\x68\x51\x88\x79\x8f\x82\xda\xc9\x27\x11\xa3\x63\x7e\x84\xd2\x4d\x2e\x35\x64\x6a\x5a\x68\x36\xd6\x61\x8f\xca\xd2\x20\x8d\x46\xd2\x1e\x46\x18\xb0\x54\xf9\x87\x85\x92\xb8\xbf\x82\xc0\xb7\xa5\x7d\xc9\xdd\x53\xc3\x6c\x89\xdd\x49\x85\xf3\x9a\x6d\xec\x9e\xb8\x5a\x26\x74\x50\x6e\x78\xcb\x28\x47\x8e\xd2\xc9\x37\x62\x23\xec\x88\xa6\xee\x8f\xba\x41\xca\x62\xed\x9a\x67\x9a\x49\x53\x73\xeb\xd3\x5f\x37\xb4\xcf\x5a\x10\x0e\xe0\xd9\x65\x22\xbc\x72\x13\x24\x2c\xb3\x7b\x29\xfd\x79\x3f\xc5\xd5\x0b\xdf\x4e\xea\x27\x07\x1a\xb1\xa7\x06\x95\x28\x41\x9e\xa8\xf4\x1f\xd2\x07\x66\xbb\x38\x06\xb4\xb8\xe5\x5e\x65\xd1\x69\xbd\x80\x73\xcc\x61\x99\x6d\x07\xf8\xa5\xcb\x84\xde\x19\xec\xe4\x0b\xa9\xc2\xa8\x8f\x86\xd6\x8f\x2f\xce\xc9\xee\x78\x54\xd2\x94\x23\x9c\x9c\x9a\xde\x24\xd9\xaf\xfb\x2c\xc3\xf4\xcf\xeb\x61\xac\xb5\x8c\xb8\xec\x78\x27\x84\xaa\xa9\xce\xfb\x46\x56\x0f\xea\xac\xd6\xb9\xe6\x86\x5b\x50\xf6\x0f\xea\xb0\x0b\x17\xf0\xe0\x09\x2f\x35\xe4\x5b\xa4\x8a\x2b\x47\x7a\xcb\x6f\x82\x7b\x71\x10\xab\x9e\x7c\x52\x35\x42\xff\x34\xcf\xe3\x62\x72\x6f\x3e\xf2\x25\xf3\xc1\xc3\x5a\x47\xe7\xf8\xbb\xe2\x2a\xe7\x68\x7e\xdc\xe4\xa2\xf6\x8d\x89\x0f\x71\xe9\x0e\x57\x3e\xfd\xa4\x6a\x30\x80\x7d\xcc\x23\xf7\xfd\x80\xb2\xf6\xf9\xf5\x8d\xc9\xb2\xc0\x11\x78\x70\x4f\xbe\x4b\x8f\x4c\xdd\x6a\x45\xda\xb8\x99\xdc\x51\x70\x07\x54\xb3\xd2\x1d\x16\x5b\xa8\x3b\xbf\x10\xeb\xb9\x1f\x9e\xcd\x7b\x2e\xc4\x9b\x4f\xa5\x45\xc2\x92\x63\x0f\x2a\x66\x2d\x60\x02\xf5\x8a\x12\x72\xf1\x1f\xd6\x69\x4d\xb5\x19\x9e\xaf\xbf\xae\xb1\xaf\x28\xc9\x65\xc7\xf2\x08\x33\x11\x1b\x89\xa5\x53\xf0\x08\xa6\xec\xf2\x39\xf0\x8d\xb6\x8a\xb6\x9e\xcc\x4a\x55\x7d\xf7\xcc\x33\x71\x99\xe3\xde\x35\xad\x10\xb4\xf1\xd7\x4d\xda\x61\x87\xf8\x25\xb9\x64\x4f\xfa\xa5\x5e\x17\x8a\xec\x5e\xf2\x16\xd8\x2b\x97\x74\x51\xc9\x6a\x4e\x7e\x61\x4b\xb6\x50\xf6\xe2\xb7\x3c\x5d\xa3\x10\xfb\x76\xe6\xb8\xb3\x5e\x9c\x65\x2b\xeb\x3b\xcc\xd7\x6f\x2a\x1e\x7d\x07\x12\x99\x78\x2c\x85\x32\xea\x27\xd2\x53\xe7\x20\x1f\x98\xf8\x0a\xde\x3a\x46\xb5\x0b\xe7\xe2\x2f\x44\xab\x16\x9b\x8e\xb2\xb5\xa5\xf7\xd4\x4c\x58\x58\x5d\x86\x9d\x71\x8e\x7a\x8e\x7e\x42\x67\x58\x39\xd9\x7d\xdc\x73\xdc\x79\xd9\x93\xf0\x59\xf7\xad\xf0\x4a\xef\x56\xdf\x78\x1c\xea\xb9\xe8\xd4\x13\x23\xde\x3d\xb5\xa6\x4a\xeb\x4e\xfe\xe4\x72\xa1\x0b\x19\x81\x99\x7b\x75\x0c\x7f\x80\xb2\xfc\xcc\xec\x1e\x42\xcc\x2e\x95\xa3\xad\x56\x34\x74\x83\x39\x18\x52\x31\x7a\xfa\xe8\x2a\x20\x53\xa2\xb0\x75\xb3\x6b\xf0\xe9\xfa\xb7\xbc\x40\x68\x34\x81\x0a\xfd\x6c\xd0\x9c\x4d\xa0\x37\x43\x29\x81\x5f\xf7\x68\x31\xb1\x22\x10\xe6\x20\x31\x2e\x39\xba\xf9\xa6\xf5\xb0\x5a\xbd\x54\x59\x2b\x67\x5f\x26\x63\x2b\x64\x4a\x60\x16\xfb\x8d\x28\x07\xff\x32\x9e\xb8\x39\x74\x37\x50\x75\xa8\xe1\x9c\x15\xdf\xea\x48\x1d\xd1\x6e\xad\x61\xd9\x36\x11\xe2\x5b\x33\xbe\xbe\xc6\x53\x6c\xee\x7f\x0f\x36\xc1\x89\x32\xd4\x58\x2c\x2c\x73\xa8\x5f\x6b\xdc\xbc\x19\x24\x93\x7f\xe7\x68\x5b\xd3\xd7\x02\x6d\xfb\x04\xca\x18\x6d\x7e\xbd\x65\xd3\xb8\x8c\x2e\x26\x62\xb2\xf9\xb9\x80\x80\x78\xd8\x8e\x94\xb4\x0a\x7c\x34\xe0\x70\x9f\xeb\x01\x79\x18\x29\xaf\x18\xb9\x0a\xe7\xbc\xae\xf7\xa5\x1b\xa5\xe4\x92\x86\xc2\x71\xfc\x68\x64\xd4\xb8\x88\x1e\x5e\xd7\xdf\xb2\x86\x0a\x0c\x26\xea\xe5\x63\x10\x41\x01\x94\x11\x26\x58\xa2\x90\x84\x4e\x97\xfb\xcc\x02\x09\x4d\xb6\x4c\xbc\x22\xa6\xbe\x42\x56\xb4\x6c\xe1\xef\x17\x6d\xa4\x18\x53\xc1\x5a\x8f\xa5\x7b\x39\xaa\xab\x69\x99\x11\xce\x24\xe4\xf2\x3c\xfe\xb2\xb9\x4d\x8d\x34\x8a\xa7\xa5\x6a\xde\x5e\xba\xfc\xf4\x95\x49\x67\x9c\xc8\x99\x54\x26\x15\xba\x9a\x48\x03\xab\x8d\xf2\xae\x93\x0f\x2e\xee\x49\x62\xca\x60\xc7\x50\x75\x3e\xb9\x0f\xca\x63\x3e\x1f\x68\xf0\x71\x8f\xcf\x40\x45\xbf\x16\xbc\x8c\xfc\x7a\xa4\x1a\x67\xb0\x67\xab\x56\xbf\xa4\x3e\x1d\xdb\x92\x83\xc8\x2c\x39\x5e\x21\xa5\xac\x42\xb3\x06\xdc\x0b\xc7\x13\xf0\x57\x02\x9e\xcb\xd4\x2d\x74\x2a\xba\x75\xaa\xb7\x17\x91\x3e\xf9\x6c\xec\x32\xd5\x5f\xf9\x5c\xef\xa4\x10\x6b\x5b\x52\x71\xdd\x2f\x68\x89\xe9\xa0\xb3\x21\x97\x10\x7a\xe9\x23\xe4\x42\x81\x12\x9a\xfa\x91\x62\xf8\x04\xb9\x28\xe4\x92\x47\xaf\x21\x4a\x5c\xe0\x63\xbe\xc8\x22\x22\x7e\xec\xc3\x09\x58\x3d\xb2\xa5\xdd\xe2\xd9\x1d\x7a\x5a\x11\xba\x81\x3e\xbc\x3a\xdc\xe6\x12\x3d\xe2\xb9\x82\x64\x31\x8f\x9c\x4a\xba\x5d\x9b\x83\x40\x8c\x44\x47\xb2\xe4\xc7\x17\x62\x56\x58\x90\x65\xe9\x43\xd2\xdc\x09\x1f\x4e\xa5\x2e\xa8\x2e\x28\xab\x55\x83\xdd\x1b\xa4\x30\xf0\x89\x28\x2c\x9f\x7a\x38\x34\x0f\x19\x77\xc1\xab\x26\x18\x24\xd0\xcd\xa3\x05\x8e\x3e\x7e\x1b\x8a\x7c\x7f\x9b\xb5\xb2\x2c\x3a\x3d\xe8\xf6\x3e\x99\x8f\xd3\xb0\x10\xb4\xc0\x21\x7f\x59\x58\x5b\x67\x46\x7c\x9c\x30\x37\x1f\x9e\x63\x65\x19\x4a\x89\x29\x30\xbc\xb5\xd6\xd0\xe6\x76\xf2\x9b\x57\x0d\xed\x8b\x41\xf8\x03\x00\x90\x09\x00\x00\x1d\xc0\x01\x78\x07\xf8\xfb\x04\xf8\xfb\xfd\x6a\x57\xa3\xff\x69\x57\xd5\x13\xae\xc0\xda\xb8\x4f\x81\xd6\xea\x5b\x79\xe5\xd3\x58\xb6\x44\x32\xd2\xeb\xe0\x4f\xa9\x04\xcf\x16\xd0\x6c\xec\xa6\xaa\x8e\x2f\x95\xe3\x9c\xee\x58\x48\x33\x81\xad\x61\x95\xef\x7b\x41\x40\x7a\x91\xb8\x62\x73\x7f\x42\x7f\xfc\x3d\x45\xaf\xa2\xd0\xa1\x28\x03\x7d\xc4\xf3\xec\xc8\xba\x88\xec\xd3\x32\x8d\xac\x5d\x35\x62\x7e\x65\xfe\x13\x8c\x55\x43\x9d\x67\x5d\xe6\x1a\x47\xb0\xcb\x9d\xa1\x17\x9f\x23\xcc\xb6\x25\xbe\x12\xb7\xc5\xb4\x6d\xb4\x27\xbc\xa5\x27\x43\x81\xff\xbc\x83\x7a\x50\x57\x73\xee\x52\x00\x80\x2e\x08\x00\x4e\x01\x1c\x00\x0e\xeb\xeb\x8a\xb5\xf3\x70\xf4\xfb\xb7\x8b\x28\xad\xc5\x2b\x52\xfc\x54\xf4\x9a\x62\x54\x47\xe9\x78\x84\x6e\xb5\xfd\xe9\x0f\xe7\x57\x1a\x95\x01\xc9\x96\xf7\x2e\x3c\x1a\xda\x2e\x5d\x8f\x1b\x0e\x1f\xbd\x2a\x3b\x9b\xa7\xd7\x47\xdb\xb5\x77\x82\x62\xf7\x53\x7b\x52\x29\x09\xba\x29\x65\xc1\x24\x6d\x9d\xdb\xe6\x0c\x31\xbc\x16\x19\x49\xcb\x30\x00\x53\x6e\x1a\x4a\x37\xd7\xdf\xdd\xe0\xaf\x7d\x97\x5b\x3a\x68\xf6\x84\xc0\x2e\x3e\x62\x2a\xd3\x6a\x5a\x1c\xae\x04\xf7\x2f\xf1\x84\x23\x7b\xe8\x12\xf0\xd1\x9c\x18\x1a\x7d\x73\x55\x6e\xf4\xd7\x51\xf3\x0c\xce\xe7\xa0\x0a\x32\x9d\x9e\xf6\x4e\x90\x06\x8d\xbd\x81\x1d\x09\xdb\x74\x45\xe1\xb3\x5e\x1e\x4e\xe8\x74\x41\x57\x4a\x91\xca\x7d\xdd\xfa\xd6\xbd\x54\xe4\x15\x64\xb5\xfd\x78\x0a\x74\x7d\x26\x30\x5f\xc2\xa9\x55\x91\xea\x69\x85\x18\x6a\x64\x08\x34\xe7\x8b\xfb\x52\x4c\x3f\x2e\xca\x89\xb6\xb4\x0a\x19\xa1\xab\xee\xc8\x70\x23\x6c\xfb\x5c\x4e\xe6\xa3\x39\xf5\x0e\x0e\xa1\x5c\x96\x79\xb5\xe9\xd3\x0a\xe7\x10\xc2\xc9\x78\xcc\x73\x12\x3f\x1b\xe6\xbf\x21\x3e\xf4\xf3\x79\x34\x23\xc2\x38\x50\x4a\x29\x8c\xf3\xf7\xf2\xb3\x7b\x93\x92\xb7\xae\x6a\x7a\x44\x2e\xe9\x8e\x42\x4c\xea\xe5\x18\x18\x17\xb2\x57\xd8\x45\x0e\x4a\x71\xb8\x0c\xd3\x23\x03\x4f\xce\x85\x93\xd2\x95\xbb\x41\x3d\x23\x6d\xa8\xb3\x1f\xf4\xfa\x3e\x94\x64\xf8\x27\xa5\xd1\x22\xc9\xd2\x8f\xf6\x59\x92\x5e\x33\x05\x5f\x08\x39\x85\x9f\x7a\x76\x31\x91\x57\xa1\x6c\x77\xfc\x25\x1b\x89\x3d\x8a\x9d\xb0\x82\xe9\x4d\x0d\x6c\x8a\xaa\x1a\x72\xbc\xf9\x4c\x97\x4e\xd0\x48\xaf\x81\x9c\xb2\x08\x3e\x11\xef\x3a\x0e\xb1\xb8\x18\x19\x3c\x6d\x76\xec\x16\xdd\x92\x16\x99\x33\x5a\xee\x6d\xe4\x6f\xf7\x04\x25\xfe\x91\x3f\xde\xdb\x50\x97\x82\x92\x8b\xea\xbf\xb5\x11\xce\x02\x7f\x3e\x14\x00\x00\x14\x45\xfc\xf9\xf6\x57\x53\x01\xfc\xb3\xa9\xf0\xdb\x7a\xfe\xaa\xf5\xcf\xec\xf5\xab\x16\x0b\x15\xf0\x77\x12\x03\xff\xdc\xda\xff\x43\xec\x9f\x71\xe4\x57\x31\x23\x10\xf0\x7f\xe1\x04\xfc\x33\x9c\xfc\x0f\xad\x7f\x9a\xa6\x5f\xb5\x76\xa9\x81\xbf\x2d\x14\xf8\xe7\xce\xfd\x97\x59\xfe\x27\xb5\xbf\x8a\xb5\x41\x81\xff\x60\xf8\xdf\x27\xfa\x4f\x9a\x7e\xd5\x0b\x81\x01\xbf\xb1\xf5\xdf\xfe\x10\x0c\xf9\x73\x14\x04\x80\x00\x24\x0a\x00\x50\x67\xf8\xf3\xeb\xff\x05\x00\x00\xff\xff\x94\x6b\x2c\xf9\x2f\x12\x00\x00")

func terraformModulesFuncsZipBytes() ([]byte, error) {
	return bindataRead(
		_terraformModulesFuncsZip,
		"terraform/modules/funcs.zip",
	)
}

func terraformModulesFuncsZip() (*asset, error) {
	bytes, err := terraformModulesFuncsZipBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "terraform/modules/funcs.zip", size: 4655, mode: os.FileMode(420), modTime: time.Unix(1627387497, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _terraformTemplatesMainTf = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x55\x4b\x8f\xe4\x34\x10\xbe\xe7\x57\x94\xb2\x73\x58\x10\x93\x5e\x34\x42\x5a\x8d\x34\x07\x60\x41\x20\x60\x67\xb5\x33\x5c\xb8\x44\x8e\x53\xe9\x98\x71\xec\xc8\x8f\x9e\xed\x8e\xf2\xdf\x91\x1f\x71\x92\x4e\x37\xd2\xe6\xd2\xed\x7a\x7c\x5f\xe5\xab\x72\x85\x4b\x4a\xb8\x86\x21\x03\x20\xaf\xba\x54\xb8\x67\x52\x40\x78\x1e\x20\x47\x7b\x4b\x51\x18\x45\xf8\xed\xf7\x39\x5c\x7d\xde\xc0\xf3\xe3\x87\x47\x88\xd9\xaf\x2d\x2a\x04\x85\x5a\x5a\x45\x51\xc3\x2b\xe3\x1c\x2a\x04\xaa\x90\x18\xac\xe1\x2d\x7e\xa1\xd8\x1b\xa0\x5c\xda\xba\x51\x52\x18\xa8\x99\x36\x8a\x55\xd6\x84\x7c\x46\x5b\x60\x1a\xf6\x5c\x56\x84\x7f\x13\x6b\xeb\x95\x6c\x18\xc7\x54\xdb\x30\x14\x8f\x6a\x4f\x04\x3b\x11\x97\x56\x7c\x24\x1d\x8e\xe3\xa6\xca\x58\xdb\x94\xdd\x48\x05\x56\x23\x30\x01\xfe\xdd\x1d\x34\x50\xce\x32\x80\x5a\xe8\xf2\x24\x05\xa6\xd4\x0b\x24\x1f\x3e\x3e\xfd\x23\xc5\x86\x67\x12\x40\x5a\x83\x3f\xdc\x41\x47\x04\xd9\x63\x0d\x1e\x2d\xc8\x51\x0b\x0d\x0a\xa9\x54\xf5\x46\x10\x47\x2d\x3b\xc2\xc4\x02\xf0\xab\xa9\x49\xcf\xc0\x2a\x9e\x01\xf4\xc4\xb4\xab\x88\x80\x15\xe5\xc9\x00\x28\x2a\x53\x12\x25\xfe\x8f\xec\x67\x54\xe6\x47\x25\xae\x90\x69\xcd\x3d\x0a\x6b\x18\x25\x26\x88\x6a\x5a\x84\x6f\x8b\xf8\x22\x6f\xa7\x5e\x33\x01\xa4\x3e\x10\x41\xd1\xb5\xb1\x57\xf2\x5f\xa4\xa6\xac\x2c\x7d\x41\x93\x98\x7f\xf2\xc7\x0b\xbd\xdb\x32\xc7\x4c\x47\x18\xc1\x80\x4a\xd1\xb0\xbd\x55\xbe\xf0\x9d\x36\xc4\xe0\xae\xb1\x82\xba\xa3\xbe\x56\x89\x21\x15\xc7\x52\x90\x0e\x57\x1a\x3c\x3b\xf3\x42\xaa\x19\xe7\xc1\xdf\x11\x80\x61\xb8\x05\x45\xc4\x1e\xa1\xf8\x75\x72\x8e\x63\x74\xc5\xcc\x14\x0c\xa0\xef\xca\x17\x3c\x46\xf4\xa7\xbb\x3f\xf0\x18\x80\xdd\xc3\x3a\xb2\xc7\x85\xfb\x77\x77\x5e\x45\x28\x2b\x0c\xeb\x30\xfa\x3f\x87\xd3\xec\xee\x6d\xc5\x19\x75\x6c\x43\xf1\xc9\xff\x8f\x95\x00\x74\xd8\x49\x75\x2c\x35\x3b\x61\xf0\xff\xe5\x0d\x4f\xec\x84\x29\xc6\x81\x49\x6b\x82\xff\x39\x1c\x92\xd3\x4f\x51\xe0\xfd\x44\x4c\x3b\x93\x5a\xc5\xa3\xfd\xef\xcf\x7f\xce\xe6\x96\x88\x9a\xa3\x8a\xae\xdf\xc2\x69\x76\xa3\x38\x2c\x54\x59\xca\x78\xf3\x82\xc7\xef\xe0\xe6\x40\xb8\x45\xb8\x7f\x80\xe2\x17\x71\x48\x55\xb8\x40\x17\xe0\x35\xcd\x87\x21\x84\xcd\xb0\x01\x08\x45\x9d\x32\xc2\xef\x98\x9d\xfb\xc6\x6c\xcc\x32\x83\x4a\x91\x46\xaa\xce\x57\x52\x11\xfa\x82\xa2\x86\x5c\xdf\xe5\xb1\xb4\x38\x5e\x67\x63\xe9\x5d\xae\x4f\x61\x4e\x12\x4a\x18\xb6\xc2\x34\xfe\x37\x84\xc5\x1d\x78\xbe\x3b\xa7\x0a\x7a\x25\x0f\xac\x46\x05\x39\x79\xd5\x81\x75\xb5\x73\x57\xd7\xd6\x2f\xa8\x62\xde\xcb\x0e\xa0\x93\xb5\xe5\x08\xb9\x9b\xcc\x08\x10\xf6\xec\xe2\x32\xb7\xc6\xf4\xf7\xbb\x9d\x4f\x6f\xa5\x36\xf7\xef\xdf\xbd\x7f\xb7\x9b\xcb\x0e\x18\xda\xdf\x12\x5d\x9c\x58\x9f\x6f\xb6\xdf\x44\x3e\x59\x37\x3b\x2a\x05\x78\xab\x5b\xd1\x3d\x2b\x2b\xa2\xb1\x8c\x93\x13\xdc\xee\xb0\x59\x39\x93\x73\xb2\xae\x2e\xda\x32\x20\x59\x33\x7f\x93\xd2\xda\x98\x23\xd6\x1b\x25\x83\xf8\xcd\x28\x97\xf3\x16\x0b\xdf\x54\x3c\x2e\xf5\xac\x8f\x82\x74\xb2\xae\xce\x24\xfd\x0a\x35\x27\x84\x49\xd0\xb4\x5c\x26\xe2\x79\xe7\x38\xe2\x37\x80\x5f\x7a\xa9\xd1\x7f\x80\xe2\x0c\x10\x51\x5f\xfa\x50\xe9\x16\x39\x07\x4d\x15\xeb\x8d\xce\xa4\x35\xbd\x35\x7e\x80\xe2\x5c\x84\x9a\xc3\x0d\xba\x3c\x35\xcb\x9c\x48\x70\x2d\x29\xba\x57\x59\x3d\x2b\xad\xe2\xeb\x0c\xaf\x8b\xbe\xdf\xed\x6e\x86\xa5\xae\x63\x3a\xbb\xd6\x8f\xf9\x12\x27\xb5\x73\x8d\x14\x04\x2c\xc2\x34\xce\x2d\x1f\xb7\x79\xb1\xcd\x97\x4a\x3f\x1b\x84\x39\x79\x6a\x4b\xe9\xe5\xbf\xc8\x9c\x3a\xb7\xea\xd0\x7f\x01\x00\x00\xff\xff\xd2\x00\xcd\x87\x1b\x09\x00\x00")

func terraformTemplatesMainTfBytes() ([]byte, error) {
	return bindataRead(
		_terraformTemplatesMainTf,
		"terraform/templates/main.tf",
	)
}

func terraformTemplatesMainTf() (*asset, error) {
	bytes, err := terraformTemplatesMainTfBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "terraform/templates/main.tf", size: 2331, mode: os.FileMode(420), modTime: time.Unix(1627306109, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _awsProjectPolicyJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x8e\xc1\x4a\xc4\x30\x10\x86\xef\x7d\x8a\x90\x63\xd9\x74\xb7\xbb\x07\x61\x6e\x2b\x78\x13\x05\x05\x2f\xe2\x21\x84\x59\x89\xb6\x49\x99\x4c\x29\x1a\xf2\xee\x92\xb4\x14\xf1\x24\x4b\xe7\xf4\x33\xf3\xfd\xc3\x17\x2b\x21\x84\x90\x2f\x48\xc1\x7a\x27\x41\xc8\xe3\xa1\x3d\xaa\xf6\xa0\xda\x1b\xb9\x9b\x8f\xcf\xac\x19\x7b\x74\x2c\x41\xbc\x96\x55\x9e\xb8\xa6\x02\xdd\x5d\x2e\x68\x32\x21\xcf\x5d\xe7\xa7\xa5\xbb\x9e\xcf\x86\x97\xff\xe1\x04\xf7\x36\xf0\xed\x68\x3e\x91\xff\x62\x4f\x18\xfc\x48\x06\x33\xa8\xc9\x81\x9e\x02\x84\x13\x00\xc4\xd8\xcc\x8d\x94\xe4\x5a\x49\xbb\x2d\x6c\xea\x6b\x24\xf6\xf5\xa6\x1a\x68\xe8\x7f\x1e\x05\x84\x1a\x08\x07\x1f\x2c\x7b\xfa\xda\xf7\xda\xb1\xed\xd4\x40\xfe\x03\x0d\xab\x18\x9b\x47\x7a\xd7\xce\x7e\xeb\xfc\xbd\x79\xd0\x3d\xa6\x94\xd7\x73\xfa\xe5\x5d\xd2\x5b\x95\xaa\x9f\x00\x00\x00\xff\xff\x8a\x71\xb2\xae\x06\x02\x00\x00")

func awsProjectPolicyJsonBytes() ([]byte, error) {
	return bindataRead(
		_awsProjectPolicyJson,
		"aws/project-policy.json",
	)
}

func awsProjectPolicyJson() (*asset, error) {
	bytes, err := awsProjectPolicyJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "aws/project-policy.json", size: 518, mode: os.FileMode(420), modTime: time.Unix(1626946877, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"terraform/modules/dynamodb.zip": terraformModulesDynamodbZip,
	"terraform/modules/funcs.zip":    terraformModulesFuncsZip,
	"terraform/templates/main.tf":    terraformTemplatesMainTf,
	"aws/project-policy.json":        awsProjectPolicyJson,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"aws": &bintree{nil, map[string]*bintree{
		"project-policy.json": &bintree{awsProjectPolicyJson, map[string]*bintree{}},
	}},
	"terraform": &bintree{nil, map[string]*bintree{
		"modules": &bintree{nil, map[string]*bintree{
			"dynamodb.zip": &bintree{terraformModulesDynamodbZip, map[string]*bintree{}},
			"funcs.zip":    &bintree{terraformModulesFuncsZip, map[string]*bintree{}},
		}},
		"templates": &bintree{nil, map[string]*bintree{
			"main.tf": &bintree{terraformTemplatesMainTf, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
