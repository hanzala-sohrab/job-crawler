# Job Crawler Platform

Resume-driven job aggregator — upload your resume, get personalized job listings from multiple sources.

## Tech Stack

- **Backend**: Go (Chi, JWT, Gemini AI, Colly scrapers)
- **Frontend**: Next.js 15 (App Router, TypeScript)
- **Database**: PostgreSQL 16, Redis 7
- **Infrastructure**: Docker Compose

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.23+ (for local development)
- Node.js 20+ (for local development)
- Gemini API key (optional, for AI resume parsing)

### With Docker Compose
```bash
docker compose up -d
```

### Local Development

1. **Start databases:**
```bash
docker compose up postgres redis -d
```

2. **Backend:**
```bash
cd backend
cp .env.example .env  # Edit with your settings
go run ./cmd/server
```

3. **Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | No | Register |
| POST | `/api/v1/auth/login` | No | Login |
| GET | `/api/v1/auth/me` | Yes | Profile |
| POST | `/api/v1/resume/upload` | Yes | Upload resume |
| GET | `/api/v1/resume` | Yes | Get parsed resume |
| PUT | `/api/v1/resume` | Yes | Update resume data |
| GET | `/api/v1/jobs` | Yes | List matched jobs |
| GET | `/api/v1/jobs/:id` | Yes | Job detail |
| PUT | `/api/v1/jobs/:id/status` | Yes | Update status |
| GET | `/api/v1/jobs/sources` | Yes | Active sources |
| POST | `/api/v1/jobs/refresh` | Yes | Re-match jobs |

## Environment Variables

See `backend/.env` for all configuration options.
