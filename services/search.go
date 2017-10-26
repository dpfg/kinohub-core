package services

import (
	"github.com/dpfg/kinohub-core/domain"
	"github.com/dpfg/kinohub-core/providers/kinopub"
)

// ContentSearch provides a way to find available media streams
type ContentSearch interface {
	Search(q string) ([]domain.SearchResult, error)
}

type ContentSearchImpl struct {
	Kinopub kinopub.KinoPubClient
}

func (cs ContentSearchImpl) Search(q string) ([]domain.SearchResult, error) {
	r, err := cs.Kinopub.SearchItemBy(kinopub.ItemsFilter{
		Title: q,
	})

	if err != nil {
		return nil, err
	}

	result := make([]domain.SearchResult, 0)
	for _, item := range r {

		result = append(result, domain.SearchResult{
			Type:  item.DomainType(),
			Title: item.Title,
		})
	}

	return result, nil
}
