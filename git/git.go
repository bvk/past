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

func (g *Dir) RootDir() string {
	return g.dir
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

func (g *Dir) AddFile(path string, data []byte, mode os.FileMode) (status error) {
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
