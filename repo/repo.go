package repo

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ParseCoordinates(repoURL, groupID, artifactID, packaging, version string) (groupURL, versionURL, filename string) {
	groupPath := strings.Join(strings.Split(groupID, "."), "/")
	filename = fmt.Sprintf("%s-%s.%s", artifactID, version, packaging)
	groupURL = fmt.Sprintf("%s/%s/", strings.TrimRight(repoURL, "/"), groupPath)
	versionURL = fmt.Sprintf("%s/%s/%s/%s/", strings.TrimRight(repoURL, "/"), groupPath, artifactID, version)
	return
}

func SHA1(furl string, b []byte, cl *http.Client) (bool, error) {
	sha1url := furl + ".sha1"
	req, err := http.NewRequest("GET", sha1url, nil)
	if err != nil {
		return false, fmt.Errorf("could not form request to check sha1 %s: %v", furl, err)
	}
	if cl == nil {
		cl = http.DefaultClient
	}
	resp, err := cl.Do(req)
	if err != nil {
		return false, fmt.Errorf("error fetching sha1 file: %s: %v", sha1url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("http response %d downloading sha1 file (%s)", resp.StatusCode, sha1url)
	}
	r := bufio.NewReader(resp.Body)
	defer resp.Body.Close()
	mdsha1, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("error reading content of sha1 file %s: %v", sha1url, err)
	}
	mdsha1 = strings.ToLower(strings.TrimSpace(strings.Fields(mdsha1)[0]))

	// Check the md5 of the metadata
	hash := sha1.New()
	hash.Write(b)
	h := hex.EncodeToString(hash.Sum(nil))
	if strings.ToLower(h) != mdsha1 {
		return false, fmt.Errorf("checksum (%s) does not match. expected: %s got: %s", sha1url, mdsha1, h)
	}
	return true, nil
}
