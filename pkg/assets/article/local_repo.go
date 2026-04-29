package article

import (
	"errors"
	"io"
	"io/fs"
	"kzhikcn/pkg/utils"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrUnsafePath = errors.New("unsafe path")
)

type LocalRepository struct {
	mapMu        sync.RWMutex
	contentLocks map[string]*sync.RWMutex
	AssetsDir    string
	BasePath     string
}

func (l *LocalRepository) assetsDirPath(id string) string {
	return filepath.Join(l.BasePath, id, l.AssetsDir)
}

func (l *LocalRepository) contentPath(id string) string {
	return filepath.Join(l.BasePath, id, "index.md")
}

func (l *LocalRepository) assetsSafePath(id, name string) (string, bool) {
	return utils.SafePath(l.assetsDirPath(id), name)
}

func (l *LocalRepository) getContentLock(id string) *sync.RWMutex {
	if l.contentLocks == nil {
		l.contentLocks = make(map[string]*sync.RWMutex)
	}

	l.mapMu.RLock()
	lock, ok := l.contentLocks[id]
	l.mapMu.RUnlock()
	if ok {
		return lock
	}

	l.mapMu.Lock()
	defer l.mapMu.Unlock()
	if lock, ok := l.contentLocks[id]; ok {
		return lock
	}
	lock = &sync.RWMutex{}
	l.contentLocks[id] = lock
	return lock
}

func (l *LocalRepository) OpenAsset(id, name string) (io.ReadWriteCloser, error) {
	p, ok := l.assetsSafePath(id, name)
	if !ok {
		return nil, ErrUnsafePath
	}

	if err := os.MkdirAll(filepath.Dir(p), 0750); err != nil {
		return nil, err
	}

	return os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0640)
}

func (l *LocalRepository) RemoveAsset(id, name string) error {
	p, ok := l.assetsSafePath(id, name)
	if !ok {
		return ErrUnsafePath
	}

	err := os.Remove(p)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func (l *LocalRepository) ListAssets(id string) (list []string, err error) {
	list = []string{}
	base := l.assetsDirPath(id)

	err = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return ErrAssetsDirNotFound
		}
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		list = append(list, rel)
		return nil
	})

	return
}

func (l *LocalRepository) HasAsset(id, name string) (bool, error) {
	p, ok := l.assetsSafePath(id, name)
	if !ok {
		return false, ErrUnsafePath
	}

	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

type lockedReader struct {
	f  *os.File
	mu *sync.RWMutex
}

func (r *lockedReader) Read(p []byte) (int, error) {
	return r.f.Read(p)
}

func (r *lockedReader) Close() error {
	err := r.f.Close()
	r.mu.RUnlock()
	return err
}

func (l *LocalRepository) ContentReader(id string) (io.ReadCloser, error) {
	lock := l.getContentLock(id)
	lock.RLock()

	f, err := os.OpenFile(l.contentPath(id), os.O_RDONLY, 0)
	if os.IsNotExist(err) {
		lock.RUnlock()
		return nil, ErrContentNotFound
	}

	if err != nil {
		lock.RUnlock()
		return nil, err
	}

	return &lockedReader{f: f, mu: lock}, nil
}

type lockedWriter struct {
	f  *os.File
	mu *sync.RWMutex
}

func (w *lockedWriter) Write(p []byte) (int, error) {
	return w.f.Write(p)
}

func (w *lockedWriter) Close() error {
	err := w.f.Close()
	w.mu.Unlock()
	return err
}

func (l *LocalRepository) ContentWriter(id string) (io.WriteCloser, error) {
	lock := l.getContentLock(id)
	lock.Lock()

	if err := os.MkdirAll(filepath.Dir(l.contentPath(id)), 0750); err != nil {
		lock.Unlock()
		return nil, err
	}

	f, err := os.OpenFile(l.contentPath(id), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		lock.Unlock()
		return nil, err
	}

	return &lockedWriter{f: f, mu: lock}, nil
}

func (l *LocalRepository) Remove(id string) error {
	return os.RemoveAll(filepath.Join(l.BasePath, id))
}
