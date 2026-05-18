# PRD: Medlemsregister MVP

## 1. Prosjektoversikt
Dette prosjektet er en MVP for et fleksibelt medlemsregister. Målet er ekstrem datakorrekthet og historisk sporbarhet. Arkitekturen er bygget for å senere støtte en dynamisk regelmotor, men MVP-en har kun hardkodede kjernefunksjoner.

## 2. Arkitekturmønstre (STRENGE REGLER)
- **Event Sourcing:** Tilstanden lagres som en uforanderlig logg av hendelser (`event_store`). Ingen rader slettes eller oppdateres her.
- **CQRS:** Vi skiller strengt mellom skriving og lesing.
  - **Skriv (Commands):** Går gjennom en deterministisk kjernelogikk (`internal/domain`) som validerer og produserer hendelser.
  - **Les (Queries):** Går rett mot en flat, lynrask database-tabell (`member_view`) som oppdateres asynkront av en projeksjons-worker.
- **Deterministic Simulation Testing (DST):** Kjernelogikken (`internal/domain`) SKAL være 100 % fri for I/O (ingen database, ingen nettverk, ingen `time.Now()`). All tid injiseres.
- **Ingen ORM:** Bruk standard `database/sql` (evt. med `pgx` eller `sqlx`) for full kontroll over SQL.

## 3. Teknologivalg
- **Backend:** Go (1.22+).
- **Database:** PostgreSQL (via Supabase på Dokploy).
- **Frontend:** Vue 3 (Composition API) + TailwindCSS.
- **Autentisering:** Supabase Auth (JWT).

## 4. Datamodell (PostgreSQL)

### Tabell: `event_store`
- `id` (BIGSERIAL, PK)
- `aggregate_id` (UUID, indeksert)
- `version` (INT)
- `event_type` (VARCHAR)
- `payload` (JSONB) - Inneholder data for hendelsen.
- `created_at` (TIMESTAMPTZ)

### Tabell: `member_view` (Lesemodell)
- `id` (UUID, PK)
- `team_id` (UUID, indeksert for Row-Level Security)
- `name` (VARCHAR)
- `email` (VARCHAR)
- `status` (VARCHAR)
- `metadata` (JSONB) - For alle fleksible/egendefinerte felter.
- `updated_at` (TIMESTAMPTZ)

## 5. Kjerne-hendelser (Events) for MVP
1. `MemberRegistered` { ID, Name, Email, TeamID, Metadata }
2. `MemberUpdated` { ID, UpdatedFields }

## 6. API-Kontrakter
- `POST /api/commands/register-member` -> Skriver til event_store.
- `GET /api/members` -> Henter fra member_view. Må støtte `team_id` filtrering basert på JWT-token.
- `GET /api/reports/as-of?date=YYYY-MM-DD` -> Tidsmaskin: Spoler event_store opp til gitt dato og returnerer in-memory tilstand.
