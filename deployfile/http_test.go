package deployfile

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	testUsername     = "username"
	testPassword     = "password"
	groupID          = "log4j"
	artifactID       = "log4j"
	newVersion       = "1.2.18"
	mavenMetadataURL = "/log4j/log4j/maven-metadata.xml"
	mavenMetaData    = `<?xml version="1.0" encoding="UTF-8"?>
<metadata modelVersion="1.1.0">
  <groupId>log4j</groupId>
  <artifactId>log4j</artifactId>
  <versioning>
    <latest>1.2.17</latest>
    <release>1.2.17</release>
    <versions>
      <version>1.1.3</version>
      <version>1.2.4</version>
      <version>1.2.5</version>
      <version>1.2.6</version>
      <version>1.2.7</version>
      <version>1.2.8</version>
      <version>1.2.9</version>
      <version>1.2.11</version>
      <version>1.2.12</version>
      <version>1.2.13</version>
      <version>1.2.14</version>
      <version>1.2.15</version>
      <version>1.2.16</version>
      <version>1.2.17</version>
    </versions>
    <lastUpdated>20140318154402</lastUpdated>
  </versioning>
</metadata>
`
	mavenMetaDataSHA1 = "d290cc8eba0504881f1d165820c27fd7ea5b1d0f"
)

func testServer(badsha bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			switch r.RequestURI {
			case mavenMetadataURL:
				w.Write([]byte(mavenMetaData))
				return
			case mavenMetadataURL + ".sha1":
				if badsha {
					w.Write([]byte("invalid"))
					return
				}
				w.Write([]byte(mavenMetaDataSHA1))
				return
			}
		case http.MethodPut:
			u, p, ok := r.BasicAuth()
			if !ok || u != testUsername || p != testPassword {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}))
}

func TestUpload(t *testing.T) {
	s := testServer(false)
	defer s.Close()

	file, err := ioutil.TempFile(os.TempDir(), "gomvn-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("mockartifact")

	u, err := Upload(s.URL, groupID, artifactID, "jar", newVersion, file.Name(), testUsername, testPassword, nil)
	if err != nil {
		t.Error(err)
	}
	if len(u) != 9 {
		t.Errorf("expected number of uploaded URLs to be 9 actual: %d", len(u))
	}
}

func TestUploadInvalidMetadata(t *testing.T) {
	s := testServer(true)
	defer s.Close()

	file, err := ioutil.TempFile(os.TempDir(), "gomvn-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("mockartifact")

	_, err = Upload(s.URL, groupID, artifactID, "jar", newVersion, file.Name(), testUsername, testPassword, nil)
	if err == nil {
		t.Error("upload should have errored for an invalid sha1 of the metadata")
	}
}
