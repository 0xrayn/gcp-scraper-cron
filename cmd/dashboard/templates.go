// cmd/dashboard — templates.go holds the inline HTML templates for the dashboard.
// Kept separate from main.go so the request handlers stay readable.
package main

const listHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Scraper dashboard</title>
  <style>
    body { font-family: -apple-system, sans-serif; max-width: 720px; margin: 40px auto; padding: 0 20px; color: #1a1a1a; }
    h1 { font-size: 20px; font-weight: 500; margin-bottom: 4px; }
    p.sub { color: #6b6b6b; font-size: 14px; margin-bottom: 24px; }
    table { width: 100%; border-collapse: collapse; }
    th { text-align: left; font-size: 12px; text-transform: uppercase; color: #888; padding: 8px 0; border-bottom: 1px solid #e5e5e5; }
    td { padding: 12px 0; border-bottom: 1px solid #f0f0f0; font-size: 14px; }
    a { color: #185fa5; text-decoration: none; }
    a:hover { text-decoration: underline; }
    .badge { display: inline-block; padding: 2px 8px; border-radius: 99px; font-size: 12px; }
    .ok { background: #eaf3de; color: #3b6d11; }
    .err { background: #fcebeb; color: #a32d2d; }
    .empty { color: #888; padding: 40px 0; text-align: center; }
  </style>
</head>
<body>
  <h1>Scraper run history</h1>
  <p class="sub">Most recent 20 runs</p>

  {{if .}}
  <table>
    <tr><th>Run ID</th><th>Started</th><th>Status</th><th>Items</th></tr>
    {{range .}}
    <tr>
      <td><a href="/runs/{{.ID}}">{{.ID}}</a></td>
      <td>{{.StartedAt.Format "02 Jan 2006, 15:04"}}</td>
      <td>
        {{if eq .Status "success"}}<span class="badge ok">success</span>{{else}}<span class="badge err">error</span>{{end}}
      </td>
      <td>{{.ItemCount}}</td>
    </tr>
    {{end}}
  </table>
  {{else}}
  <div class="empty">No runs yet. Trigger the scraper at least once.</div>
  {{end}}
</body>
</html>`

const detailHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Run {{.RunID}}</title>
  <style>
    body { font-family: -apple-system, sans-serif; max-width: 720px; margin: 40px auto; padding: 0 20px; color: #1a1a1a; }
    h1 { font-size: 20px; font-weight: 500; margin-bottom: 4px; }
    a.back { color: #185fa5; text-decoration: none; font-size: 14px; }
    table { width: 100%; border-collapse: collapse; margin-top: 20px; }
    th { text-align: left; font-size: 12px; text-transform: uppercase; color: #888; padding: 8px 0; border-bottom: 1px solid #e5e5e5; }
    td { padding: 10px 0; border-bottom: 1px solid #f0f0f0; font-size: 14px; }
    .source { color: #888; font-size: 12px; }
  </style>
</head>
<body>
  <a class="back" href="/">&larr; Back to all runs</a>
  <h1>Run {{.RunID}}</h1>

  <table>
    <tr><th>Title</th><th>Value</th><th>Source</th></tr>
    {{range .Items}}
    <tr>
      <td>{{.Title}}</td>
      <td>{{.Value}}</td>
      <td class="source">{{.Source}}</td>
    </tr>
    {{end}}
  </table>
</body>
</html>`
