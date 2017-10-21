package services

import (
	"github.com/dpfg/kinohub-core/domain"
)

// ContentSearch provides a way to find available media streams
type ContentSearch interface {
	SearchTV(q string) []domain.Show
}

type ContentSearchImpl struct {
}
