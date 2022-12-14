package utils

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
)

func CloneRemote(url string) string {
	tmpdir, _ := ioutil.TempDir(os.TempDir(), "ark-remote")

	fmt.Println("\nCloning: " + url)
	fmt.Printf("Repo Dir: " + tmpdir + "\n\n")

	_, err := git.PlainClone(tmpdir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	fmt.Printf("Done.\n\n")

	if err != nil {
		panic(err)
	}

	return tmpdir
}
