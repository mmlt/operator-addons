// Get content from remote GIT repository.
package repogit

// Package repogit provides a simple wrapper around git cli.
// Environment $HOME is expected to have .ssh/ directory to authenticate against remote repo.

import (
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/mmlt/operator-addons/internal/exe"
	"hash/fnv"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Repo represents the GIT repository to clone/pull from.
type Repo struct {
	// name of repo.
	name string
	// url of repo.
	url string
	// branch to use.
	branch string
	// tempDir is the file path were the repository is cloned.
	tempDir string

	// Log is the repo specific logger.
	log logr.Logger
}

// New creates an environment to Get data from a remote GIT repo.
func New(url, branch, token string, log logr.Logger) (*Repo, error) {
	name := path.Base(url)
	r := Repo{
		name: name,
		// As token is part of the url a change of token results in a new environment.
		// This way we don't have update and existing environment when a token changes,
		// Drawback is that a long running Pod accumulates repo clones that won't be used anymore.
		url:    urlWithToken(url, token),
		branch: branch,
		log:    log.WithName("Repo").WithValues("repo", name),
	}

	// Create directory to clone the repo into.
	p := filepath.Join(os.TempDir(), Hashed(url, branch))
	err := os.MkdirAll(p, 0755)
	if err != nil {
		return nil, err
	}
	r.tempDir = p
	r.log.V(2).Info("Create dir", "path", p)

	return &r, nil
}

// SHAremote returns the SHA of the last commit to the remote repo.
func (r *Repo) SHAremote() (string, error) {
	//TODO return cached value if called within 1 minute
	o, _, err := exe.Run("git", exe.Args{"ls-remote", r.url, "refs/heads/" + r.branch}, exe.Opt{}, r.log)
	if err != nil {
		return "", err
	}

	// parse result
	//  Warning: Permanently added the RSA host key for IP address '11.22.33.44' to the list of known hosts.
	//  a3a053fb28df45e33db1b634c1a45cb76e3d8bdf	refs/heads/master
	ss := strings.Fields(o)
	n := len(ss)
	if n < 2 && len(ss[n-2]) < 30 {
		return "", fmt.Errorf("sha of at least 30 chars expected, got: %s", o)
	}

	return ss[n-2], nil
}

// SHAlocal returns the SHA of the last commit to the local repo.
func (r *Repo) SHAlocal() (string, error) {
	o, _, err := exe.Run("git", exe.Args{"rev-parse", "refs/heads/" + r.branch}, r.optRepoDir(), r.log)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(o, "\n\r"), nil
}

// Get (clone or pull) the contents of the remote repo.
func (r *Repo) Get() error {
	_, err := os.Stat(filepath.Join(r.tempDir, r.name, ".git"))
	if os.IsNotExist(err) {
		// repo not cloned yet
		_, _, err = exe.Run("git", exe.Args{"clone", r.url}, r.optTempDir(), r.log)
		if err != nil {
			return err
		}

		_, _, err = exe.Run("git", exe.Args{"checkout", r.branch}, r.optRepoDir(), r.log)
		if err != nil {
			return err
		}
	} else {
		// repo already cloned
		_, _, err = exe.Run("git", exe.Args{"pull", "origin", r.branch}, r.optRepoDir(), r.log)
		if err != nil {
			return err
		}
	}

	sha, err := r.SHAlocal()
	if err != nil {
		return err
	}
	r.log.V(1).Info("Clone/pull", "commit", sha)

	return nil
}

// Remove removes the temporary directory that contains the cloned repository.
func (r *Repo) Remove() error {
	r.log.V(2).Info("Remove dir", "path", r.tempDir)
	return os.RemoveAll(r.tempDir)
}

// FQName is the fully qualified name of the repo.
// It includes the repo URL and branch.
// The URL is partially hashed so the result is usable as element of a path.
func (r *Repo) FQName() string {
	return Hashed(r.url, r.branch)
}

// Dir returns the absolute path to the repo root.
func (r *Repo) Dir() string {
	return filepath.Join(r.tempDir, r.name)
}

// Update updates the git repo to the latest commit in the branch.
func (r *Repo) Update() error {
	same, err := r.sameSHA()
	if err != nil {
		return err
	}
	if same {
		// Already up-to-date
		return nil
	}
	return r.Get()
}

// SameSHA returns true when the SHA's of the local and remote GIT repo are the same.
func (r *Repo) sameSHA() (bool, error) {
	lsha, err := r.SHAlocal()
	if err != nil {
		var e *os.PathError
		if errors.As(err, &e) {
			// any PathError is interpreted as; git directory does not exists
			return false, nil
		}
		return false, err
	}
	rsha, err := r.SHAremote()
	if err != nil {
		return false, err
	}

	return lsha == rsha, nil
}

// URLToken merges an optional token into url that starts with 'https://'
func urlWithToken(url, token string) string {
	const prefix = "https://"
	if !strings.HasPrefix(url, prefix) {
		// no action needed
		return url
	}
	return prefix + token + "@" + url[len(prefix):]
}

// Hashed returns a short version of url/branch in alphanum chars only.
func Hashed(url, branch string) string {
	// url part is max 24 chars incl. 8 chars hash.
	const max = 24 - 8

	h := fnv.New32a()
	h.Write([]byte(url))

	b := path.Base(url)
	l := len(b)
	if l > max {
		l = max
	}
	return fmt.Sprintf("%x-%s-%s", h.Sum32(), b[len(b)-l:], branch)
}

// OptTempDir returns options to exe commands in the temporary directory.
func (r *Repo) optTempDir() exe.Opt {
	return exe.Opt{Dir: r.tempDir}
}

// OptRepoDir returns options to exe commands in the repository root directory.
func (r *Repo) optRepoDir() exe.Opt {
	return exe.Opt{Dir: r.Dir()}
}
