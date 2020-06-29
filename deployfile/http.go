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
	"os"

	"github.com/jcmturner/gomvn/metadata"
	"github.com/jcmturner/gomvn/pom"
	"github.com/jcmturner/gomvn/repo"
)

func Upload(repoURL, groupID, artifactID, packaging, version, file, username, password string, cl *http.Client) error {
	if cl == nil {
		cl = http.DefaultClient
	}
	groupURL, versionURL, fileName := repo.ParseCoordinates(repoURL, groupID, artifactID, packaging, version)

	// open readers of the artifact
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("could not open artifact file: %v", err)
	}
	defer f.Close()
	rw := new(bytes.Buffer)
	t := io.TeeReader(f, rw)

	// PUT the artifact
	req, err := http.NewRequest("PUT", versionURL+fileName, t)
	if err != nil {
		return fmt.Errorf("could not create upload request for the artifact")
	}
	req.SetBasicAuth(username, password)
	resp, err := cl.Do(req)
	if err != nil {
		return fmt.Errorf("error uploading the artifact: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("return code when uploading the artifact %d", resp.StatusCode)
	}
	// PUT artifact hash files
	err = uploadHashFiles(rw, versionURL, fileName, username, password, cl)
	if err != nil {
		return err
	}

	// PUT POM
	p := pom.New(groupID, artifactID, version, packaging)
	pb, err := p.Marshal()
	if err != nil {
		return fmt.Errorf("error marshaling pom: %v", err)
	}
	pPath := fmt.Sprintf("%s/%s-%s.pom", versionURL, artifactID, version)
	prw := new(bytes.Buffer)
	pt := io.TeeReader(bytes.NewReader(pb), prw)
	req, err = http.NewRequest("PUT", pPath, pt)
	if err != nil {
		return fmt.Errorf("could not create upload request for %s : %v", pPath, err)
	}
	req.SetBasicAuth(username, password)
	resp, err = cl.Do(req)
	if err != nil {
		return fmt.Errorf("error uploading %s : %v", pPath, err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("uploading %s: return code %d", pPath, resp.StatusCode)
	}
	// PUT POM hash files
	pPath = fmt.Sprintf("%s%s/%s/", groupURL, artifactID, version)
	pomName := fmt.Sprintf("%s-%s.pom", artifactID, version)
	err = uploadHashFiles(prw, pPath, pomName, username, password, cl)
	if err != nil {
		return fmt.Errorf("error uploading pom hash files: %v", err)
	}

	// Generate and PUT metadata
	md, err := metadata.Generate(repoURL, groupID, artifactID, version, cl)
	if err != nil {
		return fmt.Errorf("error updating metadata: %v", err)
	}
	mdb, err := md.Marshal()
	if err != nil {
		return fmt.Errorf("error marshaling metadata: %v", err)
	}
	mdPath := fmt.Sprintf("%s/%s/%s", groupURL, artifactID, metadata.MavenMetadataFile)
	mdrw := new(bytes.Buffer)
	mdt := io.TeeReader(bytes.NewReader(mdb), mdrw)
	req, err = http.NewRequest("PUT", mdPath, mdt)
	if err != nil {
		return fmt.Errorf("could not create upload request for %s : %v", mdPath, err)
	}
	req.SetBasicAuth(username, password)
	resp, err = cl.Do(req)
	if err != nil {
		return fmt.Errorf("error uploading %s : %v", mdPath, err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("uploading %s: return code %d", mdPath, resp.StatusCode)
	}
	// PUT metadata hash files
	mdPath = fmt.Sprintf("%s/%s/", groupURL, artifactID)
	err = uploadHashFiles(mdrw, mdPath, metadata.MavenMetadataFile, username, password, cl)
	if err != nil {
		return fmt.Errorf("error uploading metadata hash files: %v", err)
	}
	return nil
}

func uploadHashFiles(r io.Reader, locationURL, filename, username, password string, cl *http.Client) error {
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
			return fmt.Errorf("could not generate request to put %s : %v", turl, err)
		}
		resp, err := cl.Do(req)
		if err != nil {
			return fmt.Errorf("error uploading %s : %v", turl, err)
		}
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("uploading %s: return code %d", turl, resp.StatusCode)
		}
		if i < len(readers) {
			t = readers[i]
		}
	}
	return nil
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
