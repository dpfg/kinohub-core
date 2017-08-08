package providers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
)

type PreferenceStorage interface {
	Load(key string, value interface{}) error
	Save(key string, value interface{}) error
}

type JSONPreferenceStorage struct {
	Path string
}

func (jts JSONPreferenceStorage) getFilePath(key string) string {
	return path.Join(jts.Path, key+".json")
}

func (jts JSONPreferenceStorage) ensureFileExists(filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			// create new file
			if _, err = os.Create(filePath); err != nil {
				return err
			}
		} else {
			// unexpected error
			return err
		}
	}

	return nil
}

func (jts JSONPreferenceStorage) Load(key string, value interface{}) error {
	filePath := jts.getFilePath(key)
	err := jts.ensureFileExists(filePath)

	if err != nil {
		return errors.New("There is no connection to " + key)
	}

	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if len(dat) == 0 {
		return errors.New("There is no connection to " + key)
	}

	err = json.Unmarshal(dat, value)
	if err != nil {
		return err
	}

	return nil
}

// Save token to json file.
func (jts JSONPreferenceStorage) Save(key string, value interface{}) error {
	dat, err := json.Marshal(value)
	if err != nil {
		return err
	}

	filePath := jts.getFilePath(key)
	err = jts.ensureFileExists(filePath)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, dat, os.ModeExclusive)
}
