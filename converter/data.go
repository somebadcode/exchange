package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Data struct {
	Disclaimer string                 `json:"disclaimer,omitempty"`
	License    string                 `json:"license,omitempty"`
	Timestamp  int                    `json:"timestamp,omitempty"`
	Base       string                 `json:"base,omitempty"`
	Rates      map[string]json.Number `json:"rates,omitempty"`
}

func (d *Data) Time() time.Time {
	return time.Unix(int64(d.Timestamp), 0)
}

func (d *Data) Save(filename string) error {
	dir, _ := filepath.Split(filename)

	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0o777); err != nil {
			return fmt.Errorf("cannot create directory %s: %v", dir, err)
		}
	} else if err != nil {
		return fmt.Errorf("cannot read directory %s: %v", dir, err)
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening cached data: %w", err)

	}

	defer f.Close()

	err = json.NewEncoder(f).Encode(d)
	if err != nil {
		return fmt.Errorf("encoding cached data: %w", err)
	}

	return nil
}

func (d *Data) Load(filename string) error {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening cached data: %w", err)
	}

	defer f.Close()

	err = json.NewDecoder(f).Decode(d)
	if err != nil {
		return fmt.Errorf("decoding cached data: %w", err)
	}

	return nil
}
