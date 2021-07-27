// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// github/mantil-workflow.yml
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

var _githubMantilWorkflowYml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\x92\x4d\x6f\xd3\x40\x10\x86\xef\xf9\x15\x73\xe0\x06\xeb\x08\xb8\xf9\xd4\x28\xf5\xa1\x2a\x2d\x12\x8e\x84\x10\x42\xd6\x76\x3d\x89\x97\xd8\x3b\x66\x67\x36\x4b\x28\xfd\xef\x95\x3f\xe2\x5a\xce\xcd\xde\x79\xde\x47\xaf\x66\xd7\xe9\x06\x53\x78\xd0\x4e\x6c\x0d\x91\xfc\x71\x5f\x53\x5c\x91\x4b\xe1\x67\x1b\xb8\xfa\x30\x9d\x15\xa5\xe5\x56\x8b\xa9\x7e\xad\x7e\xd3\x13\xa7\x2b\x80\x5b\x6c\x6b\x3a\x77\x5f\x00\x3e\x38\x56\x5d\x2c\x3c\x05\x27\x41\xd5\x5a\x90\xa5\x1f\xa1\x3b\x0d\x0c\xc0\xc3\xe6\x71\x77\xf7\xa5\xd8\x7d\xbd\xcf\x1e\x53\x78\xf7\xfc\xcc\x68\x3c\x0a\x27\xf3\xc1\xcb\x4b\x4f\xb3\x60\xcb\x97\xa0\x82\xa1\xe8\xb6\x42\x73\xa4\x20\xe0\xb1\x25\xb6\x42\xfe\x3c\x12\x00\x81\x91\x53\xd0\x46\x2c\x39\x5e\x9b\x91\xbc\x39\x7d\x5a\x38\x72\x94\xd0\x82\xa0\xf7\x7a\x4f\xbe\x59\xe4\x2b\xcd\x95\x35\xe4\xdb\x35\x77\x9c\x9a\xb8\x9b\xd3\xc7\x65\x19\x72\x7b\x7b\x08\x1e\x41\x47\x06\xe3\xb1\x44\x27\x56\xd7\xbc\xac\x14\x59\x4d\xb5\x2e\x19\xd5\x9d\xce\x32\x6f\x7a\x80\x68\xa5\x4a\xa7\x3f\x18\x05\x06\x99\xd5\x11\xcf\xca\x96\xfd\xee\xe0\xb2\xbc\xcd\xf7\xbc\xd8\x6c\xb7\x59\x9e\x17\xf7\xd9\x8f\xe2\xee\x16\xc6\x15\xbe\xa5\x07\x74\x26\xb9\x36\xe4\xd9\xf6\x5b\xb6\x9b\x89\xae\x2d\x1e\x0f\xb6\xbb\x63\x0c\xca\xa0\x13\xaf\x6b\xb5\xdc\xc9\xf0\x26\xa6\x9c\x0f\x2e\x85\xff\x33\x4d\x3c\xa0\x80\xfa\x03\x95\x48\xcb\xe9\x7a\xdd\xf4\x2f\x4f\x95\x14\x5d\x4d\xba\xe4\x84\x3f\x27\x73\x7d\xa2\x1b\xfd\x8f\x9c\x8e\x9c\x18\x6a\x46\x7c\xe6\x33\x55\x43\x25\xbc\xff\x0b\x57\x93\x64\x84\xa1\x1c\x2a\xbd\x06\x00\x00\xff\xff\xb5\x3e\xbf\xd7\xea\x02\x00\x00")

func githubMantilWorkflowYmlBytes() ([]byte, error) {
	return bindataRead(
		_githubMantilWorkflowYml,
		"github/mantil-workflow.yml",
	)
}

func githubMantilWorkflowYml() (*asset, error) {
	bytes, err := githubMantilWorkflowYmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "github/mantil-workflow.yml", size: 746, mode: os.FileMode(420), modTime: time.Unix(1627380551, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _awsProjectPolicyJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x8e\x41\x4b\xf4\x30\x10\x86\xcf\xcd\xaf\x08\x39\x96\x6d\x77\xbb\x7b\xf8\x60\x6e\xfd\x40\xbc\x88\x82\x8a\x17\xf1\x10\xc2\xac\xc6\x6d\x93\x65\x32\x65\xd1\x90\xff\x2e\x69\x4b\x41\x0f\xa2\xb2\x39\x0d\xf3\x3e\xef\xe4\x89\x42\x4a\x29\xd5\x03\x52\xb0\xde\x29\x90\x6a\xbb\x69\xb6\x55\xb3\xa9\x9a\x7f\x6a\x35\x85\x77\xac\x19\x7b\x74\xac\x40\x3e\x8e\xab\xfc\xe2\x32\x8d\xd0\xc5\x7e\x8f\x26\x13\xaa\xed\x3a\x7f\x9a\xbb\x4b\xdc\x1a\x9e\xef\x87\x1d\x5c\xd9\xc0\xff\x07\x73\x40\xfe\x8a\xdd\x62\xf0\x03\x19\xcc\xa0\x26\x07\xfa\x14\x20\xec\x00\x20\xc6\x7a\x6a\xa4\xa4\x96\x4a\x5a\x9d\xc3\xa6\xfc\x8b\xc4\xba\x3c\xab\x06\x1a\xfa\x99\xc7\x08\x42\x09\x84\x47\x1f\x2c\x7b\x7a\x5b\xf7\xda\xb1\xed\xaa\x23\xf9\x57\x34\x5c\xc5\x58\xdf\xd0\xb3\x76\xf6\x5d\xe7\xeb\xf5\xb5\xee\x31\xa5\xbc\x9e\xa6\x4f\xde\x45\x14\xc5\xef\x4d\x2f\x91\xdb\x81\x5f\x3c\xcd\x7f\xdc\xfb\x03\xba\xef\xec\x4b\x25\x8a\x34\xc6\x4f\x22\x89\x8f\x00\x00\x00\xff\xff\xdb\x19\xd8\x94\x74\x02\x00\x00")

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

	info := bindataFileInfo{name: "aws/project-policy.json", size: 628, mode: os.FileMode(420), modTime: time.Unix(1627382097, 0)}
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
	"github/mantil-workflow.yml": githubMantilWorkflowYml,
	"aws/project-policy.json":    awsProjectPolicyJson,
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
	"github": &bintree{nil, map[string]*bintree{
		"mantil-workflow.yml": &bintree{githubMantilWorkflowYml, map[string]*bintree{}},
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
