package core

import (
	"fmt"
	"os"
	"path"
	"sort"
	"time"
)

type WorktreeInfo struct {
	Component string
	Branch    string
	Path      string
}

func (ws *Workspace) ListWorktrees(componentFilter string) ([]WorktreeInfo, error) {
	root := WorktreesRoot(ws)
	if !Pc.FileExists(root) {
		return nil, nil
	}

	entries, err := Pc.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var result []WorktreeInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		compName := entry.Name()
		if componentFilter != "" && compName != componentFilter {
			continue
		}
		if _, err := ws.ComponentByName(compName); err != nil {
			continue
		}
		compRoot := path.Join(root, compName)
		found, err := collectWorktrees(compRoot, compName, "")
		if err != nil {
			return nil, err
		}
		result = append(result, found...)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Component != result[j].Component {
			return result[i].Component < result[j].Component
		}
		return result[i].Branch < result[j].Branch
	})

	return result, nil
}

func collectWorktrees(dir, component, relBranch string) ([]WorktreeInfo, error) {
	if Pc.FileExists(path.Join(dir, ".git")) {
		if relBranch == "" {
			return nil, nil
		}
		return []WorktreeInfo{{
			Component: component,
			Branch:    relBranch,
			Path:      dir,
		}}, nil
	}

	entries, err := Pc.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []WorktreeInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		nextRel := name
		if relBranch != "" {
			nextRel = relBranch + "/" + name
		}
		found, err := collectWorktrees(path.Join(dir, name), component, nextRel)
		if err != nil {
			return nil, err
		}
		result = append(result, found...)
	}
	return result, nil
}

func (info WorktreeInfo) String() string {
	return fmt.Sprintf("%s\t%s\t%s", info.Component, info.Branch, info.Path)
}

// DirInfo is a minimal directory FileInfo used by tests.
type DirInfoName struct {
	NameVal string
}

func (d DirInfoName) Name() string       { return d.NameVal }
func (d DirInfoName) Size() int64        { return 0 }
func (d DirInfoName) Mode() os.FileMode  { return os.ModeDir | 0755 }
func (d DirInfoName) ModTime() time.Time { return time.Time{} }
func (d DirInfoName) IsDir() bool        { return true }
func (d DirInfoName) Sys() interface{}   { return nil }

func DirInfo(name string) os.FileInfo {
	return DirInfoName{NameVal: name}
}
