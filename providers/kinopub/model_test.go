package kinopub

import "testing"

func TestItem_ImdbID(t *testing.T) {
	item := Item{Imdb: 898266}
	if item.ImdbID() != "tt0898266" {
		t.Errorf("Invalid IMDB ID: %s. Expected tt0898266.", item.ImdbID())
	}
}
