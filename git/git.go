package git

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/xerrors"
)

type Dir struct {
	dir string
}

func NewDir(dir string) (*Dir, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, xerrors.Errorf("could not compute absolute directory: %w", err)
	}
	targetDir, err := filepath.EvalSymlinks(absDir)
	if err != nil {
		return nil, xerrors.Errorf("could not resolve target for symlink %q: %w", absDir, err)
	}
	if _, err := os.Stat(targetDir); err != nil {
		return nil, xerrors.Errorf("could not stat directory %q: %w", targetDir, err)
	}
	return &Dir{dir: targetDir}, nil
}

func Init(dir string) (*Dir, error) {
	cmd := exec.Command("git", "init", dir)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf("could not init git repo (stderr: %s): %w", stderr.String(), err)
	}
	return &Dir{dir: dir}, nil
}

func (g *Dir) RootDir() string {
	return g.dir
}

func (g *Dir) Remotes() ([]string, error) {
	cmd := exec.Command("git", "-C", g.dir, "remote")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf("could not list remotes (stderr: %s): %w", stderr.String(), err)
	}
	return strings.Fields(stdout.String()), nil
}

func (g *Dir) RemoteURL(remote string) (string, error) {
	cmd := exec.Command("git", "-C", g.dir, "remote", "get-url", remote)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", xerrors.Errorf("could not get url for remote %q (stderr: %s): %w", remote, stderr.String(), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (g *Dir) ListFiles() ([]string, error) {
	stdout := &bytes.Buffer{}
	cmd := exec.Command("git", "-C", g.dir, "ls-files")
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf("could not list git files: %w", err)
	}
	return strings.Fields(string(stdout.Bytes())), nil
}

func (g *Dir) Stat(path string) (os.FileInfo, error) {
	file := filepath.Join(g.dir, path)
	return os.Stat(file)
}

func (g *Dir) Rename(oldpath, newpath string) error {
	cmd := exec.Command("git", "-C", g.dir, "mv", oldpath, newpath)
	if err := cmd.Run(); err != nil {
		return xerrors.Errorf("could not rename file %q to %q: %w", oldpath, newpath, err)
	}
	return nil
}

func (g *Dir) CreateFile(path string, data []byte, mode os.FileMode) (status error) {
	file := filepath.Join(g.dir, path)
	if _, err := os.Stat(file); err == nil {
		return xerrors.Errorf("target path %q already exists: %w", path, os.ErrExist)
	}
	if err := ioutil.WriteFile(file, data, mode); err != nil {
		return xerrors.Errorf("could not create file %q: %w", err)
	}
	defer func() {
		if status != nil {
			if err := g.Reset("HEAD"); err != nil {
				log.Panicf("could not undo updating file %q: %v", path, err)
				return
			}
		}
	}()
	addCmd := exec.Command("git", "-C", g.dir, "add", path)
	if err := addCmd.Run(); err != nil {
		return xerrors.Errorf("could not add new file %q: %w", path, err)
	}
	return nil
}

func (g *Dir) ReadFile(path string) ([]byte, error) {
	file := filepath.Join(g.dir, path)
	return ioutil.ReadFile(file)
}

func (g *Dir) RemoveFile(path string) error {
	removeCmd := exec.Command("git", "-C", g.dir, "rm", path)
	if err := removeCmd.Run(); err != nil {
		return xerrors.Errorf("could not remove file %q: %w", path, err)
	}
	return nil
}

func (g *Dir) WriteFile(path string, data []byte, mode os.FileMode) (status error) {
	file := filepath.Join(g.dir, path)
	if err := ioutil.WriteFile(file, data, mode); err != nil {
		return xerrors.Errorf("could not write to file %q: %w", path, err)
	}
	addCmd := exec.Command("git", "-C", g.dir, "add", file)
	if err := addCmd.Run(); err != nil {
		return xerrors.Errorf("could not add changes to file %q: %w", path, err)
	}
	return nil
}

func (g *Dir) UpdateFile(path string, data []byte) (status error) {
	file := filepath.Join(g.dir, path)
	stat, err := os.Stat(file)
	if err != nil {
		return xerrors.Errorf("could not stat file %q: %w", path, err)
	}
	if err := ioutil.WriteFile(file, data, stat.Mode()); err != nil {
		return xerrors.Errorf("could not write to file %q: %w", path, err)
	}
	defer func() {
		if status != nil {
			if err := g.Reset("HEAD"); err != nil {
				log.Panicf("could not undo updating file %q: %v", path, err)
				return
			}
		}
	}()
	addCmd := exec.Command("git", "-C", g.dir, "add", file)
	if err := addCmd.Run(); err != nil {
		return xerrors.Errorf("could not add changes to file %q: %w", path, err)
	}
	return nil
}

func (g *Dir) Commit(msg string) error {
	commitCmd := exec.Command("git", "-C", g.dir, "commit", "-m", msg)
	if err := commitCmd.Run(); err != nil {
		return xerrors.Errorf("could not commit changes to the git repository: %w", err)
	}
	return nil
}

func (g *Dir) Reset(target string) error {
	resetCmd := exec.Command("git", "-C", g.dir, "reset", "--hard", target)
	if err := resetCmd.Run(); err != nil {
		return xerrors.Errorf("could not reset changes to the git repository: %w", err)
	}
	return nil
}

func (g *Dir) Apply(msg string, cb func() error) (status error) {
	defer func() {
		if status != nil {
			if err := g.Reset("HEAD"); err != nil {
				log.Panicf("could not revert changes to the git repo: %v", err)
				return
			}
		}
	}()

	if err := cb(); err != nil {
		return err
	}

	if err := g.Commit(msg); err != nil {
		return xerrors.Errorf("could not commit changes to the git repo: %w", err)
	}
	return nil
}

func (g *Dir) AddRemote(name, url string) error {
	cmd := exec.Command("git", "-C", g.dir, "remote", "add", name, url)
	if err := cmd.Run(); err != nil {
		return xerrors.Errorf("could not add git remote %q: %w", url, err)
	}
	return nil
}

func (g *Dir) FetchAll() error {
	cmd := exec.Command("git", "-C", g.dir, "fetch", "--all")
	if err := cmd.Run(); err != nil {
		return xerrors.Errorf("could not fetch all remotes: %w", err)
	}
	return nil
}

func (g *Dir) SetConfg(key, value string) error {
	cmd := exec.Command("git", "-C", g.dir, "config", "--local", key, value)
	if err := cmd.Run(); err != nil {
		return xerrors.Errorf("could not git config key %q: %w", key, err)
	}
	return nil
}

func (g *Dir) GetConfig(key string) (string, error) {
	return "", xerrors.Errorf("TODO")
}
