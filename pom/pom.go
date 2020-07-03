package pom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/jcmturner/gomvn/repo"
)

const (
	pomFile      = "pom.xml"
	modelVersion = "4.0.0"
)

type POM struct {
	XMLName      xml.Name      `xml:"project"`
	ModelVersion string        `xml:"modelVersion"`
	GroupID      string        `xml:"groupId"`
	ArtifactID   string        `xml:"artifactId"`
	Version      string        `xml:"version"`
	Packaging    string        `xml:"packaging"`
	Description  string        `xml:"description,omitempty"`
	URL          string        `xml:"url,omitempty"`
	Name         string        `xml:"name,omitempty"`
	Licenses     *[]License    `xml:"licenses>license,omitempty"`
	Dependencies *[]Dependency `xml:"dependencies>dependency,omitempty"`
	Repositories *[]Repository `xml:"repositories>repository,omitempty"`
}

type License struct {
	Name         string `xml:"name"`
	URL          string `xml:"url"`
	Distribution string `xml:"distribution"`
}

type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Type       string `xml:"type"`
	Scope      string `xml:"scope"`
	Optional   bool   `xml:"optional"`
}

type Repository struct {
	ID        string     `xml:"id"`
	Name      string     `xml:"name"`
	URL       string     `xml:"url"`
	Layout    string     `xml:"layout"`
	Snapshots RepoPolicy `xml:"snapshots"`
	Releases  RepoPolicy `xml:"releases"`
}

type RepoPolicy struct {
	Enabled        bool   `xml:"enabled"`
	UpdatePolicy   string `xml:"updatePolicy"`
	ChecksumPolicy string `xml:"checksumPolicy"`
}

func New(groupID, artifactID, version, packaging string) POM {
	return POM{
		ModelVersion: modelVersion,
		GroupID:      groupID,
		ArtifactID:   artifactID,
		Version:      version,
		Packaging:    packaging,
	}
}

func URL(repoURL, groupID, artifactID, version string) (*url.URL, error) {
	_, versionURL, _ := repo.ParseCoordinates(repoURL, groupID, artifactID, "", version)
	return url.Parse(fmt.Sprintf("%s%s-%s.pom", versionURL, artifactID, version))
}

func Get(repoURL, groupID, artifactID, version string, cl *http.Client) (p POM, err error) {
	pomURL, err := URL(repoURL, groupID, artifactID, version)
	if err != nil {
		return
	}

	// Get the POM file
	req, err := http.NewRequest("GET", pomURL.String(), nil)
	if err != nil {
		err = fmt.Errorf("error forming request of %s: %v", pomURL, err)
		return
	}
	if cl == nil {
		cl = http.DefaultClient
	}
	resp, err := cl.Do(req)
	if err != nil {
		err = fmt.Errorf("error getting %s: %v", pomURL, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http response %d downloading POM file", resp.StatusCode)
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading body from %s: %v", pomURL, err)
		return
	}
	resp.Body.Close()

	ok, err := repo.SHA1(pomURL.String(), b, cl)
	if !ok || err != nil {
		err = fmt.Errorf("integrity check failed: %v", err)
		return
	}

	// unmarshal bytes into POM type
	err = p.Unmarshal(b)
	return
}

func Load(path string) (POM, error) {
	var p POM
	fh, err := os.Open(path)
	if err != nil {
		return p, fmt.Errorf("could not open POM file at %s: %v", path, err)
	}
	defer fh.Close()
	decoder := xml.NewDecoder(fh)
	err = decoder.Decode(&p)
	if err != nil {
		return p, fmt.Errorf("could not decode POM file at %s: %v", path, err)
	}
	return p, nil
}

func (p *POM) Marshal() ([]byte, error) {
	b := []byte(xml.Header)
	pb, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		return pb, err
	}
	b = append(b, pb...)
	return b, nil
}

func (p *POM) Unmarshal(b []byte) error {
	rdr := bytes.NewReader(b)
	decoder := xml.NewDecoder(rdr)
	err := decoder.Decode(&p)
	if err != nil {
		return fmt.Errorf("error unmarshaling pom: %v", err)
	}
	return nil
}
