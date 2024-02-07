package internal_db

import (
	"bufio"
	"net/http"
	"os"
	"strings"
	"sync"
)

type InternalDB struct {
	vDB  map[string]struct{}
	kvDB map[string]string
	kind uint8
	mu   sync.RWMutex
	file string
}

func (idb *InternalDB) Add(key, value string) {
	idb.mu.Lock()
	switch idb.kind {
	case ValueDb:
		idb.vDB[key] = struct{}{}
	case KeyValueDb:
		idb.kvDB[key] = value
	}
	idb.mu.Unlock()
}

func (idb *InternalDB) Del(key string) {
	idb.mu.Lock()
	switch idb.kind {
	case ValueDb:
		delete(idb.vDB, key)
	case KeyValueDb:
		delete(idb.kvDB, key)
	}
	idb.mu.Unlock()
}

func (idb *InternalDB) GetContent() string {
	idb.mu.RLock()
	var res string
	switch idb.kind {
	case ValueDb:
		for k, _ := range idb.vDB {
			res += k + "\n"
		}
	case KeyValueDb:
		for k, v := range idb.kvDB {
			res += k + " " + v + "\n"
		}
	}
	idb.mu.RUnlock()
	return res
}

func (idb *InternalDB) unsafeReset() {
	switch idb.kind {
	case ValueDb:
		idb.vDB = make(map[string]struct{})
	case KeyValueDb:
		idb.kvDB = make(map[string]string)
	}
}

func (idb *InternalDB) Load() error {
	var err error
	var fd *os.File
	var scanner *bufio.Scanner
	var k, v string

	if fd, err = os.Open(idb.file); err != nil {
		return err
	}
	scanner = bufio.NewScanner(fd)

	idb.mu.Lock()
	idb.unsafeReset()
	switch idb.kind {
	case ValueDb:
		for scanner.Scan() {
			k = strings.TrimSpace(scanner.Text())
			idb.vDB[k] = struct{}{}
		}
		err = scanner.Err()
	case KeyValueDb:
		for scanner.Scan() {
			k = strings.Split(strings.TrimSpace(scanner.Text()), " ")[0]
			v = strings.Split(strings.TrimSpace(scanner.Text()), " ")[1]
			idb.kvDB[k] = v
		}
		err = scanner.Err()
	}
	idb.mu.Unlock()
	if err != nil {
		_ = fd.Close()
		return err
	}
	return fd.Close()
}

func (idb *InternalDB) Save() error {
	var err error
	var fd *os.File

	if fd, err = os.OpenFile(idb.file, os.O_WRONLY|os.O_SYNC|os.O_TRUNC, 0600); err != nil {
		return err
	}

	if _, err = fd.WriteString(idb.GetContent()); err != nil {
		_ = fd.Close()
		return err
	}

	return fd.Close()
}

func (idb *InternalDB) HttpAddHandler(w http.ResponseWriter, r *http.Request) {
	idb.httpGenericHandler(true, w, r)
}
func (idb *InternalDB) HttpDelHandler(w http.ResponseWriter, r *http.Request) {
	idb.httpGenericHandler(false, w, r)
}
func (idb *InternalDB) HttpGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("Method not allowed"))
		return
	}
	_, _ = w.Write([]byte(idb.GetContent()))
}

func (idb *InternalDB) httpGenericHandler(add bool, w http.ResponseWriter, r *http.Request) {
	var err error
	var scanner *bufio.Scanner

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("Method not allowed"))
		return
	}

	scanner = bufio.NewScanner(r.Body)

	for scanner.Scan() {
		switch idb.kind {
		case ValueDb:
			if add {
				idb.Add(strings.TrimSpace(scanner.Text()), "")
			} else {
				idb.Del(strings.TrimSpace(scanner.Text()))
			}
		case KeyValueDb:
			if add {
				idb.Add(strings.Split(strings.TrimSpace(scanner.Text()), " ")[0], strings.Split(strings.TrimSpace(scanner.Text()), " ")[1])
			} else {
				idb.Del(strings.TrimSpace(scanner.Text()))
			}
		}
	}

	_ = r.Body.Close()

	if scanner.Err() != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error: " + err.Error()))
		return
	} else {
		_, _ = w.Write([]byte("OK"))
	}

}

func NewInternalDb(kind uint8, file string) (*InternalDB, error) {
	var idb *InternalDB = new(InternalDB)
	switch kind {
	case ValueDb, KeyValueDb:
		idb.kind = kind
		idb.file = file
		idb.unsafeReset()
		return idb, nil
	default:
		return nil, ErrUnknownDbKind
	}
}
