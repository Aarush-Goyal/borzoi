package lib

import (
	"fmt"
	"strings"
	"sync"

	"github.com/automation-co/borzoi/internal/config"
	"github.com/automation-co/borzoi/internal/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// =============================================================================

// Clones the repos in the given config file
func Clone(username string, accessToken string) {

	fmt.Println("Cloning the repositories...")
	fmt.Println("")

	// Read the config file
	conf := config.ReadConfig()

	// Get username
	usernameLocal := utils.GetUsername()
	if username == "" {
		username = usernameLocal
	}

	// Create waitgroup
	var wg sync.WaitGroup = sync.WaitGroup{}

	// Iterate over the repos in the config file
	for path, url := range conf {
		wg.Add(1)
		go func(url interface{}, path string) {
			// Get the url of the repo
			repoUrl := url.(string)

			fmt.Printf("  [x]  Cloning %s\n", repoUrl)

			// Clone the repo
			_, err := git.PlainClone(path, false, &git.CloneOptions{
				URL:               repoUrl,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
			if err != nil {

				if err.Error() == "authentication required" {
					// Check if repo is ssh or http
					isHttpUrl := strings.HasPrefix(repoUrl, "http")

					var auth transport.AuthMethod
					if isHttpUrl {
						auth = &http.BasicAuth{
							Username: username,
							Password: accessToken, // personal access token
							// needs to be created using github api
						}

					} else {

						// TODO: make public keys work

						publicKeys, err := ssh.NewPublicKeysFromFile("git", "privateKeyFile", "password")
						if err != nil {
							fmt.Printf("generate publickeys failed: %s\n", err.Error())
							return
						}

						auth = publicKeys

					}

					_, err := git.PlainClone(path, false, &git.CloneOptions{
						URL:               repoUrl,
						Auth:              auth,
						RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
					})
					if err != nil {
						if err.Error() == "repository already exists" {
							fmt.Println("  [o]  Skipping " + path + " because it already exists")
						} else {
							panic(err)
						}
					}
				} else if err.Error() == "repository already exists" {
					fmt.Println("  [o]  Skipping " + path + " because it already exists")
				} else {
					panic(err)
				}
			}
			wg.Done()

		}(url, path)

	}
	wg.Wait()

	fmt.Println("")
	fmt.Println("Woof 👍")
}

// =============================================================================
