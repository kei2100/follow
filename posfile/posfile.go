package posfile

import (
	"encoding/gob"
	"os"
	"sync"

	"github.com/kei2100/follow/stat"
)

// PositionFile interface
type PositionFile interface {
	// Close closes this PositionFile
	Close() error
	// FileStat returns FileStat
	FileStat() *stat.FileStat
	// Offset returns offset value
	Offset() int64
	// IncreaseOffset increases offset value
	IncreaseOffset(incr int) error
	// Set set fileStat and offset
	Set(fileStat *stat.FileStat, offset int64) error
	// SetOffset set offset value
	SetOffset(offset int64) error
	// SetFileStat set fileStat
	SetFileStat(fileStat *stat.FileStat) error
}

type entry struct {
	FileStat *stat.FileStat
	Offset   int64
}

// Open opens named PositionFile
func Open(name string) (PositionFile, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0600)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	var ent entry
	if fi.Size() == 0 {
		return &positionFile{f: f, entry: ent}, nil
	}
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&ent); err != nil {
		return nil, err
	}
	return &positionFile{f: f, entry: ent}, nil
}

type positionFile struct {
	f *os.File
	entry
	mu sync.RWMutex
}

func (pf *positionFile) Close() error {
	return pf.f.Close()
}

func (pf *positionFile) FileStat() *stat.FileStat {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.entry.FileStat
}

func (pf *positionFile) Offset() int64 {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.entry.Offset
}

func (pf *positionFile) IncreaseOffset(incr int) error {
	return pf.Set(pf.FileStat(), pf.Offset()+int64(incr))
}

func (pf *positionFile) Set(fileStat *stat.FileStat, offset int64) error {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	pf.entry.FileStat = fileStat
	pf.entry.Offset = offset

	if _, err := pf.f.Seek(0, 0); err != nil {
		return err
	}
	enc := gob.NewEncoder(pf.f)
	if err := enc.Encode(&pf.entry); err != nil {
		return err
	}
	return nil
}

func (pf *positionFile) SetOffset(offset int64) error {
	return pf.Set(pf.FileStat(), offset)
}

func (pf *positionFile) SetFileStat(fileStat *stat.FileStat) error {
	return pf.Set(fileStat, pf.Offset())
}

// InMemory creates a inMemory PositionFile
func InMemory(fileStat *stat.FileStat, offset int64) PositionFile {
	return &inMemory{entry: entry{FileStat: fileStat, Offset: offset}}
}

type inMemory struct {
	entry
	mu sync.RWMutex
}

func (pf *inMemory) Close() error {
	return nil
}

func (pf *inMemory) FileStat() *stat.FileStat {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.entry.FileStat
}

func (pf *inMemory) Offset() int64 {
	pf.mu.RLock()
	defer pf.mu.RUnlock()
	return pf.entry.Offset
}

func (pf *inMemory) IncreaseOffset(incr int) error {
	return pf.Set(pf.FileStat(), pf.Offset()+int64(incr))
}

func (pf *inMemory) Set(fileStat *stat.FileStat, offset int64) error {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	pf.entry.FileStat = fileStat
	pf.entry.Offset = offset
	return nil
}

func (pf *inMemory) SetOffset(offset int64) error {
	return pf.Set(pf.FileStat(), offset)
}

func (pf *inMemory) SetFileStat(fileStat *stat.FileStat) error {
	return pf.Set(fileStat, pf.Offset())
}
