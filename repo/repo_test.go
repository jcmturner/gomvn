package repo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseCoordinates(t *testing.T) {
	groupURL, versionURL, filename := ParseCoordinates("https://repourl", "gro.up.id",
		"artifactid", "pkg", "1.0.0")
	expected := "https://repourl/gro/up/id/"
	if groupURL != expected {
		t.Errorf("groupURL; expected %s ; got %s", expected, groupURL)
	}
	expected = "https://repourl/gro/up/id/artifactid/1.0.0/"
	if versionURL != expected {
		t.Errorf("versionURL; expected %s ; got %s", expected, versionURL)
	}
	expected = "artifactid-1.0.0.pkg"
	if filename != expected {
		t.Errorf("filename; expected %s ; got %s", expected, filename)
	}
}

func TestSHA1(t *testing.T) {
	s := testServer()
	defer s.Close()
	b := []byte(mavenMetaData)
	ok, err := SHA1(s.URL+mavenMetadataURL, b, nil)
	if err != nil {
		t.Errorf("error from SHA1 check: %v", err)
	}
	if !ok {
		t.Errorf("SHA1 check did not pass when it should have")
	}
	// Provide invalid bytes for the sha by removing the first byte
	ok, err = SHA1(s.URL+mavenMetadataURL, b[1:], nil)
	if err == nil {
		t.Error("error from SHA1 is nil but should not have been")
	}
	if ok {
		t.Errorf("SHA1 check passed when it should not have")
	}
}

const (
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

func testServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			switch r.RequestURI {
			case mavenMetadataURL:
				w.Write([]byte(mavenMetaData))
				return
			case mavenMetadataURL + ".sha1":
				w.Write([]byte(mavenMetaDataSHA1))
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}))
}
