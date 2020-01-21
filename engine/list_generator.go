package engine

import (
	"time"

	"github.com/cloudsonic/sonic-server/model"
)

type ListGenerator interface {
	GetNewest(offset int, size int) (Entries, error)
	GetRecent(offset int, size int) (Entries, error)
	GetFrequent(offset int, size int) (Entries, error)
	GetHighest(offset int, size int) (Entries, error)
	GetRandom(offset int, size int) (Entries, error)
	GetByName(offset int, size int) (Entries, error)
	GetByArtist(offset int, size int) (Entries, error)
	GetStarred(offset int, size int) (Entries, error)
	GetAllStarred() (artists Entries, albums Entries, mediaFiles Entries, err error)
	GetNowPlaying() (Entries, error)
	GetRandomSongs(size int, genre string) (Entries, error)
}

func NewListGenerator(ds model.DataStore, npRepo NowPlayingRepository) ListGenerator {
	return &listGenerator{ds, npRepo}
}

type listGenerator struct {
	ds     model.DataStore
	npRepo NowPlayingRepository
}

// TODO: Only return albums that have the Sort field != empty
func (g *listGenerator) query(qo model.QueryOptions, offset int, size int) (Entries, error) {
	qo.Offset = offset
	qo.Max = size
	albums, err := g.ds.Album().GetAll(qo)

	return FromAlbums(albums), err
}

func (g *listGenerator) GetNewest(offset int, size int) (Entries, error) {
	qo := model.QueryOptions{Sort: "CreatedAt", Order: "desc"}
	return g.query(qo, offset, size)
}

func (g *listGenerator) GetRecent(offset int, size int) (Entries, error) {
	qo := model.QueryOptions{Sort: "PlayDate", Order: "desc"}
	return g.query(qo, offset, size)
}

func (g *listGenerator) GetFrequent(offset int, size int) (Entries, error) {
	qo := model.QueryOptions{Sort: "PlayCount", Order: "desc"}
	return g.query(qo, offset, size)
}

func (g *listGenerator) GetHighest(offset int, size int) (Entries, error) {
	qo := model.QueryOptions{Sort: "Rating", Order: "desc"}
	return g.query(qo, offset, size)
}

func (g *listGenerator) GetByName(offset int, size int) (Entries, error) {
	qo := model.QueryOptions{Sort: "Name"}
	return g.query(qo, offset, size)
}

func (g *listGenerator) GetByArtist(offset int, size int) (Entries, error) {
	qo := model.QueryOptions{Sort: "Artist"}
	return g.query(qo, offset, size)
}

func (g *listGenerator) GetRandom(offset int, size int) (Entries, error) {
	albums, err := g.ds.Album().GetRandom(model.QueryOptions{Max: size, Offset: offset})
	if err != nil {
		return nil, err
	}

	r := make(Entries, len(albums))
	for i, al := range albums {
		r[i] = FromAlbum(&al)
	}
	return r, nil
}

func (g *listGenerator) GetRandomSongs(size int, genre string) (Entries, error) {
	options := model.QueryOptions{Max: size}
	if genre != "" {
		options.Filters = map[string]interface{}{"genre": genre}
	}
	mediaFiles, err := g.ds.MediaFile().GetRandom(options)
	if err != nil {
		return nil, err
	}

	r := make(Entries, len(mediaFiles))
	for i, mf := range mediaFiles {
		r[i] = FromMediaFile(&mf)
	}
	return r, nil
}

func (g *listGenerator) GetStarred(offset int, size int) (Entries, error) {
	qo := model.QueryOptions{Offset: offset, Max: size, Sort: "starred_at", Order: "desc"}
	albums, err := g.ds.Album().GetStarred(qo)
	if err != nil {
		return nil, err
	}

	return FromAlbums(albums), nil
}

func (g *listGenerator) GetAllStarred() (artists Entries, albums Entries, mediaFiles Entries, err error) {
	options := model.QueryOptions{Sort: "starred_at", Order: "desc"}

	ars, err := g.ds.Artist().GetStarred(options)
	if err != nil {
		return nil, nil, nil, err
	}

	als, err := g.ds.Album().GetStarred(options)
	if err != nil {
		return nil, nil, nil, err
	}

	mfs, err := g.ds.MediaFile().GetStarred(options)
	if err != nil {
		return nil, nil, nil, err
	}

	artists = FromArtists(ars)
	albums = FromAlbums(als)
	mediaFiles = FromMediaFiles(mfs)

	return
}

func (g *listGenerator) GetNowPlaying() (Entries, error) {
	npInfo, err := g.npRepo.GetAll()
	if err != nil {
		return nil, err
	}
	entries := make(Entries, len(npInfo))
	for i, np := range npInfo {
		mf, err := g.ds.MediaFile().Get(np.TrackID)
		if err != nil {
			return nil, err
		}
		entries[i] = FromMediaFile(mf)
		entries[i].UserName = np.Username
		entries[i].MinutesAgo = int(time.Now().Sub(np.Start).Minutes())
		entries[i].PlayerId = np.PlayerId
		entries[i].PlayerName = np.PlayerName

	}
	return entries, nil
}
