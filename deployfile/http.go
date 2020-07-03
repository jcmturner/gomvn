package deployfile

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/jcmturner/gomvn/metadata"
	"github.com/jcmturner/gomvn/pom"
	"github.com/jcmturner/gomvn/repo"
)

func Upload(repoURL, groupID, artifactID, packaging, version, file, username, password string, cl *http.Client) ([]*url.URL, error) {
	var uploaded []*url.URL
	if cl == nil {
		cl = http.DefaultClient
	}
	groupURL, versionURL, fileName := repo.ParseCoordinates(repoURL, groupID, artifactID, packaging, version)

	// open readers of the artifact
	f, err := os.Open(file)
	if err != nil {
		return uploaded, fmt.Errorf("could not open artifact file: %v", err)
	}
	defer f.Close()
	rw := new(bytes.Buffer)
	t := io.TeeReader(f, rw)

	// PUT the artifact
	u, err := url.Parse(versionURL + fileName)
	if err != nil {
		return uploaded, fmt.Errorf("target URL for artifact not avlid: %v", err)
	}
	req, err := http.NewRequest("PUT", u.String(), t)
	if err != nil {
		return uploaded, fmt.Errorf("could not create upload request for the artifact")
	}
	req.SetBasicAuth(username, password)
	resp, err := cl.Do(req)
	if err != nil {
		return uploaded, fmt.Errorf("error uploading the artifact: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return uploaded, fmt.Errorf("return code when uploading the artifact %d", resp.StatusCode)
	}
	uploaded = append(uploaded, u)
	// PUT artifact hash files
	us, err := uploadHashFiles(rw, versionURL, fileName, username, password, cl)
	if err != nil {
		return uploaded, err
	}
	uploaded = append(uploaded, us...)

	// PUT POM
	p := pom.New(groupID, artifactID, version, packaging)
	pb, err := p.Marshal()
	if err != nil {
		return uploaded, fmt.Errorf("error marshaling pom: %v", err)
	}
	purl, err := pom.URL(repoURL, groupID, artifactID, version)
	if err != nil {
		return uploaded, fmt.Errorf("URL for POM not valid: %v", err)
	}
	prw := new(bytes.Buffer)
	pt := io.TeeReader(bytes.NewReader(pb), prw)
	req, err = http.NewRequest("PUT", purl.String(), pt)
	if err != nil {
		return uploaded, fmt.Errorf("could not create upload request for %s : %v", purl.String(), err)
	}
	req.SetBasicAuth(username, password)
	resp, err = cl.Do(req)
	if err != nil {
		return uploaded, fmt.Errorf("error uploading %s : %v", purl.String(), err)
	}
	if resp.StatusCode != http.StatusCreated {
		return uploaded, fmt.Errorf("uploading %s: return code %d", purl.String(), resp.StatusCode)
	}
	uploaded = append(uploaded, purl)
	// PUT POM hash files
	pomName := fmt.Sprintf("%s-%s.pom", artifactID, version)
	us, err = uploadHashFiles(prw, versionURL, pomName, username, password, cl)
	if err != nil {
		return uploaded, fmt.Errorf("error uploading pom hash files: %v", err)
	}
	uploaded = append(uploaded, us...)

	// Generate and PUT metadata
	md, err := metadata.Generate(repoURL, groupID, artifactID, version, cl)
	if err != nil {
		return uploaded, fmt.Errorf("error updating metadata: %v", err)
	}
	mdb, err := md.Marshal()
	if err != nil {
		return uploaded, fmt.Errorf("error marshaling metadata: %v", err)
	}
	mdPath := fmt.Sprintf("%s%s/%s", groupURL, artifactID, metadata.MavenMetadataFile)
	murl, err := url.Parse(mdPath)
	if err != nil {
		return uploaded, fmt.Errorf("URL for metadata not valid: %v", err)
	}
	mdrw := new(bytes.Buffer)
	mdt := io.TeeReader(bytes.NewReader(mdb), mdrw)
	req, err = http.NewRequest("PUT", murl.String(), mdt)
	if err != nil {
		return uploaded, fmt.Errorf("could not create upload request for %s : %v", mdPath, err)
	}
	req.SetBasicAuth(username, password)
	resp, err = cl.Do(req)
	if err != nil {
		return uploaded, fmt.Errorf("error uploading %s : %v", mdPath, err)
	}
	if resp.StatusCode != http.StatusCreated {
		return uploaded, fmt.Errorf("uploading %s: return code %d", mdPath, resp.StatusCode)
	}
	uploaded = append(uploaded, murl)
	// PUT metadata hash files
	mdPath = fmt.Sprintf("%s%s/", groupURL, artifactID)
	us, err = uploadHashFiles(mdrw, mdPath, metadata.MavenMetadataFile, username, password, cl)
	if err != nil {
		return uploaded, fmt.Errorf("error uploading metadata hash files: %v", err)
	}
	uploaded = append(uploaded, us...)
	return uploaded, nil
}

func uploadHashFiles(r io.Reader, locationURL, filename, username, password string, cl *http.Client) ([]*url.URL, error) {
	var uploaded []*url.URL
	if cl == nil {
		cl = http.DefaultClient
	}
	hs := []struct {
		suffix string
		hash   hash.Hash
	}{
		{"sha1", sha1.New()},
		{"md5", md5.New()},
		//{"sha256", sha256.New()},
	}
	// Create a chain of readers to use as each is read to upload
	readers := make([]*bytes.Buffer, len(hs)-1)
	t := r
	for i := 1; i < len(hs); i++ {
		rw := new(bytes.Buffer)
		t = io.TeeReader(t, rw)
		readers[len(readers)-i] = rw
	}
	for i, h := range hs {
		turl := locationURL + filename + "." + h.suffix
		req, err := hashPutRequest(t, h.hash, turl, username, password)
		if err != nil {
			return uploaded, fmt.Errorf("could not generate request to put %s : %v", turl, err)
		}
		resp, err := cl.Do(req)
		if err != nil {
			return uploaded, fmt.Errorf("error uploading %s : %v", turl, err)
		}
		if resp.StatusCode != http.StatusCreated {
			return uploaded, fmt.Errorf("uploading %s: return code %d", turl, resp.StatusCode)
		}
		u, err := url.Parse(turl)
		if err == nil {
			uploaded = append(uploaded, u)
		}
		if i < len(readers) {
			t = readers[i]
		}
	}
	return uploaded, nil
}

func hashPutRequest(f io.Reader, h hash.Hash, targetURL, username, password string) (*http.Request, error) {
	_, err := io.Copy(h, f)
	if err != nil {
		return nil, err
	}
	hashb := h.Sum(nil)
	hexb := make([]byte, hex.EncodedLen(len(hashb)))
	hex.Encode(hexb, hashb)
	req, err := http.NewRequest("PUT", targetURL, bytes.NewReader(hexb))
	if err != nil {
		return req, err
	}
	req.SetBasicAuth(username, password)
	return req, nil
}
