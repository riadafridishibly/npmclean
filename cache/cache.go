package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type CacheEntry struct {
	Path           string
	Size           int64
	LastModifiedAt time.Time
	ScannedAt      time.Time
}

type Cache struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS node_modules (
    path TEXT PRIMARY KEY,
    size INTEGER NOT NULL,
    last_modified_at INTEGER NOT NULL,
    scanned_at INTEGER NOT NULL
);
`

func NewCache() (*Cache, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	dbPath := filepath.Join(cacheDir, "npmclean.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	db.Exec(`PRAGMA journal_mode=WAL;`)
	db.Exec(`PRAGMA synchronous=NORMAL;`)
	db.Exec(`PRAGMA busy_timeout=5000;`)
	db.Exec(`PRAGMA temp_store=MEMORY;`)
	db.Exec(`PRAGMA mmap_size=30000000000;`)

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Cache{db: db}, nil
}

func (c *Cache) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func getCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "npmclean"), nil
}

func (c *Cache) InsertOrUpdate(entry *CacheEntry) error {
	query := `
        INSERT INTO node_modules (path, size, last_modified_at, scanned_at)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(path) DO UPDATE SET
            size = excluded.size,
            last_modified_at = excluded.last_modified_at,
            scanned_at = excluded.scanned_at
    `
	_, err := c.db.Exec(query, entry.Path, entry.Size, entry.LastModifiedAt.Unix(), entry.ScannedAt.Unix())
	return err
}

func (c *Cache) GetAll() ([]*CacheEntry, error) {
	rows, err := c.db.Query("SELECT path, size, last_modified_at, scanned_at FROM node_modules")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*CacheEntry
	for rows.Next() {
		var path string
		var size int64
		var lastModUnix, scannedUnix int64
		if err := rows.Scan(&path, &size, &lastModUnix, &scannedUnix); err != nil {
			return nil, err
		}
		entry := &CacheEntry{
			Path:           path,
			Size:           size,
			LastModifiedAt: time.Unix(lastModUnix, 0),
			ScannedAt:      time.Unix(scannedUnix, 0),
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (c *Cache) Delete(path string) error {
	_, err := c.db.Exec("DELETE FROM node_modules WHERE path = ?", path)
	return err
}

func (c *Cache) Get(path string) (*CacheEntry, error) {
	var size int64
	var lastModUnix, scannedUnix int64
	err := c.db.QueryRow("SELECT size, last_modified_at, scanned_at FROM node_modules WHERE path = ?", path).Scan(&size, &lastModUnix, &scannedUnix)
	if err != nil {
		return nil, err
	}
	return &CacheEntry{
		Path:           path,
		Size:           size,
		LastModifiedAt: time.Unix(lastModUnix, 0),
		ScannedAt:      time.Unix(scannedUnix, 0),
	}, nil
}
