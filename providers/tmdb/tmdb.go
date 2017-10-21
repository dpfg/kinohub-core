package tmdb

type TMDBClient interface {
	GetTVShowByID(id int)
	GetTVShowExternalIDS(id int)
	GetTVShowImages(id int)

	GetTVEpisode(tvID int, seasonNum int, episodeNum int)
	GetTVEpisodeImages(tvID int, seasonNum int, episodeNum int)
}
