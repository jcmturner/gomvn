package metadata

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jcmturner/gomvn/repo"
	"github.com/jcmturner/gomvn/version"
)

const (
	LastUpdatedLayout = "20060102150405"
	MavenMetadataFile = "maven-metadata.xml"
	modelVersion      = "1.1.0"
)

type MetaData struct {
	XMLName      xml.Name         `xml:"metadata"`
	ModelVersion string           `xml:"modelVersion,attr,omitempty"`
	GroupID      string           `xml:"groupId"`
	ArtifactID   string           `xml:"artifactId"`
	Version      *version.Version `xml:"version,omitempty"`
	Versioning   Versioning       `xml:"versioning"`
}

type Versioning struct {
	Latest           *version.Version  `xml:"latest,omitempty"`
	Release          *version.Version  `xml:"release,omitempty"`
	Snapshot         *Snapshot         `xml:"snapshot,omitempty"`
	Versions         *version.Versions `xml:"versions>version"`
	LastUpdated      *TimeStamp        `xml:"lastUpdated"`
	SnapshotVersions []SnapshotVersion `xml:"snapshotVersions,omitempty"`
}

type Snapshot struct {
	TimeStamp   TimeStamp `xml:"timestamp"`
	BuildNumber int       `xml:"buildNumber"`
	LocalCopy   bool      `xml:"localCopy"`
}

type SnapshotVersion struct {
	Classifier string `xml:"classifier"`
	Extension  string `xml:"extension"`
	Value      string `xml:"value"`
	Updated    string `xml:"updated"`
}

type TimeStamp struct {
	time.Time
}

func New(groupID, artifactID string) MetaData {
	return MetaData{
		ModelVersion: modelVersion,
		GroupID:      groupID,
		ArtifactID:   artifactID,
	}
}

func (t *TimeStamp) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(t.Format(LastUpdatedLayout), start)
}

func (t *TimeStamp) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	tt, err := time.Parse(LastUpdatedLayout, s)
	if err != nil {
		return err
	}
	*t = TimeStamp{tt}
	return nil
}

func (m *MetaData) Marshal() ([]byte, error) {
	b := []byte(xml.Header)
	mb, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		return mb, err
	}
	b = append(b, mb...)
	return b, nil
}

func (m *MetaData) Unmarshal(b []byte) error {
	rdr := bytes.NewReader(b)
	decoder := xml.NewDecoder(rdr)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("error unmarshaling metadata: %v", err)
	}
	sort.Sort(m.Versioning.Versions)
	return nil
}

type NotFound struct {
	ErrorString string
}

func (e NotFound) Error() string {
	return e.ErrorString
}

func Get(repoURL, groupID, artifactID string, cl *http.Client) (md MetaData, err error) {
	groupPath := strings.Join(strings.Split(groupID, "."), "/")
	url := fmt.Sprintf("%s/%s/%s/%s", strings.TrimRight(repoURL, "/"), groupPath, artifactID, MavenMetadataFile)

	// Get the metadata
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("error forming request of %s: %v", url, err)
		return
	}
	if cl == nil {
		cl = http.DefaultClient
	}
	resp, err := cl.Do(req)
	if err != nil {
		err = fmt.Errorf("error getting %s: %v", url, err)
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		err = NotFound{
			ErrorString: fmt.Sprintf("http response %d downloading metadata (%s)", resp.StatusCode, url),
		}
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http response %d downloading metadata (%s)", resp.StatusCode, url)
		return
	}
	mb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading body from %s: %v", url, err)
		return
	}
	resp.Body.Close()

	ok, err := repo.SHA1(url, mb, cl)
	if !ok || err != nil {
		err = fmt.Errorf("integrity check failed: %v", err)
		return
	}

	// unmarshal bytes into MetaData type
	err = md.Unmarshal(mb)
	return
}

func Generate(repo, groupID, artifactID, newVersion string, cl *http.Client) (MetaData, error) {
	var md MetaData
	// Get the current hosted metadata
	md, err := Get(repo, groupID, artifactID, cl)
	if err != nil {
		if _, ok := err.(NotFound); ok {
			// No current metadata so create a new one
			md = New(groupID, artifactID)
			md.Versioning = Versioning{
				Versions: new(version.Versions),
			}
		} else {
			return md, fmt.Errorf("error getting existing metadata: %v", err)
		}
	}
	// Add the version, resort and update the latest versions
	nv, err := version.New(newVersion)
	if err != nil {
		return md, err
	}
	*md.Versioning.Versions = append(*md.Versioning.Versions, nv)
	sort.Sort(md.Versioning.Versions)
	md.Versioning.Latest = &((*md.Versioning.Versions)[len(*md.Versioning.Versions)-1])
	md.Versioning.Release = md.Versioning.Latest
	md.Version = md.Versioning.Latest
	// Set the last update timestamp
	md.Versioning.LastUpdated = &TimeStamp{time.Now().UTC()}
	return md, nil
}
