package pak

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const Magic = "GOPAK0"

var namespace = make(map[string]*offsetInfo)

func Init() error {

	file, err := os.Open(os.Args[0])
	if err != nil {
		return fmt.Errorf("gopak: Failed to open pak @ '%v'", os.Args[0])
	}

	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("gopak: Failed to stat pak @ '%v'", os.Args[0])
	}

	size := footerSize()
	offset := size

	for {

		if offset > fi.Size() {
			break
		}

		_, err := file.Seek(-offset, os.SEEK_END)
		if err != nil {
			return fmt.Errorf("gopak: Failed to seek: %v", err)
		}

		footer := &pakFooter{}
		if err := binary.Read(file, binary.BigEndian, footer); err != nil {
			return fmt.Errorf("gopak: read error: %v", err)
		}

		if string(footer.Magic[:6]) != Magic {
			break
		}

		offset += footer.Size
		n := strings.TrimRight(string(footer.Namespace[:1024-1]), "\000")
		namespace[n] = &offsetInfo{
			Offset: fi.Size() - offset,
			Size:   footer.Size,
		}

		offset += size

	}

	return nil
}

func Open(n string, p string) (io.ReadCloser, error) {

	if oi, ok := namespace[n]; ok {
		return pakOpen(oi, p)
	}

	return realOpen(filepath.Join(n, p))
}

func realOpen(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func pakOpen(oi *offsetInfo, path string) (io.ReadCloser, error) {

	file, err := os.Open(os.Args[0])
	if err != nil {
		return nil, fmt.Errorf("gopak: Failed to open pak in '%v'", os.Args[0])
	}

	if _, err = file.Seek(oi.Offset, os.SEEK_SET); err != nil {
		return nil, fmt.Errorf("gopak: Failed to seek to pak@%v in '%v'", oi.Offset, os.Args[0])
	}

	r := io.NewSectionReader(file, oi.Offset, oi.Size)
	z, err := zip.NewReader(r, oi.Size)
	if err != nil {
		return nil, fmt.Errorf("gopak: Failed to unzip pak@%v in '%v'", oi.Offset, os.Args[0])
	}

	for _, v := range z.File {
		if v.Name == path {
			return v.Open()
		}
	}

	return nil, fmt.Errorf("File '%v' not found in pak@%v in '%v'", path, oi.Offset, os.Args[0])
}

type PackWriter struct {
	r     string
	n     string
	z     *zip.Writer
	w     io.WriteSeeker
	start int64
}

type offsetInfo struct {
	Size   int64
	Offset int64
}

type pakFooter struct {
	Magic     [6]byte
	Namespace [1024]byte
	Size      int64
}

func NewPackWriter(r string, w io.WriteSeeker) (*PackWriter, error) {

	base := filepath.Base(r)

	start, err := w.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	return &PackWriter{
		r:     r,
		n:     base,
		z:     zip.NewWriter(w),
		w:     w,
		start: start,
	}, nil
}

func (w *PackWriter) AddFile(p string) error {

	zp := strings.TrimPrefix(p, w.r)[1:]

	tf, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("Failed to open target file: %v", err)
	}

	zf, err := w.z.Create(zp)
	if err != nil {
		return err
	}

	if _, err := io.Copy(zf, tf); err != nil {
		return fmt.Errorf("Failed to copy file: %v", err)
	}

	return nil
}

func (w *PackWriter) Close() error {

	if err := w.z.Close(); err != nil {
		return err
	}

	end, err := w.w.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}

	footer := newFooter(w.n, end-w.start)
	if err := binary.Write(w.w, binary.BigEndian, footer); err != nil {
		return err
	}

	return nil
}

func newFooter(n string, s int64) pakFooter {
	pf := pakFooter{Size: s}
	copy(pf.Magic[:], Magic)
	copy(pf.Namespace[:], n)
	return pf
}

// Reflect and unsafe were a no go for getting the size reliably.
func footerSize() int64 {
	var b bytes.Buffer
	bio := bufio.NewWriter(&b)
	binary.Write(bio, binary.BigEndian, newFooter("", 0))
	bio.Flush()
	return int64(b.Len())
}
