package fetchsymbols

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CachePath returns the local cache location for a device file, rooted at baseDir and
// preserving the on-device directory layout. Device-controlled paths are normalized so
// they cannot escape baseDir.
func CachePath(baseDir, devicePath string) (string, error) {
	cleaned := path.Clean("/" + strings.ReplaceAll(devicePath, `\`, "/"))
	rel := strings.TrimPrefix(cleaned, "/")
	if rel == "" || rel == "." {
		return "", fmt.Errorf("CachePath: invalid device path %q", devicePath)
	}
	return filepath.Join(baseDir, filepath.FromSlash(rel)), nil
}

// IsCached reports whether dest already holds a fully downloaded file of the expected size.
func IsCached(dest string, size uint64) bool {
	stat, err := os.Stat(dest)
	return err == nil && stat.Mode().IsRegular() && stat.Size() >= 0 && uint64(stat.Size()) == size
}

// DownloadToCache downloads the file at fileIndex (its position in the ListFiles result)
// to dest, creating parent directories as needed. The download goes to a temporary
// '.download' file first and is renamed only after it completed, so an interrupted
// download never leaves a truncated file that IsCached would treat as complete.
// progress, if non-nil, is called with the running byte count while downloading.
func (c *Connection) DownloadToCache(dest string, fileIndex int, info FileInfo, progress func(written uint64)) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("DownloadToCache: failed to create cache dir: %w", err)
	}
	err := writeFileAtomically(dest, func(w io.Writer) error {
		if progress != nil {
			w = &progressWriter{w: w, progress: progress}
		}
		return c.DownloadFile(w, fileIndex, info)
	})
	if err != nil {
		return fmt.Errorf("DownloadToCache: %w", err)
	}
	return nil
}

// writeFileAtomically writes the output of write to a temporary file next to dest and
// renames it to dest only if write succeeded.
func writeFileAtomically(dest string, write func(w io.Writer) error) error {
	tmp := dest + ".download"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("writeFileAtomically: failed to create temp file: %w", err)
	}
	if err := write(f); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("writeFileAtomically: failed to close temp file: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("writeFileAtomically: failed to move temp file in place: %w", err)
	}
	return nil
}

type progressWriter struct {
	w        io.Writer
	written  uint64
	progress func(written uint64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	p.written += uint64(n)
	p.progress(p.written)
	return n, err
}
