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

var _terraformModulesDynamodbZip = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x0a\xf0\x66\x66\x11\x61\x60\x60\xe0\x60\xd8\x99\xf4\x35\xe8\x9b\x3e\xbb\xf8\x19\x06\x06\x86\x47\x8c\x0c\x0c\xec\x0c\x32\x0c\xb9\x89\x99\x79\x7a\x25\x69\xa1\x21\x9c\x0c\xcc\x76\x5f\xbf\x27\x14\x7d\xff\x9e\x50\x5a\xc1\xcd\xc0\xc8\xf2\x95\x91\x81\x01\xac\xb1\x75\x82\x2f\x5f\x93\x81\x80\xdb\x77\xdb\xa3\x4b\xd8\x3f\xb0\xe9\xc5\x6c\x70\xbf\xd0\xba\x3a\xad\x81\x4d\x4a\x4b\x35\xf7\xa2\xc0\xe1\xbb\xb9\xdc\xa1\x1f\xd5\xcb\x2e\x1f\xfd\x9d\xf6\xec\xac\x47\xe8\x22\x93\x02\x65\x06\xd9\x79\x96\x6b\xee\x8b\xe9\xad\x30\x97\xb3\xb8\xb7\x78\xdb\x21\xae\xab\x1b\xa7\xf5\x4f\xbc\xbc\x6f\xe9\xfb\x9e\x9b\x1b\x9f\xcf\x7b\x2a\x27\xb9\xf4\xed\x66\xa5\x69\x33\xbf\x4d\x3d\x2b\x3a\x71\xe7\xa6\x4e\x81\xb3\xbd\x52\x25\x27\x1b\x1f\x5d\xdd\x90\x55\x22\xb6\x4a\x6b\xc6\x51\xa7\x22\xae\x92\x9f\xd7\x1f\x45\x4f\xa8\x9b\x9f\x71\x67\x41\xe6\xb3\x8b\x82\x6b\x6c\x62\x9d\x92\xc3\x3f\x0b\x74\x76\xae\x70\x91\xd1\x6d\x99\xb2\x68\xf5\x4c\x69\x99\x6f\x9e\xe5\xea\xd3\xd4\x24\x4f\x4f\xfa\x6b\x78\x92\xff\x4a\xfa\x85\x28\xdd\xae\x83\x9f\x16\x6c\x3a\xf1\x47\x8a\x7d\xc3\x9c\x67\xdc\x7c\xe6\x05\x33\x5f\xe6\x5f\x6f\x7f\xe4\x7e\xf8\xaa\xad\xf3\xfa\xce\x3b\x8c\xa0\xc0\xe0\x62\x00\x81\x82\x84\xaf\x41\x32\x62\xb3\x57\xe9\x32\x30\x30\x80\x30\x17\x83\x0c\x43\x7e\x69\x49\x41\x69\x49\x31\x2c\x3c\x3e\x7f\xc0\x1a\x1e\x10\x55\x0a\x4a\x25\x89\x49\x39\xa9\xf1\x79\x89\xb9\xa9\x4a\x0a\xd5\x5c\x0a\x0a\x65\x89\x39\xa5\xa9\x0a\xb6\x0a\x39\xf9\xc9\x89\x39\x7a\x20\x71\xae\x5a\x2e\x44\xe8\x1f\x4d\xfa\x1a\xa4\xcc\x20\xb7\xc9\x8d\x81\x81\xc1\x9f\x81\x81\x81\x87\x41\x86\xa1\x2c\xb1\x28\x13\x64\x0a\xdc\xca\x40\xec\x51\xa0\xed\xad\x73\xca\xc7\xe7\x94\x67\x68\xc0\x29\xef\x33\xbe\xa1\x41\x2b\x9e\x05\x05\x68\xac\xd4\xf1\x08\x6d\x60\x8a\x6e\x7d\x34\x69\xd1\x99\x3d\x25\x0b\x0a\xbc\xb2\x0a\x8b\x02\xd3\x1c\x3d\xd3\xea\xe4\x02\x22\x4e\xf2\xaf\x88\x53\xb4\xf7\x5d\xa8\xa2\x1c\xba\x90\x25\xa5\x45\x28\x5c\x57\x9c\x21\xc0\x9b\x91\x49\x8e\x19\x57\x5a\x90\x00\x07\x0b\x23\x03\x03\xc3\x92\x46\x10\x0b\x9e\x32\x58\x21\x29\x03\xcd\x49\x10\xc3\x70\x85\x25\xb2\x61\xbc\x8c\x0c\x28\x21\xcb\x0a\x09\x59\xac\xe6\xe1\x0a\x2a\x64\xf3\xea\x18\x19\xd0\x02\x8e\x15\x12\x70\x18\x26\xb2\xb2\x81\x74\x31\x33\x30\x33\xbc\x07\xb9\x8a\x09\xc4\x03\x04\x00\x00\xff\xff\x8d\x43\x60\x95\x0f\x03\x00\x00")

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

	info := bindataFileInfo{name: "terraform/modules/dynamodb.zip", size: 783, mode: os.FileMode(420), modTime: time.Unix(1627574669, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _terraformModulesFuncsZip = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x98\x77\x54\x13\xdd\xba\xc6\x07\x48\x42\x33\x91\x44\x8a\x54\x29\x41\x40\xe9\x55\xa4\xe4\x93\x5e\xa4\x4a\x2f\x26\xa1\x77\x90\x12\x90\xa2\x20\x42\x42\x6f\x82\x82\xd2\x41\x41\x8a\x22\x22\xbd\x57\xa5\x09\x06\xa5\x8a\x01\xa5\x77\xa5\x23\xdc\xe5\x75\x7d\x47\x8f\xdf\xbd\xe7\xcc\x5a\x7b\xad\x99\x3f\xf6\xb3\xf6\xde\x33\xbf\x77\x9e\xe7\xd5\xd7\xa6\x00\xd1\x03\x00\x40\x05\x94\x9b\xad\x1a\x0a\xbd\x5a\x08\xfc\x40\x0e\x00\x5f\x4f\x01\x00\x04\x60\x03\xb0\x5e\xce\xc2\xbe\x0e\xc6\x46\xd4\x00\x85\xa9\xe9\x2a\xe6\xc7\xf0\x0b\xa0\x05\xc8\x40\xdb\x64\x00\xf0\xbf\xf3\x56\x4d\xac\x5d\x27\x44\xe9\x43\x06\x9f\xd5\x04\x88\x08\xb9\x3e\x8a\x6f\x82\xbe\xe1\xb7\x34\xa7\xb4\x64\xcf\x88\x75\xbf\xcb\x49\x7f\x28\xbd\xe9\x62\x99\x85\x74\xfc\xaa\x75\x7c\x58\x2e\xf8\x24\x69\x96\x08\x2d\x7c\x2b\xc6\xa8\xa1\xe4\x8b\xf2\x5d\x93\x1d\x8e\xe2\xbd\x60\xd0\x43\xd7\x3a\x8b\x62\x98\xbd\xd9\x37\x67\x35\xab\x77\x5d\x3c\x4a\xeb\x21\x38\x41\x69\xee\xeb\x2b\xca\xc4\x4d\x7a\x12\x3d\x02\xbe\xd5\x21\x83\x0a\x2c\xb9\xbe\x0e\x2e\x90\x1f\x9c\x75\x83\x51\xeb\x39\xc2\xf1\xcb\xdc\xb4\x7e\x82\xe2\x6f\xc9\x8b\xe8\x97\xe2\x15\xc6\xaa\x4c\x19\x72\x1b\xbd\x05\x6e\xa5\xe2\xaa\x5d\xb9\x98\xb7\xbc\xc8\xbe\x59\x78\xd6\xad\x95\xf1\x37\xa9\xf5\xcc\x10\x66\x76\x49\xbb\xfd\x72\xf1\xfc\xa8\x79\x84\x84\xfb\x0b\x52\x6e\xb0\xc0\xdd\x26\xe5\xce\x25\x97\x46\x0c\x51\xc9\x74\x04\x72\xbf\xf9\xb3\x66\x58\x13\x75\x0b\x06\x68\xaf\xdc\x04\x6f\x26\x45\x17\x69\x38\x75\xe9\x6b\xa8\xfa\x3c\xa0\x72\x08\x56\xd3\x76\xe8\x3c\x05\x0e\x0c\x7b\xee\xd1\x24\xc5\x37\x89\x2f\xd3\xda\xb9\x88\x2b\x2a\x38\x8f\x90\xe6\xcd\x3a\x30\xe8\x80\x0f\xf9\x48\xe4\xbc\xb9\x65\x3d\x56\x6d\xb9\xdb\xff\x9e\x62\x69\xc9\x60\x15\x51\x28\xbe\x74\xa9\xaa\xfb\x38\xea\xb6\x8d\x8d\xd4\xc3\xb7\xd4\xf6\x31\x26\x67\xfc\x8a\x5c\x92\x95\x99\x4b\x60\xa4\x6b\x44\x5a\xf8\x94\xc8\xe1\x75\xbb\x9a\x79\x0c\xaa\x47\xd8\xca\x89\x0c\x62\xe6\x94\xf9\x61\x54\xf0\x4a\x3b\x46\x72\x27\xf6\x5e\xf6\x40\x37\x53\xab\x9c\x84\xdf\x15\x03\xe6\x93\xb9\x11\x4d\x4f\xf6\x41\xcd\x91\x2d\x48\x6f\xb1\xa7\x38\x09\x1a\xe0\x19\xbe\x7b\xff\x5c\x17\xd9\xa1\xfc\xe5\x66\x9b\x76\xce\xbd\x4b\x5b\x9f\xa4\x02\xee\xb7\x7e\xa9\x9d\x97\x54\x71\xd6\x0c\x7e\x4e\xda\x79\xf0\x10\x62\xdc\xf5\x39\x40\xab\xfd\x01\x7b\x83\x2e\xf5\xcc\xd5\x59\x2b\x13\xdb\x44\xd8\x0e\x8d\x27\x35\x56\xa0\x8a\xeb\x95\xda\x67\xab\x4e\xf5\xc9\xce\xb9\xb0\xc5\xb4\x51\xe6\xc6\x42\xdd\xfa\xf7\x10\x8b\x54\xdd\xa4\xd6\xe5\xab\xfd\x19\xdf\x4b\xdf\xdf\xcb\xeb\x9d\xac\xc5\xc6\xd5\xe3\x76\xec\x3b\x22\x4d\x5b\x3f\x78\x9e\x7f\x96\xb8\x59\xfe\xb2\xcc\x36\x55\xa8\x13\x2a\x2e\x2c\x61\x9b\x7e\x28\xcb\xb5\xff\x46\x82\x19\x4b\xb9\xaf\xc3\xfa\xcd\xa3\x7f\xfd\xfe\x45\xcf\xde\xbd\x03\xa9\x3d\x3e\x1c\xca\xe5\x40\x2f\x63\x5a\xb7\x70\x37\x77\xa6\x79\x5b\xa1\x6a\x7b\xdb\x9f\xb0\x19\x7b\xac\x8e\xe1\x70\xbc\xdd\xdf\xd6\xfc\xe6\xfb\x42\x94\x58\x9b\x3c\xdc\x5e\x4b\x98\xc8\x35\x78\xfe\x75\x08\xa4\xd0\xb8\x8f\xbb\xb3\x8f\x57\x31\x42\x13\xff\xb8\xc9\x78\xa5\xa8\xd3\x40\xee\x89\x61\xc8\x01\x15\x7b\x55\xec\xbc\xcf\x16\xbd\x96\xf4\xa7\xab\xe1\x50\x37\x0b\x23\xda\xe3\xa5\x4f\xf5\x08\x66\x87\x5c\xc2\x06\x66\xb8\xf2\x9b\x2f\x2e\x2b\xaa\x48\xfe\xe9\x99\x09\x8f\xd3\xa2\x0a\xf2\x05\xeb\x0c\x8d\x29\x7b\x2c\xed\xe2\xc8\xa0\xaa\x7a\x3c\xa2\x98\xeb\x21\x57\xa6\x35\x92\x3c\x35\x2d\x0f\x79\xd6\x82\xd1\xec\x3e\x7f\xff\x7c\xcf\x12\xdf\x3e\xef\x7d\x07\xf5\x57\xe6\x2b\xc7\x41\x19\x37\x22\x4a\x0f\x68\x8b\xb2\xf7\xad\xed\xad\x4e\x26\xf1\xf5\x1a\x05\xf2\x03\x56\xfc\xb8\x42\x6a\x6f\xf5\x12\x68\xbb\xb1\x04\xa5\x7c\x0b\x7b\x1b\x4d\x47\x60\x53\xbd\x31\xcb\xfa\xe9\x5f\x58\xbc\xbc\x31\x6b\x48\xde\xd8\x94\xbb\x0f\x00\xc0\x1a\x19\x00\x50\xfe\xc0\xc2\xcf\xd7\xe9\x6f\x2e\x7c\x8d\x3e\x61\x7e\x8c\x3f\xb8\x88\x48\x6c\xd5\xa2\x10\xa3\x8b\x5c\xbf\x19\x7d\xa3\x11\x1f\xc8\x10\x4d\x20\x50\xcd\x44\x8f\xa4\x68\x9c\xcd\xa8\x60\x75\x4a\xf1\x13\x9f\x78\x11\xbd\xbe\x66\x57\x3f\x5c\x48\x6f\xd5\x63\x3f\xdf\x7b\xf2\x3d\xdd\xc7\x8c\xef\xdb\xf3\xab\x25\x67\x91\x64\xb8\x61\x24\x0c\xdd\x41\xa7\x88\x89\x71\xeb\x3d\x47\xc5\x2c\x4e\xcd\xcf\x09\xd5\xc9\x25\xb0\x8a\x73\x80\x73\xdb\xac\x71\xb6\xb1\xcb\xe2\x15\xae\xd7\x2c\xe6\xb5\x28\xa3\x4d\x37\xdc\xff\x12\xb2\xa6\x7f\x70\xce\x7b\x70\x4e\xdf\xd8\x0b\xcc\xe4\x45\x2d\x01\xd6\x62\xbb\x0a\x7c\xb5\x3d\x75\x4d\x97\x5d\xc4\xd9\xb3\x45\xba\x86\x81\x66\xf5\xb9\x3f\x9e\x83\x4f\x10\xbf\x8e\x9e\xa6\x9f\x9a\x5f\xd0\xa8\x12\x1f\xbd\x73\x4e\xbd\xfb\x00\xd5\x96\x76\xa6\xe1\x23\x75\x56\x1e\x68\x71\x6a\xd4\x6c\x56\xba\x80\xbd\xd1\xfe\x60\x85\xbe\x1a\x9a\xcb\xf7\xea\x83\xad\xcf\xbd\x44\x9b\x0a\xcd\x69\x0e\xa6\xda\xe2\x75\x8e\x75\xb4\xcb\x6e\x5d\x4d\x08\x4e\xcf\xef\xfb\x73\x7e\x4c\xaf\x73\x09\x0b\xae\x47\xec\x34\x66\xac\x49\xa1\xf1\x88\x34\xe4\x59\xe9\x6f\x89\xc9\xef\x3d\x02\xff\x3a\xb9\xf0\xc4\xef\x86\x78\xd1\xe9\x70\x07\x0a\x00\x08\x85\xfd\x2c\x28\xce\x58\xf7\xbf\x0f\xce\x6f\x88\x1c\x8b\x1b\x22\xc7\xfe\x71\x70\x23\xa6\xd6\x9e\xe3\xa2\xf4\x21\xf3\xa6\x66\x58\xbd\xa2\xb0\xe1\x10\xa3\x98\x3a\x4a\xb5\xfa\xe2\xe1\x71\x63\x5f\x8b\x73\x4f\x69\x6a\x07\x0a\xcd\x74\x6c\x1c\xe4\x9e\xc3\x36\xa7\xfd\xdd\x29\x35\x34\xc8\x7d\x97\x38\x32\xb4\x78\x37\x15\x76\xe7\x3e\xfa\xc0\x6e\x24\xe6\xa8\x5c\x60\xa6\x4b\x18\x20\x26\xbe\x0b\xe0\x57\x05\x71\x6c\x72\xb3\x4c\xe3\x89\x3c\x5f\x21\x5c\xd4\x71\x61\xf4\x09\x7a\x1d\x41\x12\x09\xa0\x7c\x83\x09\x15\xa6\x5b\x91\xea\x2f\x1f\x9f\x7e\x7a\x37\x2b\x45\xfb\x5a\x81\xf0\x6a\xe8\xa3\xe8\x07\x5b\x20\xb2\x6b\x17\x26\x72\xeb\xcd\x30\x52\x4f\x8d\x58\xc2\x57\x5d\x2e\xc1\xe1\x8e\x7a\x04\x59\xd9\xcd\xc1\xa6\xdb\xad\x45\x8b\xf1\x74\xeb\x7c\x1a\x9d\xaa\x33\xe0\x0e\xb9\xcb\x2b\x1c\xfb\x53\xca\xcd\x41\x2b\xec\x35\x9f\xaa\x1e\xb3\xd8\x68\x48\x46\x7a\xb2\x96\xe7\xde\xb4\x1d\xcb\xf8\x60\x7b\x59\xdb\xd8\x02\xac\x10\x54\x1f\x46\xfa\x84\x89\xab\x4d\x96\xc7\x06\x2b\xf1\xdc\x2d\xbb\xa6\xe3\x9e\x78\x6d\xa0\xc5\xae\x5e\x73\x5c\xf2\xf1\x6d\x95\x3b\x76\xaf\xd2\x09\x8d\x0c\xb8\xa4\x12\x9c\x34\x0b\x99\xc4\x54\x40\x23\x3f\xa2\x92\xd7\x09\x8b\x5a\x67\x59\xeb\xeb\xab\x76\x92\xb4\xd3\x3c\x97\xc5\xaf\x73\x26\xd2\x4d\x43\x6c\x66\xc4\xc2\xdd\xa0\x9b\x2e\x6a\xb3\x61\x7c\xb5\xd4\x8f\xa2\xac\xc9\x31\xfe\xbb\x8b\xeb\x4a\x1d\x57\x71\x2d\xef\x69\xb8\x01\xdf\x32\x0c\xeb\x5a\x20\xf4\xc6\x52\x79\xbf\xf5\x3c\x26\xb1\xdf\x8d\xb3\xcc\x3b\xc0\x3d\x39\xed\x7e\x8c\xd0\x8b\xaf\xd0\xa1\x0c\x3b\x14\x0d\x5e\xd4\x30\xb2\xeb\xfc\x98\x6a\x81\x67\x22\x63\x75\x8c\x6e\xf8\x18\xf2\xa5\xb3\x80\xb3\x30\xbe\x98\xc7\x0f\xcc\x3f\x3c\x6e\x35\xf8\xb0\xe4\x45\xc4\x38\xc2\x2e\x65\x12\x81\x51\x5d\x77\x17\x70\x98\xa5\x33\x28\x94\x95\xc0\x99\xea\xd1\x04\xd1\x1e\xc0\xbb\x8f\x21\x4a\x0b\x1c\xb6\x39\x46\xaa\x23\x65\xe9\x4c\xdc\x9d\x9b\x59\xe9\xbe\x3c\xb5\x90\x67\xf7\x7c\x4c\xc5\x0c\xd0\x48\xae\xc1\xc8\x0b\x29\x0e\x5c\xd9\x11\x04\x97\x7c\xc6\x30\x04\xe5\x87\x16\x43\x6a\x14\x52\x87\xa8\xa2\xd1\x5b\x99\xe6\xac\xca\x34\x67\x76\xcf\x24\xe6\xe2\x72\x89\x5c\xd0\x5c\x79\xa6\x6a\xd6\xac\xe5\x36\x4e\x6f\x82\x7b\x63\xa4\x0e\x77\x5f\x60\x64\x22\xa7\x2a\x9b\x39\xfa\xf8\xb6\xb9\x6a\x1b\x1f\xd2\xea\xee\x6e\x10\xcd\x8b\x3b\xcd\x04\x41\x39\x95\xe2\x24\xa4\x48\xb7\x09\xd7\x12\xd2\x0c\x6a\x4d\x9f\xf3\x57\x08\xfb\xa2\x76\x9b\xeb\x82\x07\x42\x46\xa1\xf5\x43\xc2\xe2\xdd\x24\x10\xa7\xd1\x1e\x7a\xf0\x6e\x70\x87\x2a\xd4\x5b\x85\xd9\xec\x33\xa0\xc6\x49\x56\x3d\x13\x1a\x5b\x95\x55\x94\xf3\xae\xb5\x2c\x5d\x21\x5a\xe0\x8e\x17\xe3\x57\xc8\xdb\xfd\xd4\x3b\xe2\xca\x45\x03\xf5\x1e\xbb\x7b\x7b\xb3\xf8\x5d\x76\xa4\x5c\x8a\x72\x42\x6b\xb7\x7b\xf0\x1b\x65\xaa\x90\xbd\x38\xb4\x1f\xc5\x76\x1c\xcd\x2d\xfe\xbf\xf6\xa3\xde\x1e\x3f\xd0\x85\x83\xc5\xe7\xde\x78\x95\x05\x0b\x05\x13\xa5\xb6\x94\xf5\xaf\x6e\x1d\x7e\x6f\x7e\x75\xa9\x2b\x4b\x5a\xac\xc7\x74\x95\x0e\xa3\xde\x5a\x3c\x4b\x29\x9c\x19\x38\xaf\xbd\x29\x2d\x43\x4b\xbb\xa6\xdd\xa2\x2b\x9d\x6b\xdc\x73\x8d\x80\x41\x17\xb8\x7e\x92\x0f\xb2\x2f\x69\xb8\xa1\x97\xef\xd8\x55\xe8\x2d\xb2\xb2\xfc\x42\xd9\x64\x0e\x93\xa4\xd1\xce\x3c\xe2\xd6\x80\x85\x05\x11\x65\x91\xf7\xbf\x5c\x0d\x75\x18\xf8\xe2\x3b\xf9\xb4\x72\xa1\xc3\xc9\xce\x64\x30\x88\x49\xbb\x3a\xf6\x2e\x92\xe8\x31\xa5\x42\x3b\x2c\x33\xe2\xe1\x3d\xc2\xc6\x5f\xb2\x73\x3b\xd8\xd8\x31\x53\x43\xbe\xa8\x0f\x36\xdd\xd1\x6e\x2a\x30\xfc\xfa\xb1\x87\x6a\x4f\x55\x57\x65\x0d\xc2\x27\xd5\x05\xf4\x48\xa7\x02\x5e\x95\x5d\x60\x74\x3e\x6c\x6c\xd7\x3d\xb0\xf7\x88\x01\x8a\x3e\xd8\xdc\x18\x84\xf4\x28\x70\x47\x12\x33\xde\x6e\x55\x05\x5a\x3c\x80\xa5\xc5\xf7\x70\x50\xe8\xc6\x8f\x4f\xc3\x5c\xbf\x44\x5e\x52\x83\x2d\x32\x4e\xb8\x0e\xbe\x14\xf3\x90\x3b\xa4\xfd\x45\x29\xd7\x8d\x1d\x43\x10\xab\xb8\x64\x0c\x18\x00\x78\x60\x3f\xeb\x9b\x3b\xd6\xd9\xe3\x6f\x4c\xa7\xd0\xfb\x98\x1a\x9b\xfd\x3f\xeb\x5b\xaa\xa9\xbb\xe7\x84\x34\xdd\x71\x39\xfa\x46\x0a\x6d\x87\x30\x9b\xac\xfb\x73\x1c\x6d\x22\x74\x2c\x83\x91\xec\xb2\xaf\xd3\x61\xf1\x32\xa7\x3e\x42\xc8\xe7\xb1\x46\x78\xc1\xa3\x9a\xb0\x93\xc3\x35\x0d\xad\xb4\x8c\x02\x1d\x28\x92\xf0\xba\x7e\xfe\x63\xb6\x6d\x81\x70\xfa\x4b\xaf\x87\x00\xff\x7b\xaf\x4b\xf8\x6f\xaf\xcb\x48\xac\xaf\xb5\x14\xa7\xc4\x53\x8c\xeb\xb0\x4a\xfd\x5c\xfd\x86\x47\x7e\x8f\x06\x76\xba\xf5\x3c\x16\xef\xb8\xab\x5c\x17\xed\xfd\x0a\xf5\x29\x09\xc9\xe0\x6c\x03\x09\xfa\xeb\x6f\xf2\xba\x16\x56\xb6\x18\x81\x59\x3b\x6c\x22\xdb\x5f\x18\x17\x9e\xb1\x20\xf3\x19\x20\xe4\xf9\x4d\xe6\x29\xf9\xbb\x0c\xad\x09\x09\x56\x85\x6d\xb3\xa5\x71\xd2\x27\x46\xb4\xaf\x32\x97\x92\x8e\x64\x8c\x80\x75\x81\x90\x15\xe7\xa6\x39\xb9\x86\x1b\x0e\x32\xd6\x0e\xa4\x8b\x81\x0f\x06\xf2\x42\x9d\xcc\xfa\xd5\xc5\xbf\xbf\xd6\x2f\x95\x48\x3e\xdc\x18\xe4\x0f\xd6\x94\x7d\xbd\x20\xd7\x59\xf7\xf1\xc9\xc6\x5b\xc7\x8a\x1c\xe0\xfe\x8c\x03\x5d\x54\x77\x2d\xa3\x07\x98\xcb\x6f\x28\xde\x39\xf5\xcc\xc7\xb6\x16\xf8\x1e\xf5\xd7\xcc\xb6\xf5\x1c\x0b\x81\x20\x54\x3e\x7f\xa0\xf9\xb2\xe3\x98\xce\x65\x1e\x39\xc5\x12\xdd\x9d\x4b\x6d\x4a\x65\x3c\x3b\x34\x12\x42\xb5\x37\x4f\xe1\xe8\x33\xf9\xac\xa9\x22\x3f\x1c\x39\x64\xdc\x3a\x3e\x5a\x12\x60\xdc\x23\xa3\x74\xf0\x8a\xc7\x68\x99\x1e\xa1\xb4\x13\x8b\xf5\x19\x9a\x17\x5b\x0c\xb5\x58\x23\x32\xd4\x88\x63\xe1\xd4\x87\x61\x7a\x4c\x95\xbe\x21\xc1\x44\xce\xcf\x20\xf0\x6d\x49\x6f\x52\xcf\xf4\x08\x4b\x7c\x4f\x42\xfe\x82\x7a\x3b\xab\x3b\xae\x96\x01\x1d\x90\x1d\xda\x3a\xc6\x96\xa5\x70\xf2\xa5\xbe\x09\x76\x44\x55\xf7\x57\xdd\x10\x79\xa1\x66\x75\xb9\x7a\xc2\xf4\xfc\xc6\xcc\xe7\x4d\xcd\xb3\x66\x84\x03\x78\x66\x89\x10\xb7\xcc\x24\x11\xcb\xe8\x5a\x4c\x7b\xde\x47\x7e\xed\xc2\x97\x93\x86\xa9\xc1\x26\xec\xa9\x21\x05\x72\x90\x3b\x2a\xf5\xbb\xe4\x81\xc9\x2e\x8e\x0e\x2d\x6a\xbe\x57\x51\x70\x5a\xc7\xef\x1c\x63\x48\x7a\xfb\x01\x7e\xf9\x32\xa1\x6f\x16\x3b\xf5\x5c\x22\x3f\xe2\xbd\xbe\xe5\xe3\x8b\xf3\xd2\x3b\x6e\x15\x54\xa5\x08\x07\x87\xe6\x9a\x04\xdb\x0d\xaf\x15\x98\xee\x79\x1d\x8c\xa5\x86\x01\x87\x0d\xf7\xa4\x40\x15\xc5\x79\xef\xf0\xaa\x21\xad\xb5\x3a\xe7\xec\x50\x33\xf2\x81\x21\x2d\x56\xc1\x3c\x2e\x3c\xe1\x85\x9a\x6c\xab\x44\x61\xc5\x68\x5f\xe9\x4d\x70\x1f\x0e\x62\xd1\x9b\x4b\xac\x42\xe8\x9e\xe6\x7a\x5c\x48\xea\xcb\x45\xbe\x60\x3c\x78\x58\x6b\xef\x18\x7b\x57\x54\xe9\x1c\xd5\xf7\x9b\x1c\x94\xde\x51\xb1\x41\x4e\x3d\xa1\x8a\xa7\x9f\x54\x0e\xf9\xb1\x8e\xbb\x65\xbf\x1d\x54\xd4\x3c\xbf\xb1\x39\x55\xe2\x3f\x0a\x0f\xec\xcd\x75\xea\x95\xaa\x5b\x2b\x4b\x99\x30\x91\x39\x0a\xec\x84\xaa\x57\xb8\xc2\xa2\xf3\xb5\x17\x16\xa3\xdd\xf7\x43\x33\xb9\xcf\x05\x79\xf2\x28\xb5\x8a\x99\xb3\xed\x41\x45\x2c\xf9\x8c\xa0\x1e\x11\x02\x4e\xbe\x23\x5a\x6d\xc9\x56\x23\x0b\x0d\xd7\xd5\xf6\xe5\xc5\x39\x6c\x98\x1e\x61\x26\xa3\xc3\xb1\x34\x72\x6e\x81\xe4\xdd\x5e\x07\xde\x91\x16\x91\x96\x53\x19\xc9\xca\x6f\xca\xdd\xe3\x57\xd8\xee\x5d\xd3\x08\x42\x1b\x7e\xde\xa2\x1e\xb1\x8b\x5d\x96\x49\x74\xa7\x5d\xee\x73\x22\xcb\xec\x23\x7d\x03\x7b\x64\x13\x2f\x2a\x58\xcc\xcb\x2e\x7e\x93\xce\x97\xbe\xf8\x25\x47\xdb\x20\xc8\xb6\x83\x31\xe6\xac\x07\x7b\xc9\xea\xc6\x0e\xe3\xf5\x9b\xf2\x47\x5f\x81\x78\x06\x2e\x73\x81\xb4\x86\xc9\xd4\xe4\x79\xc8\x3b\x06\x9e\xbc\xd7\xf6\x11\x1d\x82\xd9\xf8\x0b\x91\xca\x85\xc6\x63\x2c\xed\xa9\xbd\xd5\x93\x66\x16\x97\x61\x67\x1c\x23\x9e\xa1\x9f\xd0\xe8\x57\x4c\xf5\x1c\xf7\x1e\x77\x5d\x76\x27\x7c\xd4\x7e\x2d\xb8\xda\xf7\xad\x7f\x22\x06\xf5\x4c\x78\xfa\x89\x01\xf7\x9e\x4a\x73\x85\x65\x17\x6f\x62\xa9\xc0\x85\x34\xff\xf4\xbd\x3a\xba\xbf\x40\x19\x3e\x26\x36\x0f\x21\x26\x97\x4a\xd1\x16\xab\x6a\xda\x81\x6c\x74\xc9\x18\x1d\x5d\x74\x25\x90\x2e\x96\xdf\xb6\xd5\x3d\xf4\x74\xe3\x4b\x8e\x3f\x34\x92\x40\x81\x2e\x1f\x32\x65\xe1\xeb\x4b\x53\x88\xe3\xd5\x3e\x5a\x8a\x2f\xf3\x87\xd9\x89\x4d\x88\x8f\x6d\xd5\xb4\x1d\x56\xa9\x16\x2b\x6a\x64\xed\x4b\xa5\x7d\x0b\x9a\xe6\x9b\xc3\x7e\xa9\x97\x81\x7f\x9a\x88\xdf\x1a\xbe\xeb\xaf\x3c\xdc\x78\xce\x82\x67\x6d\xb4\xae\xde\x66\xbd\x71\xc5\x3a\x1e\xe2\x5d\x3d\xb1\xb1\xce\x55\x68\xea\x7b\x0f\x36\xc9\x8e\xd2\x57\x5b\xca\x2f\xb1\x6b\x58\x6f\xda\xba\x19\x20\x95\x7b\xe7\x68\x5b\xdd\xdb\x0c\x6d\xfd\x04\x4a\x1f\x69\x7a\xbd\x75\xcb\xb0\x84\x26\x2a\x6c\xaa\xe5\x19\x1f\x9f\x68\xc8\x8e\x84\xa4\x12\x7c\xcc\xef\x70\x9f\xe3\x01\x69\x04\x29\x2b\x1f\xbe\x06\x67\xbf\xae\xf3\xa9\x07\xa5\xe0\x94\x82\xc2\xb1\x7d\x6f\xa2\x57\xbb\x88\x1e\xd9\xd0\xfd\x66\x09\xe5\x1b\x8a\xd7\xc9\xc5\x20\x02\xfc\xc8\xc3\x8c\xb0\xf5\x02\x62\x5a\xdd\xae\xb3\x8b\x44\x34\xc9\x3c\xfe\x8a\x88\xea\x2a\x49\xde\xbc\x95\x77\x40\xb8\x89\x6c\x5c\x09\x6b\x39\x9e\xea\x61\xaf\xaa\xa2\x61\x42\x38\x13\x97\xcd\xf5\xf8\xd3\xd6\x36\x25\xd2\x20\x96\x9a\xa2\x65\x7b\xf9\xf2\xd3\x97\x46\x5d\x31\x42\x67\x92\x19\x94\x68\xaa\xc3\xf5\x2c\x36\x4b\xbb\x4f\xde\x39\xb9\x26\x88\x28\x82\xed\x83\x55\x79\x64\xde\x29\x8e\x7b\xbd\xa3\xc2\xc7\x3c\x3e\x03\x15\xfe\x9c\xf7\x22\xfc\xf3\x91\x72\x8c\xde\x9e\xb5\x4a\xc3\xb2\xea\x4c\x74\x6b\x16\x22\xbd\xe8\x78\x95\x98\xb4\x06\xcd\x18\x74\xcd\x9f\x88\xc3\x5f\xf1\x7b\x26\x55\xb7\xd8\x25\xef\xd2\xa5\xda\x51\x40\xfc\xe0\xb5\xb9\xcb\xd0\x70\xe5\x63\x83\x83\x5c\xb4\x75\x51\xd9\x75\x9f\x80\x65\x86\x83\xae\xc6\x6c\x42\xf0\xa5\xf7\x90\x0b\x79\x0a\x68\xca\x47\xf2\xa1\x93\xa4\x82\xa0\x4b\x6e\x7d\xfa\x28\x51\xbe\xf7\xb9\x42\x4b\x88\xd8\xf1\x77\x27\x60\xd5\xf0\xd6\x0e\xb3\xf2\x3b\xb4\xd4\x42\x34\x83\xfd\x78\x55\xb8\xd5\x25\x5a\xc4\x33\x39\xf1\x42\x2e\x19\xa5\x54\x9b\x76\x3b\xbe\x28\xb1\xce\x44\xf1\xf7\xcf\x45\x2c\xb0\x20\xf3\xe2\x87\xc4\xf9\x13\x1e\x9c\x52\x5d\x40\x5d\x40\x46\x9b\x1a\xab\x27\x48\x6e\xf0\x43\xbd\xa0\x6c\xf2\xe1\xf0\x02\x64\xc2\x09\xaf\x1c\xa7\x17\x47\xb3\x80\xe6\x3b\x7a\xff\x65\x38\xfc\xed\x6d\xe6\x8a\x92\xc8\xd4\x80\xdb\xfb\x24\x1e\x76\xfd\x7c\xd0\x22\x9b\xec\x65\x41\x4d\xad\x59\xd1\x09\xc2\xfc\x42\x68\x96\x85\x79\x30\x39\x26\x4f\xff\xd6\x7a\x63\xbb\xcb\xc9\x1f\x5e\x35\xb8\x3f\x0a\xe1\x0b\x00\x40\x3a\x00\x00\x34\x00\x1b\xe0\xe9\xe7\xeb\xe5\xe7\xeb\xf3\x5f\xec\xaa\x6a\xdc\x15\x58\x3b\xe7\x29\xd0\x7a\x43\x1b\xb7\x6c\x0a\xd3\x37\xa1\xb4\xd4\x3a\xf8\x53\x0a\xfe\xb3\x79\x54\x9b\xbb\xc9\xca\x13\xcb\xa5\x38\x87\x3b\x66\x92\x0c\x60\x4b\x58\xc5\xdb\x3e\x10\x90\x5a\x20\x2a\xdf\x32\x10\x37\x10\x7b\x4f\xde\xa3\x20\x78\x38\x42\x4f\x17\xf1\x2c\x33\xbc\x2e\x2c\xf3\xb4\x54\x13\x73\x77\xb5\x88\x4f\x89\xef\x24\x7d\xe5\x70\xd7\x59\xa7\xf9\xa6\x51\xec\x4a\x57\xf0\xc5\x67\x08\x93\x6d\xb1\xcf\xf5\xdb\x22\x9a\x56\x9a\x93\x9e\x92\x53\xc1\xc0\xaf\x75\x83\xbd\x76\x0c\x7b\x51\x57\xb3\xee\x92\x01\x80\x36\x08\x00\x4e\x01\x6c\x00\x0e\xeb\xed\x8c\xb5\x71\xb3\xff\xd7\xca\x5f\x5b\xff\xdf\x3f\xa2\x94\x56\x8f\x70\xd1\x53\x91\xeb\xf2\x11\x9d\xc5\x13\x61\xda\x55\xb6\xa7\xdf\x9d\x5f\x6d\x52\x04\xc4\x5b\xdf\x3a\x71\xa9\x69\x3a\x75\x3f\x6e\x3c\x7c\xf4\xb2\xe4\x6c\x8e\x4e\x3f\x75\xf7\xde\x09\x8a\xd5\x47\xe5\x49\x85\x38\xe8\xa6\x84\x19\x83\xa4\x65\x76\xbb\x23\x44\xff\x5a\x78\x38\x35\xdd\x20\x4c\xb1\x79\x38\xd5\x54\x77\x77\x93\xb7\xf6\x4d\x76\xf1\x90\xc9\x13\x02\xab\xe8\xa8\xb1\x54\x9b\x71\x61\xa8\x02\xdc\xb7\xc8\x1d\x8e\xec\xa5\x89\xc3\x47\xb2\x63\xa8\x74\x4d\x95\x39\xd1\x9f\xc7\x4c\xd3\xd8\x9f\x81\xca\x48\x34\x3a\x9a\x3b\x01\x6a\x54\xb6\x7a\x36\x44\x6c\xf3\x15\xb9\x8f\x3a\x39\x38\x81\xd3\x79\xdd\x49\x05\x4a\xf7\xb5\x1b\xda\xf6\x92\x91\x57\x90\x55\xb6\x13\x49\xd0\x8d\x59\xff\x5c\x31\x87\x36\x79\x8a\xa7\x65\x22\xa8\xd1\x61\xd0\xbc\x37\xee\x53\x21\xed\x84\x30\x3b\xda\xdc\x22\x68\x94\xa6\xaa\x33\xcd\x85\xb0\xed\x75\x39\x91\x87\xea\xd4\x1b\x38\x84\x7c\x45\xea\xe5\x96\x57\x1b\x9c\x4d\x00\x27\xe5\xb6\xc0\x5e\xff\x51\x3f\xb7\xa6\xfe\xa1\x8f\xd7\xa3\x59\x21\xfa\xc1\x62\x72\x41\x9c\xaf\x87\x8f\x4d\x4d\x52\xce\x86\xb2\xf1\x11\xa9\xa8\x27\x02\x31\xa5\x93\xa5\x67\x98\xcf\x5a\x66\x13\x3e\x24\xc1\xe6\x34\x42\x8b\xf4\x3f\x39\x17\x4a\x4c\x55\xec\x01\xf5\x8e\xb6\xa3\xce\xbe\xd3\xe9\x7f\x57\x94\xe6\x9b\x90\x42\x8d\x24\x49\x3e\xda\x67\x4a\x78\xc5\x10\x78\x21\xe8\x14\x7e\xba\xfc\x62\x3c\xb7\x5c\xc9\xee\xc4\x0b\x16\x22\x6b\x04\x2b\x61\x15\xd3\x97\xec\xdf\x1c\x51\x39\x6c\x7f\xb3\x5c\x9b\x86\xdf\x40\xa7\x91\x94\xb4\x04\x3e\x11\xed\x3e\x0e\x32\xbb\x18\x1e\x38\x63\x72\xec\x12\xd9\x9a\x12\x9e\x35\x56\xea\x69\xe0\x6b\xf3\x04\x25\xfa\x9e\x37\xd6\x53\x5f\x9b\x8c\x9c\x83\xe2\xff\x6b\x23\x9c\x05\x7e\x5c\x64\x00\x00\x14\x84\xfd\xb8\xfb\xbb\xa9\x00\xfe\xd9\x54\xf8\xe3\x7d\xfe\xae\xf5\xcf\xec\xf5\xbb\x16\x13\x05\xf0\x2b\x89\x81\x7f\x7e\xda\xff\x41\xec\x9f\x71\xe4\x77\x31\x03\x10\xf0\xaf\x70\x02\xfe\x19\x4e\xfe\x83\xd6\x3f\x4d\xd3\xef\x5a\xbb\x94\xc0\x2f\x0b\x05\xfe\x69\xa1\xfe\xcb\x2e\xff\x9d\xda\xdf\xc5\xda\xa1\xc0\xbf\x31\xfc\xdf\x37\xfa\x4f\x9a\x7e\xd7\x0b\x82\x01\x7f\xb0\x05\xfe\xc9\xd6\x3f\x14\xc1\x90\x1f\xb3\x20\x00\x04\x20\x92\x01\x80\x2a\xdd\x8f\xa7\xff\x09\x00\x00\xff\xff\x40\x0b\xcc\x0e\x2f\x12\x00\x00")

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

	info := bindataFileInfo{name: "terraform/modules/funcs.zip", size: 4655, mode: os.FileMode(420), modTime: time.Unix(1627574669, 0)}
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

	info := bindataFileInfo{name: "terraform/templates/main.tf", size: 2331, mode: os.FileMode(420), modTime: time.Unix(1627463314, 0)}
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

	info := bindataFileInfo{name: "aws/project-policy.json", size: 518, mode: os.FileMode(420), modTime: time.Unix(1626962834, 0)}
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
