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

	"git.dmann.xyz/davidmann/gitops-repo-api/diff"
	"git.dmann.xyz/davidmann/gitops-repo-api/git"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test code",
	Long:  `Test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		rs := git.NewRepoSpec("https://git.dmann.dev/davidmann/manifests.git", nil)

		branchName := "main"
		preRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), plumbing.NewHash("3d0a8971857bcbe3124a73253ac235fa3eb95072"))
		postRef := plumbing.NewSymbolicReference(preRef.Name(), preRef.Name())

		diff, err := diff.Diff(ctx, rs, preRef, postRef)
		if err != nil {
			return err
		}

		spew.Dump(diff)

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
