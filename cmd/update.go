/*
Copyright © 2023 David Mann

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/codingninja/gitops-repo-api/diff"
	"github.com/codingninja/gitops-repo-api/entrypoint"
	"github.com/codingninja/gitops-repo-api/git"
	"github.com/codingninja/gitops-repo-api/resource"
	"github.com/go-git/go-git/v5/plumbing"
	r3diff "github.com/r3labs/diff/v3"
	"github.com/spf13/cobra"
)

// updateCmd represents the test command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update objects matching a selector",
	Long:  `Test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		debug := false
		if len(args) != 3 {
			return fmt.Errorf("invalid arguments, expected 3, got %+v", args)
		}
		repo := args[0]
		from := args[1]
		to := args[2]
		ctx := context.Background()

		rs := git.NewRepoSpec(repo, nil)

		if debug {
			rs.Progress = os.Stdout
		}

		preRef := plumbing.NewBranchReferenceName(to)
		postRef := plumbing.NewBranchReferenceName(from)
		epds := []entrypoint.EntrypointFactory{
			entrypoint.EntrypointDiscoverySpec{
				Type: entrypoint.EntrypointTypeKustomize,
				// Regex: *regexp.MustCompile(`/(?P<name>[^/]+)/overlays/(?P<overlay>[^/]+)/`),
				Regex: *regexp.MustCompile(`/k8-workshop/overlays/(?P<overlay>[^/]+)`),
				Context: map[string]interface{}{
					"name": "k8-workshop",
				},
			},
		}
		differ := diff.NewDiffer(rs, rs, epds)
		diff, err := differ.Diff(ctx, preRef, postRef)
		if err != nil {
			fmt.Printf("Got errors diffing resources:\n\n%s\n", err.Error())
		}

		for _, ep := range diff {
			fmt.Printf("Entrypoint %q was changed:\n", ep.Entrypoint.Directory)
			for _, res := range ep.Diff {
				fmt.Printf("Detected changes in resource %s\n", res.String())
				if res.Type == resource.DiffTypeCreate {
					fmt.Printf("	Resource %q was created\n", res.Name())
				} else {
					for _, change := range res.Diff {
						fmt.Printf("	Field %s ", strings.Join(change.Path, "."))
						if change.Type == r3diff.UPDATE {
							fmt.Printf("was updated from %q to %q\n", change.From, change.To)
						} else if change.Type == r3diff.CREATE {
							fmt.Printf("was created with an initial value of %q\n", change.To)
						} else if change.Type == r3diff.DELETE {
							fmt.Printf("was deleted, previously it's value was %q\n", change.From)
						} else {
							fmt.Print("Unknown change type!!\n")
						}
					}
				}
				fmt.Printf("\n")
			}

			fmt.Print("\n\n")
		}

		// encoded, err := json.MarshalIndent(diff, "", "  ")
		// if err != nil {
		// 	return err
		// }
		// fmt.Print(string(encoded))

		return nil

		// fmt.Printf("Cloned to %s\n", preDir)
		// w, err := repo.Worktree()
		// if err != nil {
		// 	return fmt.Errorf("error fetching work tree for %q - %w", rs.URL, err)
		// }

		// _, err = w.Commit("Test Message", &v5git.CommitOptions{
		// 	AllowEmptyCommits: true,
		// })

		// if err != nil {
		// 	return fmt.Errorf("unable to commit - %w", err)
		// }

		// err = repo.PushContext(ctx, &v5git.PushOptions{
		// 	RemoteURL: rs.URL,
		// 	Atomic:    true,
		// 	Auth:      rs.Credentials,
		// 	RefSpecs: []config.RefSpec{
		// 		config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName)),
		// 	},
		// })

		// if err != nil {
		// 	return fmt.Errorf("unable to push - %w", err)
		// }

		// return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
