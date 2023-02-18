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
	"errors"
	"fmt"
	"os"
	"regexp"
	"sync"

	"git.dmann.xyz/davidmann/gitops-repo-api/entrypoint"
	"git.dmann.xyz/davidmann/gitops-repo-api/git"
	"git.dmann.xyz/davidmann/gitops-repo-api/resource"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// setAnnotationFn
var setAnnotationFn = kio.FilterFunc(func(operand []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range operand {
		resource := operand[i]
		_, err := resource.Pipe(yaml.SetAnnotation("foo", "bar"))
		if err != nil {
			return nil, err
		}
	}
	return operand, nil
})

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test code",
	Long:  `Test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cleanup := true
		ctx := context.Background()
		// homedir, err := os.UserHomeDir()
		// if err != nil {
		// 	return fmt.Errorf("unable to deduce home dir - %w", err)
		// }
		// key, err := ssh.NewPublicKeysFromFile("git", path.Join(homedir, ".ssh/id_rsa"), "")

		// if err != nil {
		// 	return fmt.Errorf("unable to get ssh keys from %q - %w", homedir, err)
		// }

		rs := git.NewRepoSpec("https://git.dmann.dev/davidmann/manifests.git", nil)
		branchName := "main"

		_, preDir, err := rs.Checkout(ctx, plumbing.NewHashReference("main", plumbing.NewHash("3d0a8971857bcbe3124a73253ac235fa3eb95072")))
		if err != nil {
			return fmt.Errorf("unable to pre change dir - %w", err)
		}

		_, postDir, err := rs.Checkout(ctx, plumbing.NewSymbolicReference(plumbing.NewBranchReferenceName(branchName), plumbing.NewBranchReferenceName(branchName)))
		if err != nil {
			return fmt.Errorf("unable to checkout post change dir - %w", err)
		}
		defer func() {
			if !cleanup {
				fmt.Printf("\n\n=========================\n\nNot cleaning up\n%s\n%s\n\n=========================\n\n", preDir, postDir)
				return
			}

			var errs error
			if err := os.RemoveAll(preDir); err != nil {
				errs = errors.Join(errs, err)
			}
			if err := os.RemoveAll(postDir); err != nil {
				errs = errors.Join(errs, err)
			}
			if errs != nil {
				panic(errs.Error())
			}
		}()

		eps, err := entrypoint.DiscoverEntrypoints(preDir, []entrypoint.EntrypointDiscoverySpec{
			{
				Type:  "kustomization",
				Regex: *regexp.MustCompile(`/(?P<name>[^/]+)/overlays/(?P<overlay>[^/]+)/kustomization.yaml`),
			},
		})

		if err != nil {
			return err
		}

		var errs error
		wg := sync.WaitGroup{}
		for _, ep := range eps {
			if ep.Name != "k8-workshop" {
				continue

			}
			ep := ep
			wg.Add(1)
			go func() {
				defer wg.Done()
				ewg := sync.WaitGroup{}
				ewg.Add(1)
				var preResources *resmap.ResMap
				var postResources *resmap.ResMap
				var buildErrs error
				go func() {
					defer ewg.Done()
					pr, err := entrypoint.ExtractResources(preDir, ep)
					if err != nil {
						buildErrs = errors.Join(buildErrs, fmt.Errorf("unable to build pre-entrypoint - %w", err))
						return
					}
					preResources = &pr

				}()

				ewg.Add(1)
				go func() {
					defer ewg.Done()
					pr, err := entrypoint.ExtractResources(postDir, ep)
					if err != nil {
						buildErrs = errors.Join(buildErrs, fmt.Errorf("unable to build pre-entrypoint - %w", err))
						return
					}
					postResources = &pr
				}()
				ewg.Wait()
				if buildErrs != nil {
					errs = errors.Join(errs, buildErrs)
					return
				}
				if preResources == nil || postResources == nil {
					errs = errors.Join(errs, fmt.Errorf("unknown error fetching pre/post resources"))
					return
				}

				diff, err := resource.Diff(*preResources, *postResources)
				if err != nil {
					errs = errors.Join(errs, err)
					return
				}

				if len(diff) > 0 {
					fmt.Printf("Got the diff: \n\n")
					spew.Dump(diff)
					fmt.Printf("\n\n\n\n")
				}
			}()
		}
		wg.Wait()

		return errs

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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
