# Import Routes Setup Instructions

## Changes needed in `/api/cmd/server/main.go`

### 1. Add Import Repository and Service (around line 150, after `politicalPartyHandler`)

```go
// Import service for Excel functionality
importRepo := repository.NewImportRepository(db)
importService := services.NewImportService(importRepo, politicianRepo, politicalPartyRepo, locationRepo)
importHandler := handlers.NewImportHandler(importService)
```

### 2. Register Import Routes (around line 438, in admin routes section)

Add this block after the politicians routes (after `r.Post("/politicians/{id}/restore", politicianHandler.Restore)`):

```go
// Import endpoints for Excel file imports
r.Route("/import", func(r chi.Router) {
	r.Post("/politicians/validate", importHandler.ValidatePoliticianImport)
	r.Post("/politicians", importHandler.ImportPoliticians)
	r.Get("/politicians/logs", importHandler.ListImportLogs)
	r.Get("/politicians/logs/{id}", importHandler.GetImportLog)
	r.Get("/politicians/template", importHandler.DownloadTemplate)
	r.Get("/politicians/logs/{id}/errors", importHandler.DownloadErrorReport)
})
```

### 3. Run Database Migration

Before starting the server, run:

```bash
cd api
migrate -path migrations -database "postgres://politics:localdev@localhost:5432/politics_db?sslmode=disable" up
```

Or with Docker:

```bash
docker exec pulpulitiko-api /usr/local/bin/migrate -path /app/migrations -database "postgres://politics:localdev@postgres:5432/politics_db?sslmode=disable" up
```

## Verification

After making these changes and running migrations:

1. Start the API server
2. Check that these endpoints are available:
   - `POST /api/admin/import/politicians/validate` - Validate Excel file
   - `POST /api/admin/import/politicians` - Import politicians
   - `GET /api/admin/import/politicians/logs` - List import logs
   - `GET /api/admin/import/politicians/logs/:id` - Get import log
   - `GET /api/admin/import/politicians/template` - Download template
   - `GET /api/admin/import/politicians/logs/:id/errors` - Download errors

## Next Steps

After backend is set up, create the frontend import UI at:
- `/web/app/pages/admin/import/politicians.vue`
