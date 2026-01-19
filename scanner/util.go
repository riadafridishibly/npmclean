package scanner

import "time"

func GetDirectorySize(path string) (int64, error) {
	return getDirSize(path)
}

func GetLastModifiedAt(path string) (time.Time, error) {
	return getLastModTime(path)
}
