package glice

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type licenseFormat struct {
	name  string
	color color.Attribute
}

var licenseCol = map[string]licenseFormat{
	"other":      {name: "Other", color: color.FgBlue},
	"mit":        {name: "MIT", color: color.FgGreen},
	"lgpl-3.0":   {name: "LGPL-3.0", color: color.FgCyan},
	"mpl-2.0":    {name: "MPL-2.0", color: color.FgHiBlue},
	"agpl-3.0":   {name: "AGPL-3.0", color: color.FgHiCyan},
	"unlicense":  {name: "Unlicense", color: color.FgHiRed},
	"apache-2.0": {name: "Apache-2.0", color: color.FgHiGreen},
	"gpl-3.0":    {name: "GPL-3.0", color: color.FgHiMagenta},
}

// Repository holds information about the repository
type Repository struct {
	Name      string
	Shortname string
	URL       string
	Host      string
	Author    string
	Project   string
	Text      string
}

func newGitClient(c context.Context, keys map[string]string, star bool) *gitClient {
	var tc *http.Client
	var ghLogged bool
	if v, _ := keys["github.com"]; v != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: v},
		)
		tc = oauth2.NewClient(c, ts)
		ghLogged = true
	}
	return &gitClient{
		gh: githubClient{
			Client: github.NewClient(tc),
			logged: ghLogged,
		},
		star: star,
	}
}

type gitClient struct {
	gh   githubClient
	star bool
}

type githubClient struct {
	*github.Client
	logged bool
}

// GetLicense for a repository
func (gc *gitClient) GetLicense(ctx context.Context, r *Repository) error {
	switch r.Host {
	case "github.com":
		rl, _, err := gc.gh.Repositories.License(ctx, r.Author, r.Project)
		if err != nil {
			fmt.Println(r.Author, r.Project)
			return err
		}

		name, clr := licenseCol[*rl.License.Key].name, licenseCol[*rl.License.Key].color
		if name == "" {
			name = *rl.License.Key
			clr = color.FgYellow
		}
		r.Shortname = color.New(clr).Sprintf(name)
		r.Text = rl.GetContent()

		if gc.star && gc.gh.logged {
			gc.gh.Activity.Star(ctx, r.Author, r.Project)
		}
	}

	return nil
}
