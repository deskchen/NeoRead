package neoread

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type libraryCopyTarget struct {
	root         *os.Root
	trackEntries bool
	entries      []libraryCopyEntry
}

type libraryCopyEntry struct {
	name string
	info os.FileInfo
}

func verifyLibraryTarget(path string, expected os.FileInfo) error {
	current, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !current.IsDir() || !os.SameFile(expected, current) {
		return errors.New("The destination directory changed during migration")
	}
	return nil
}

func (t *libraryCopyTarget) checkEmpty() error {
	directory, err := t.root.Open(".")
	if err != nil {
		return err
	}
	defer directory.Close()
	_, err = directory.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return err
	}
	return errors.New("The selected folder is no longer empty. Choose the destination again")
}

func (t *libraryCopyTarget) createDirectory(name string) error {
	if err := t.root.Mkdir(name, 0700); err != nil || !t.trackEntries {
		return err
	}
	info, err := t.root.Lstat(name)
	if err == nil {
		t.entries = append(t.entries, libraryCopyEntry{name, info})
	}
	return err
}

func (t *libraryCopyTarget) createFile(name string) (*os.File, error) {
	file, err := t.root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil || !t.trackEntries {
		return file, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	t.entries = append(t.entries, libraryCopyEntry{name, info})
	return file, nil
}

// Remove only entries created by this migration. Leave replaced files and
// directories with unrelated additions where they are.
func (t *libraryCopyTarget) rollback() error {
	var result error
	for i := len(t.entries) - 1; i >= 0; i-- {
		entry := t.entries[i]
		current, err := t.root.Lstat(entry.name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err == nil && os.SameFile(current, entry.info) {
			err = t.root.Remove(entry.name)
		} else if err == nil {
			err = fmt.Errorf("Skipped changed migration entry: %s", entry.name)
		}
		result = errors.Join(result, err)
	}
	return result
}
