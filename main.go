package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/jcmturner/gomvn/deployfile"
)

func main() {
	repourl := flag.String("repourl", "", "URL to the maven repository")
	group := flag.String("group", "", "maven group identifier")
	artifact := flag.String("artifact", "", "artifact identifier")
	pkg := flag.String("ext", "", "file extension")
	version := flag.String("version", "", "artifact version")
	file := flag.String("file", "", "file to upload")
	username := flag.String("username", "", "username for authentication to the repository")
	password := flag.String("password", "", "password for authentication to the repository")
	flag.Parse()

	//Check all the flags has a value
	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "" {
			log.Fatalf("error: %s not defined", f.Name)
		}
	})

	//Check the repourl is a valid URL
	u, err := url.Parse(*repourl)
	if err != nil {
		log.Fatalln("repourl not valid")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Fatalln("repourl neither http nor https")
	}

	log.Println("uploading artifact...")
	us, err := deployfile.Upload(*repourl, *group, *artifact, *pkg, *version, *file, *username, *password, nil)
	if err != nil {
		log.Fatalf("error uploading: %v\n", err)
	}
	log.Println("uploaded files:")
	for _, u := range us {
		fmt.Fprintf(os.Stdout, "%s\n", u.String())
	}
	log.Println("upload complete.")
}
