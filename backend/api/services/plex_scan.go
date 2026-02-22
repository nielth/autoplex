package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

type plexLibrarySectionsResponse struct {
	Directories []plexLibrarySection `xml:"Directory"`
}

type plexLibrarySection struct {
	Key   string `xml:"key,attr"`
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr"`
}

func TriggerMoviesAndShowsScan() ([]string, error) {
	plexURL, err := requiredEnv("PLEX_URL")
	if err != nil {
		return nil, err
	}
	plexToken, err := requiredEnv("PLEX_TOKEN")
	if err != nil {
		return nil, err
	}

	plexURL = strings.TrimRight(plexURL, "/")
	sections, err := plexListLibrarySections(plexURL, plexToken)
	if err != nil {
		return nil, err
	}

	movieSection, foundMovie := pickPlexSectionByType(
		sections,
		"movie",
		[]string{"movies", "movie"},
	)
	showSection, foundShow := pickPlexSectionByType(
		sections,
		"show",
		[]string{"tv shows", "tv show", "shows", "television", "series"},
	)

	if !foundMovie || !foundShow {
		return nil, fmt.Errorf("could not find both Plex movie and show libraries")
	}

	sectionsToRefresh := []plexLibrarySection{movieSection}
	if showSection.Key != movieSection.Key {
		sectionsToRefresh = append(sectionsToRefresh, showSection)
	}

	scannedSections := make([]string, 0, len(sectionsToRefresh))
	for _, section := range sectionsToRefresh {
		if err := plexRefreshLibrarySection(plexURL, plexToken, section.Key); err != nil {
			return nil, fmt.Errorf("failed to refresh Plex section %q: %w", section.Title, err)
		}
		scannedSections = append(scannedSections, section.Title)
	}

	return scannedSections, nil
}

func pickPlexSectionByType(sections []plexLibrarySection, targetType string, preferredTitles []string) (plexLibrarySection, bool) {
	candidates := make([]plexLibrarySection, 0)
	for _, section := range sections {
		if strings.EqualFold(strings.TrimSpace(section.Type), targetType) {
			candidates = append(candidates, section)
		}
	}

	if len(candidates) == 0 {
		return plexLibrarySection{}, false
	}

	for _, section := range candidates {
		title := strings.ToLower(strings.TrimSpace(section.Title))
		if slices.Contains(preferredTitles, title) {
			return section, true
		}
	}

	return candidates[0], true
}

func plexListLibrarySections(plexURL string, plexToken string) ([]plexLibrarySection, error) {
	sectionsURL := fmt.Sprintf("%s/library/sections?X-Plex-Token=%s", plexURL, url.QueryEscape(plexToken))

	req, err := http.NewRequest(http.MethodGet, sectionsURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("plex library sections request failed with status %d: %s", res.StatusCode, string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var parsed plexLibrarySectionsResponse
	if err := xml.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse plex library sections response: %w", err)
	}

	return parsed.Directories, nil
}

func plexRefreshLibrarySection(plexURL string, plexToken string, sectionKey string) error {
	refreshURL := fmt.Sprintf(
		"%s/library/sections/%s/refresh?X-Plex-Token=%s",
		plexURL,
		url.PathEscape(strings.TrimSpace(sectionKey)),
		url.QueryEscape(plexToken),
	)

	req, err := http.NewRequest(http.MethodGet, refreshURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("plex refresh request failed with status %d: %s", res.StatusCode, string(body))
	}

	return nil
}
