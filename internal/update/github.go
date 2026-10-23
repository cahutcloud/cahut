package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"cahut.cloud/internal/release"
)

// Limits on what the updater downloads.
const (
	maxListing   = 4 << 20
	maxChecksums = 64 << 10
	maxArchive   = 256 << 20
	maxBinary    = 512 << 20
)

// Release is a published release of the repository.
type Release struct {
	Version release.Version
	Tag     string
	URL     string           // its page
	Pre     bool             // marked as a pre-release
	Assets  map[string]asset // by name
}

type asset struct {
	URL  string `json:"url"` // in the API, which gives the file with Accept: application/octet-stream
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// releases lists the published releases, newest first. Drafts, which only
// people with write access see, are left out.
func (u *Updater) releases(ctx context.Context) ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases?per_page=50", u.API, u.Repo)
	body, err := u.get(ctx, url, "application/vnd.github+json", maxListing)
	if err != nil {
		return nil, fmt.Errorf("read the releases of %s: %w", u.Repo, err)
	}
	var listed []struct {
		Tag    string  `json:"tag_name"`
		URL    string  `json:"html_url"`
		Draft  bool    `json:"draft"`
		Pre    bool    `json:"prerelease"`
		Assets []asset `json:"assets"`
	}
	if err := json.Unmarshal(body, &listed); err != nil {
		return nil, fmt.Errorf("read the releases of %s: %w", u.Repo, err)
	}
	var out []Release
	for _, l := range listed {
		v, err := release.ParseVersion(l.Tag)
		if l.Draft || err != nil {
			continue
		}
		r := Release{Version: v, Tag: l.Tag, URL: l.URL, Pre: l.Pre || v.Pre != "", Assets: map[string]asset{}}
		for _, a := range l.Assets {
			r.Assets[a.Name] = a
		}
		out = append(out, r)
	}
	slices.SortFunc(out, func(a, b Release) int { return b.Version.Compare(a.Version) })
	return out, nil
}

// download returns the asset name of r, at most limit bytes.
func (u *Updater) download(ctx context.Context, r Release, name string, limit int64) ([]byte, error) {
	a, ok := r.Assets[name]
	if !ok {
		return nil, fmt.Errorf("release %s has no %s", r.Tag, name)
	}
	if a.Size > limit {
		return nil, fmt.Errorf("%s holds %d bytes, more than the %d an update accepts", name, a.Size, limit)
	}
	data, err := u.get(ctx, a.URL, "application/octet-stream", limit)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", name, err)
	}
	return data, nil
}

// get reads url with the token, if any. GitHub answers an asset with a
// redirect to its storage, where Go's client sends no Authorization header.
func (u *Updater) get(ctx context.Context, url, accept string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "cahut-update/"+u.Running)
	if u.Token != "" {
		req.Header.Set("Authorization", "Bearer "+u.Token)
	}
	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound && u.Token == "" {
			return nil, errors.New("not found (a private repository needs GITHUB_TOKEN in /etc/cahut/update.env)")
		}
		return nil, fmt.Errorf("GitHub answers %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("the answer holds more than %d bytes", limit)
	}
	return data, nil
}
