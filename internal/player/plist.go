package player

const (
	// defines number of item to preserve in the playlist
	maxHistorySize = 3

	// PositionNone indicates that nothing is playing
	PositionNone = -1
)

// MediaEntry describes element in the player
type MediaEntry struct {
	RawURL    string                 `json:"url,omitempty"`
	MediaInfo map[string]interface{} `json:"media_info,omitempty"`
}

// PList holds the list of media items with pointer to the playing one
type PList struct {
	CurrentIndex int          `json:"current_index"`
	Entries      []MediaEntry `json:"entries,omitempty"`
	AutoPlay     bool         `json:"auto_play,omitempty"`
}

// Current return pointer to a current media entry
func (pl *PList) Current() *MediaEntry {
	if pl.CurrentIndex == PositionNone {
		return nil
	}

	return &pl.Entries[pl.CurrentIndex]
}

// Next moves pointer of a current element to the next element in the list and
// returns the media entry
func (pl *PList) Next() *MediaEntry {
	nextIndex := pl.CurrentIndex + 1

	length := len(pl.Entries)
	if length < nextIndex+1 {
		pl.CurrentIndex = PositionNone
		return nil
	}

	return pl.Select(nextIndex)
}

// Select move pointer to a current element to the specific element refered by its index
// and return the media entry
func (pl *PList) Select(position int) *MediaEntry {
	plistSize := len(pl.Entries)
	if plistSize == 0 {
		return nil
	}

	if position < 0 || plistSize < position {
		return nil
	}

	pl.CurrentIndex = position

	entry := &pl.Entries[position]

	pl.cleanUpHistory()

	return entry
}

// AddEntry adds a new media entry to the end of the playlist.
func (pl *PList) AddEntry(entry MediaEntry) int {
	pl.Entries = append(pl.Entries, entry)
	return len(pl.Entries) - 1
}

func (pl *PList) cleanUpHistory() {
	if pl.CurrentIndex == PositionNone {
		return
	}

	historySize := pl.CurrentIndex + 1
	if historySize < maxHistorySize {
		return
	}

	dropIndex := historySize - maxHistorySize - 1
	if dropIndex < 0 {
		dropIndex = 0
	}

	pl.Entries = pl.Entries[dropIndex:]
	pl.CurrentIndex = pl.CurrentIndex - dropIndex

}

// NewPlayList creates new play list with default settings
func NewPlayList(entries []MediaEntry) *PList {
	return &PList{CurrentIndex: PositionNone, AutoPlay: true, Entries: entries}
}
