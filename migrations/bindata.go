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

	info := bindataFileInfo{name: "1_create_table_event.down.sql", size: 18, mode: os.FileMode(420), modTime: time.Unix(1563365464, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __1_create_table_eventUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\xce\xc1\x8a\xc2\x30\x10\xc6\xf1\x7b\x9f\xe2\x3b\xb6\xb0\xa7\x2e\xdd\xcb\x9e\x62\x09\x58\xac\x55\x42\x14\x7b\x2a\xa1\xc9\xa1\x50\x12\x49\x26\x62\xdf\x5e\x68\x45\xa4\x82\xe7\xdf\xcc\x7f\xa6\x14\x9c\x49\x0e\xc9\x36\x35\x87\xb9\x19\x4b\x01\x69\x02\x00\x31\x0e\x1a\xe5\x96\x89\xf4\xf7\x2f\xc3\x51\x54\x7b\x26\x5a\xec\x78\xfb\x33\x73\x20\x45\x06\x67\x26\x96\x91\x3c\x43\x73\x90\x68\x4e\x75\xbd\xb8\x37\xe4\xa7\xae\x77\xd1\x12\xaa\x46\xae\xd4\xdc\xc9\x78\xab\xc6\x6e\xd0\xaf\x46\x5e\x14\xeb\x48\xef\x6c\x30\x96\xba\xb7\x4f\x9e\x10\x03\x39\x3d\x28\xfb\x6d\xfb\xaa\xa6\xd1\x29\x0d\xc9\x2f\x1f\xf7\xbd\x77\x7e\x86\x24\xfb\x4f\x1e\x01\x00\x00\xff\xff\xec\x7f\xd9\x00\x05\x01\x00\x00")

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

	info := bindataFileInfo{name: "1_create_table_event.up.sql", size: 261, mode: os.FileMode(420), modTime: time.Unix(1563521439, 0)}
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

	info := bindataFileInfo{name: "bindata.go", size: 0, mode: os.FileMode(420), modTime: time.Unix(1563521463, 0)}
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
