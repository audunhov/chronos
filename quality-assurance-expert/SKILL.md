---
name: quality-assurance-expert
description: Sikrer høy kodekvalitet gjennom Deterministic Simulation Testing (DST) for backend og integrasjonstesting for frontend. Bruk denne skill-en når du legger til nye funksjoner eller endrer eksisterende logikk for å sikre at testdekningen er i tråd med prosjektets standarder.
---

# Quality Assurance Expert

Dette er prosjektets standard for kvalitetssikring. Vi prioriterer determinisme i backend og pålitelige integrasjonstester i frontend.

## Kjerne-prinsipper

1. **Backend: Determinisme fremfor alt**
   - All domene-logikk (`internal/domain`) SKAL være fri for I/O.
   - Bruk DST (Deterministic Simulation Testing) for å verifisere kompleks logikk.
   - Se [references/dst.md](references/dst.md) for mønstre og implementasjonsdetaljer.

2. **Frontend: Verifisering av brukerstier**
   - Nye komponenter og funksjoner som involverer brukerinteraksjon eller API-kommunikasjon SKAL integrasjonstestes.
   - Bruk Playwright for å simulere brukerflyt og mock API-responser.
   - Se [references/playwright.md](references/playwright.md) for mønstre.

## Arbeidsflyt for nye funksjoner

### 1. Planlegging
Når en ny funksjon foreslås, identifiser:
- Hvilken del av logikken som hører hjemme i domenet (DST-kandidat).
- Hvilke brukerstier som er kritiske å verifisere i frontend (Integrasjonstest-kandidat).

### 2. Implementering
- Skriv domene-logikken først.
- Implementer tilhørende DST i `*_test.go`.
- Implementer frontend-komponenter og API-tjenester.
- Skriv integrasjonstester i `frontend/e2e/`.

### 3. Validering
- Kjør `go test ./internal/domain/...`
- Kjør `cd frontend && npm run test:e2e` (eller tilsvarende kommando)
