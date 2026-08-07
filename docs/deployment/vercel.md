# Vercel Deployment Guide

Odyssey is designed to be easily deployable to Vercel as a full-stack Next.js/Vite application utilizing Vercel's Serverless Functions for the Go backend.

## Prerequisites
1. A Supabase project with the schema fully migrated (`scripts/migrations/*`).
2. A Vercel account linked to your repository.

## Environment Variables

Configure the following Environment Variables in your Vercel Project Settings:

| Variable | Description |
|---|---|
| `SUPABASE_URL` | Your Supabase Project URL |
| `SUPABASE_SERVICE_KEY` | Your Supabase Service Role Key (needed to bypass RLS for the backend) |
| `SESSION_SIGNING_SECRET` | A secure random string used to sign JWT session cookies |
| `ADMIN_SECRET` | A secure random string for admin API access |
| `ODYSSEY_TIMEZONE` | e.g. `Asia/Jakarta` |

*(Note: `FIREBASE_CREDENTIALS` and `PARENT_ID` are no longer required for the Prototype phase since we use Local Auth).*

## Build Configuration

Vercel automatically detects the Go backend in the `/api` directory and the Vite frontend in the `/web` directory.

In Vercel Project Settings:
- **Framework Preset:** Vite
- **Root Directory:** `web`
- **Build Command:** `npm run build`
- **Output Directory:** `dist`

### Go Serverless Functions
Vercel automatically compiles any `.go` file in the `api` folder that exports a `Handler` function into a Serverless Function.

To ensure all dependencies are resolved, the `api/dev/main.go` wires up the domain logic, but for Vercel, the individual `api/<route>/index.go` files must be independently wired. 

*(For this prototype, if you run into `503 Service Unavailable` on Vercel, ensure you include an `init()` block in `api/bootstrap` or the specific handler to wire up the `LocalAuthProvider` and `SupabaseClient` on cold start).*
