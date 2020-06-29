package pom

import (
	"testing"
)

func TestMarsahl(t *testing.T) {
	p := New("grpID", "artID", "1.0.1", "jar")
	_, err := p.Marshal()
	if err != nil {
		t.Errorf("error mashaling pom: %v", err)
	}
}

//func TestPOM(t *testing.T) {
//	md, err := metadata.Get("http://central.maven.org/maven2", "log4j", "log4j")
//	if err != nil {
//		t.Fatalf("error getting repo metadata: %v", err)
//	}
//	p, err := RepoPOM("http://central.maven.org/maven2", "log4j", "log4j", md.Versioning.Latest)
//	if err != nil {
//		t.Fatalf("error getting pom: %v", err)
//	}
//	assert.Equal(t, "log4j", p.GroupID, "GroupID not as expected")
//	assert.Equal(t, "log4j", p.ArtifactID, "ArtifactID not as expected")
//}
