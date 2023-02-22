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
	"encoding/json"
	"fmt"
	"regexp"

	"git.dmann.xyz/davidmann/gitops-repo-api/diff"
	"git.dmann.xyz/davidmann/gitops-repo-api/entrypoint"
	"git.dmann.xyz/davidmann/gitops-repo-api/git"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test code",
	Long:  `Test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("invalid arguments, expected 3, got %+v", args)
		}
		repo := args[0]
		from := args[1]
		to := args[2]
		ctx := context.Background()

		rs := git.NewRepoSpec(repo, nil)

		branchName := to
		preRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), plumbing.NewHash(from))
		postRef := plumbing.NewSymbolicReference(preRef.Name(), preRef.Name())
		epds := []entrypoint.EntrypointDiscoverySpec{
			{
				Type: "kustomization",
				// Regex: *regexp.MustCompile(`/(?P<name>[^/]+)/overlays/(?P<overlay>[^/]+)/kustomization.yaml`),
				Regex: *regexp.MustCompile(`/k8-workshop/overlays/(?P<overlay>[^/]+)/kustomization.yaml`),
				Context: map[string]string{
					"name": "k8-workshop",
				},
			},
		}
		differ := diff.NewDiffer(rs, epds)
		diff, err := differ.Diff(ctx, preRef, postRef)
		if err != nil {
			fmt.Printf("Got errors diffing resources:\n\n%s\n", err.Error())
		}

		encoded, err := json.MarshalIndent(diff, "", "  ")
		if err != nil {
			return err
		}
		fmt.Print(string(encoded))

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
	rootCmd.AddCommand(testCmd)
}
