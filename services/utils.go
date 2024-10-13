package services

import "strings"

type AsFiles []SharedFile

func (files AsFiles) Len() int      { return len(files) }
func (files AsFiles) Swap(i, j int) { files[i], files[j] = files[j], files[i] }
func (files AsFiles) Less(i, j int) bool {
	a := files[i]
	b := files[j]

	// Place directories first
	if a.FType != b.FType {
		return a.FType > b.FType
	}

	// Place dotfiles first
	if strings.HasPrefix(a.Name, ".") && !strings.HasPrefix(b.Name, ".") {
		return true
	}

	// Default to sorting by name
	return a.Name < b.Name
}
