// Package store handles saving and reading scrape results from Firestore.
//
// Firestore structure:
//
//	scrape_runs/          (collection)
//	  {run_id}/           (document) — metadata for one scrape session
//	    items/            (subcollection)
//	      {item_id}       (document) — one scraped item
package store

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/yourusername/gcp-scraper-cron/internal/scraper"
	"google.golang.org/api/option"
)

const (
	collectionRuns  = "scrape_runs"
	collectionItems = "items"
)

// Run represents one scrape session stored in Firestore.
type Run struct {
	ID        string    `firestore:"id"`
	StartedAt time.Time `firestore:"started_at"`
	ItemCount int       `firestore:"item_count"`
	Status    string    `firestore:"status"` // "success" | "error"
	Error     string    `firestore:"error,omitempty"`
}

// Client wraps the Firestore client.
type Client struct {
	fs        *firestore.Client
	projectID string
}

// New creates a Firestore client.
// projectID is read from the GCP_PROJECT_ID env var by the caller.
// For local development, set GOOGLE_APPLICATION_CREDENTIALS to your service account key path.
func New(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	fs, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("firestore.NewClient: %w", err)
	}
	return &Client{fs: fs, projectID: projectID}, nil
}

// Close closes the Firestore connection.
func (c *Client) Close() error {
	return c.fs.Close()
}

// SaveRun saves a complete scrape session — the run metadata and all items — to Firestore.
// Returns the run ID so it can be referenced later.
func (c *Client) SaveRun(ctx context.Context, items []scraper.Item) (string, error) {
	runID := fmt.Sprintf("run_%d", time.Now().UnixMilli())

	run := Run{
		ID:        runID,
		StartedAt: time.Now(),
		ItemCount: len(items),
		Status:    "success",
	}

	// Write the run document
	runRef := c.fs.Collection(collectionRuns).Doc(runID)
	if _, err := runRef.Set(ctx, run); err != nil {
		return "", fmt.Errorf("save run: %w", err)
	}

	// Write each item into the subcollection in a batch
	batch := c.fs.Batch()
	for i, item := range items {
		itemRef := runRef.Collection(collectionItems).Doc(fmt.Sprintf("item_%03d", i))
		batch.Set(itemRef, map[string]any{
			"title":      item.Title,
			"value":      item.Value,
			"source":     item.Source,
			"scraped_at": item.ScrapedAt,
		})
	}
	if _, err := batch.Commit(ctx); err != nil {
		return runID, fmt.Errorf("save items: %w", err)
	}

	return runID, nil
}

// SaveFailedRun records a run that ended in error.
func (c *Client) SaveFailedRun(ctx context.Context, err error) error {
	runID := fmt.Sprintf("run_%d", time.Now().UnixMilli())
	run := Run{
		ID:        runID,
		StartedAt: time.Now(),
		ItemCount: 0,
		Status:    "error",
		Error:     err.Error(),
	}
	_, writeErr := c.fs.Collection(collectionRuns).Doc(runID).Set(ctx, run)
	return writeErr
}

// GetRecentRuns fetches the N most recent scrape runs (metadata only, no items).
func (c *Client) GetRecentRuns(ctx context.Context, limit int) ([]Run, error) {
	docs, err := c.fs.Collection(collectionRuns).
		OrderBy("started_at", firestore.Desc).
		Limit(limit).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, fmt.Errorf("get recent runs: %w", err)
	}

	runs := make([]Run, 0, len(docs))
	for _, doc := range docs {
		var run Run
		if err := doc.DataTo(&run); err != nil {
			continue
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// GetRunItems fetches all items from a specific run.
func (c *Client) GetRunItems(ctx context.Context, runID string) ([]scraper.Item, error) {
	docs, err := c.fs.Collection(collectionRuns).Doc(runID).
		Collection(collectionItems).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, fmt.Errorf("get run items: %w", err)
	}

	items := make([]scraper.Item, 0, len(docs))
	for _, doc := range docs {
		var item scraper.Item
		if err := doc.DataTo(&item); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}
