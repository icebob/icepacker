package icepacker

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FAT is a structure for File-Allocation-Table in package
type FAT struct {
	Count int64     `json:"count"`
	Size  int64     `json:"size"`
	Items []FATItem `json:"items"`
}

// FATItem is a structure for file item in FAT
type FATItem struct {
	Path     string   `json:"path"`
	Offset   int64    `json:"offset"`
	Size     int64    `json:"size"`
	OrigSize int64    `json:"origSize"`
	Hash     [64]byte `json:"-"`
	MTime    int64    `json:"mTime"`
	Mode     uint32   `json:"mode"`
	Perm     uint32   `json:"perm"`
}

// String Convert the whole FAT to string
func (fat FAT) String() string {
	res := []string{}
	res = append(res, fmt.Sprintf("Count: %d, Size: %d", fat.Count, fat.Size))
	for _, item := range fat.Items {
		res = append(res, "  "+item.String())
	}
	return strings.Join(res, "\n")
}

// JSON convert the FAT to JSON string
func (fat FAT) JSON() ([]byte, error) {
	res, err := json.Marshal(fat)
	return res, err
}

// FATFromJSON build FAT struct from JSON string
func FATFromJSON(buf []byte) (*FAT, error) {
	fat := new(FAT)
	err := json.Unmarshal(buf, &fat)
	return fat, err
}

// String convert the FAT item to string
func (item FATItem) String() string {
	return fmt.Sprintf("path: %s, offset: %d, size: %d, mode: %xd perm: %d", string(item.Path), item.Offset, item.Size, item.Mode, item.Perm)
}
