package testutils

type Differ interface {
	Diff(dir1, dir2 string, ignores []string) (string, error)
}

func FolderDiff(d Differ, dir1, dir2 string, ignores []string) (string, error) {
	return d.Diff(dir1, dir2, ignores)
}
