# Implementasjonsplan for Gemini CLI

**INSTRUKS TIL AI:** Implementer dette systemet sekvensielt. IKKE gå videre til neste fase før forrige fase er fullført og testet. Bruk filstrukturen definert under.

## Fase 1: Domain & DST (Ren logikk, ingen I/O)
**Mål:** Etabler kjernelogikken og sikkerhetsnettet.
1. Opprett `internal/domain/member.go`. Definer `Member`-struct og `Event`-interface.
2. Skriv funksjonen `ApplyEvent(m *Member, e Event)` som muterer `Member` i minnet basert på hendelsestypen.
3. Opprett `internal/domain/member_test.go`. Skriv en "Deterministic Simulation Test" (DST) som genererer 1000 tilfeldige `MemberRegistered` og `MemberUpdated` hendelser, kjører dem gjennom `ApplyEvent`, og verifiserer tilstanden. 
*Krav: Ingen database eller net-pakker importert i denne fasen.*

## Fase 2: Database og Projeksjon (I/O Skallet)
**Mål:** Lagre hendelser og bygge lesemodellen.
1. Opprett `internal/storage/db.go` for Postgres-tilkobling.
2. Opprett `internal/storage/event_store.go`. Implementer funksjon for å lagre (`Append`) hendelser til `event_store`-tabellen. Håndter versjonskonflikter (Optimistic Concurrency).
3. Opprett `internal/projection/worker.go`. Skriv en bakgrunnsjobb som poller eller lytter på nye rader i `event_store`, og utfører `INSERT`/`UPDATE` på `member_view`-tabellen tilsvarende.

## Fase 3: API og Autentisering
**Mål:** Eksponer systemet over HTTP.
1. Opprett `internal/api/middleware.go`. Implementer en middleware som tar et JWT-token fra `Authorization`-headeren (Supabase-format), verifiserer signaturen, og legger `UserID` og `TeamID` i request-konteksten.
2. Opprett `internal/api/handlers.go`.
   - Implementer `GET /api/members` (les fra `member_view` filtrert på `TeamID` fra kontekst).
   - Implementer `POST /api/commands/register-member` (valider request, lag Domain Event, send til `event_store`).

## Fase 4: Tidsmaskinen (Rapportering)
**Mål:** Implementer historisk spoling.
1. Opprett en funksjon i `internal/reports/time_machine.go` kalt `GetMembersAsOf(db, targetDate)`.
2. Logikken skal:
   - Kjøre SQL: `SELECT aggregate_id, event_type, payload FROM event_store WHERE created_at <= targetDate ORDER BY id ASC`.
   - Iterere over radene, dytte dem inn i `domain.ApplyEvent()` for å bygge opp en map av medlemmer i minnet.
   - Returnere den ferdige tilstanden.
3. Eksponer dette som `GET /api/reports/as-of`.

## Fase 5: Vue Frontend (MVP)
**Mål:** Enkel visning av data.
1. Initier et Vue 3 + Vite + Tailwind prosjekt i mappen `/frontend`.
2. Lag en tjeneste (`api.js`) for å snakke med Go-backend. Sørg for at den automatisk legger ved Supabase JWT-token.
3. Lag én hovedkomponent: En tabell som viser medlemslisten (hentet fra `/api/members`).
4. Legg til en datovelger over tabellen. Når en dato velges, gjør et kall til `/api/reports/as-of` og oppdater tabellen.
5. Lag en enkel "Nytt medlem"-modal som poster til `/api/commands/register-member`.
