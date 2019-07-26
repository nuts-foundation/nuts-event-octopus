// Package migrations Code generated by go-bindata. (@generated) DO NOT EDIT.
// sources:
// 1_create_table_event.down.sql
// 1_create_table_event.up.sql
// bindata.go
package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
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
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
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

// ModTime return file modify time
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

var __1_create_table_eventDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x48\x2d\x4b\xcd\x2b\x29\xb6\x06\x04\x00\x00\xff\xff\x27\x3a\x67\xc6\x12\x00\x00\x00")

func _1_create_table_eventDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__1_create_table_eventDownSql,
		"1_create_table_event.down.sql",
	)
}

func _1_create_table_eventDownSql() (*asset, error) {
	bytes, err := _1_create_table_eventDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1_create_table_event.down.sql", size: 18, mode: os.FileMode(420), modTime: time.Unix(1563875080, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __1_create_table_eventUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\xce\x41\xcb\x82\x40\x10\xc6\xf1\xbb\x9f\xe2\x39\x2a\xbc\x27\x5f\xec\xd2\x69\x93\x85\x24\xb3\x58\xb6\xc8\x93\x2c\xee\x1e\x04\x9b\x8d\x75\x37\xf2\xdb\x07\x1a\x11\x06\x9d\x7f\x33\xff\x99\x5c\x70\x26\x39\x24\xdb\x94\x1c\xe6\x6e\xc8\x0f\x88\x23\x00\x08\xa1\xd3\xc8\xb7\x4c\xc4\xff\xab\x04\x47\x51\xec\x99\xa8\xb1\xe3\xf5\xdf\xc4\xa4\xae\x06\x67\x26\xe6\x89\x34\x41\x75\x90\xa8\x4e\x65\x39\xb3\x33\xde\x8d\x4d\x6b\x03\x79\x14\x95\x5c\xa8\x79\x78\xe3\x48\xf5\x4d\xa7\xdf\x8d\x34\xcb\x96\x91\xd6\xd2\x60\xc8\x37\x1f\x8f\xbc\x20\x0c\xde\xea\x4e\xd1\xaf\xed\x9b\x1a\x7b\xab\x34\x24\xbf\x7c\xdd\x77\xce\xba\x09\xa2\x64\x1d\x3d\x03\x00\x00\xff\xff\x2f\xbb\x08\xce\x04\x01\x00\x00")

func _1_create_table_eventUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__1_create_table_eventUpSql,
		"1_create_table_event.up.sql",
	)
}

func _1_create_table_eventUpSql() (*asset, error) {
	bytes, err := _1_create_table_eventUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1_create_table_event.up.sql", size: 260, mode: os.FileMode(420), modTime: time.Unix(1564053768, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _bindataGo = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func bindataGoBytes() ([]byte, error) {
	return bindataRead(
		_bindataGo,
		"bindata.go",
	)
}

func bindataGo() (*asset, error) {
	bytes, err := bindataGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "bindata.go", size: 0, mode: os.FileMode(420), modTime: time.Unix(1564138692, 0)}
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
	"1_create_table_event.down.sql": _1_create_table_eventDownSql,
	"1_create_table_event.up.sql":   _1_create_table_eventUpSql,
	"bindata.go":                    bindataGo,
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
	"1_create_table_event.down.sql": &bintree{_1_create_table_eventDownSql, map[string]*bintree{}},
	"1_create_table_event.up.sql":   &bintree{_1_create_table_eventUpSql, map[string]*bintree{}},
	"bindata.go":                    &bintree{bindataGo, map[string]*bintree{}},
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
