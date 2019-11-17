package kinopub

import (
	"encoding/json"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

type fix struct {
	KinoPubID int `json:"kinopub_id,omitempty"`
	ImdbID    int `json:"imdb_id,omitempty"`
}

var (
	fixes []fix
)

type fixer struct {
	logger *logrus.Entry
}

func (f *fixer) fixID(item *Item) {
	f.logger.Debugf("attempt to fix: %v", item.Imdb)
	// if item.Imdb == 0 {
	item.Imdb = f.findFix(item.ID)
	// }
}

func (f *fixer) findFix(id int) int {
	if fixes == nil {
		fixes = f.loadFixes()
		f.logger.Debugf("Id fixes are loaded: %d", len(fixes))
	}

	for _, fix := range fixes {
		if fix.KinoPubID == id {
			f.logger.Debugf("Found fix: %d -> %d", fix.KinoPubID, fix.ImdbID)
			return fix.ImdbID
		}
	}

	return 0
}

func (f *fixer) loadFixes() []fix {
	dat, err := ioutil.ReadFile(".data/imdb-fixes.json")
	if err != nil {
		f.logger.Errorf("Unable to load id fixes: %s", err.Error())
		return nil
	}

	var value []fix
	err = json.Unmarshal(dat, &value)
	if err != nil {
		return nil
	}
	return value
}
