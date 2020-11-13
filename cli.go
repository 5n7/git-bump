package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/Songmu/gitconfig"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// CLI represents a CLI.
type CLI struct {
	repo *git.Repository
}

func newCLI(dir string) (*CLI, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}
	return &CLI{repo: repo}, err
}

type spec = string

const (
	major spec = "major"
	minor spec = "minor"
	patch spec = "patch"
)

func (c *CLI) tags() ([]string, error) {
	tagRefs, err := c.repo.Tags()
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0)
	if err := tagRefs.ForEach(func(r *plumbing.Reference) error {
		tag := r.Name()
		if tag.IsTag() {
			tags = append(tags, tag.Short())
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return tags, nil
}

func (c *CLI) versions(tags []string) []*semver.Version {
	vs := make([]*semver.Version, 0)
	for _, tag := range tags {
		v, err := semver.NewVersion(tag)
		if err != nil {
			continue
		}
		vs = append(vs, v)
	}
	return vs
}

func (c *CLI) askNewVersion() (*semver.Version, error) {
	var v string
	prompt := &survey.Input{
		Message: "input new version:",
		Default: "v0.1.0",
	}
	if err := survey.AskOne(
		prompt, &v, survey.WithValidator(func(input interface{}) error {
			if s, ok := input.(string); ok {
				_, err := semver.NewVersion(s)
				return err
			}
			return fmt.Errorf("invalid input: %v", input)
		}),
	); err != nil {
		return nil, err
	}
	return semver.NewVersion(v) // prefix "v" is automatically removed
}

func (c *CLI) askNextVersion(current *semver.Version) (*semver.Version, error) {
	var (
		nextMajor = current.IncMajor()
		nextMinor = current.IncMinor()
		nextPatch = current.IncPatch()

		majorOption = fmt.Sprintf("%s: %s", major, nextMajor.Original())
		minorOption = fmt.Sprintf("%s: %s", minor, nextMinor.Original())
		patchOption = fmt.Sprintf("%s: %s", patch, nextPatch.Original())
	)

	var spec string
	prompt := &survey.Select{
		Message: "select the next version",
		Options: []string{patchOption, minorOption, majorOption},
		VimMode: true,
	}
	if err := survey.AskOne(prompt, &spec); err != nil {
		return nil, err
	}

	var next semver.Version
	switch spec {
	case majorOption:
		next = nextMajor
	case minorOption:
		next = nextMinor
	case patchOption:
		next = nextPatch
	default:
		return nil, fmt.Errorf("invalid semver")
	}
	return &next, nil
}

func (c *CLI) pushTag(version *semver.Version) error {
	head, err := c.repo.Head()
	if err != nil {
		return err
	}

	commit, err := c.repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	user, err := gitconfig.User()
	if err != nil {
		return err
	}

	email, err := gitconfig.Email()
	if err != nil {
		return err
	}

	tag := version.Original()
	if _, err = c.repo.CreateTag(tag, head.Hash(), &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  user,
			Email: email,
			When:  commit.Committer.When,
		},
		Message: commit.Message,
	}); err != nil {
		return err
	}

	rs := config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag, tag))
	if err := c.repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: user,
			Password: "",
		},
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{rs},
		Progress:   os.Stdout,
	}); err != nil {
		if err := c.repo.DeleteTag(tag); err != nil {
			return err
		}
		return fmt.Errorf("the created tag was automatically deleted: %w", err)
	}

	fmt.Printf("\nbump version to %s\n", color.New(color.FgMagenta).Sprint(tag))
	return nil
}

func (c *CLI) Run() error {
	tags, err := c.tags()
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		v, err := c.askNewVersion()
		if err != nil {
			return err
		}
		return c.pushTag(v)
	}

	vs := c.versions(tags)
	sort.Sort(semver.Collection(vs))

	color.Cyan("tags:")
	last := len(vs) - 1
	for i, v := range vs {
		if i < last {
			fmt.Printf("  - %s\n", v.Original())
		} else {
			color.Yellow("  - %s (current version)\n", v.Original())
		}
	}

	fmt.Print("\n")
	current := vs[last]
	next, err := c.askNextVersion(current)
	if err != nil {
		return err
	}

	if err := c.pushTag(next); err != nil {
		return err
	}
	return nil
}
