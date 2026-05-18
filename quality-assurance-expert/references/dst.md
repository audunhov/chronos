# Deterministic Simulation Testing (DST) i Go

DST er en teknikk der vi tester logikk ved å simulere en sekvens av hendelser med en fiksert random seed. Dette gjør testene 100% reproduserbare.

## Krav til domene-logikk
- **Ingen I/O**: Ingen database-kall, filsystem-aksess eller nettverk.
- **Ingen ekstern tid**: Bruk injisert tid (`time.Time`) i hendelsene, ikke `time.Now()`.
- **Ingen global tilstand**: Alt som trengs for å mutere tilstand må finnes i hendelsen eller aggregatet.

## DST Mønster
```go
func TestDST_FeatureX(t *testing.T) {
    rng := rand.New(rand.NewSource(42)) // Fiksert seed for determinisme
    m := &Member{}
    
    var events []Event
    currentTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
    
    for i := 0; i < 1000; i++ {
        // Generer tilfeldig hendelse basert på rng
        // ...
        currentTime = currentTime.Add(time.Duration(rng.Intn(3600)) * time.Second)
        // ...
    }
    
    // Kjør alle hendelser gjennom ApplyEvent
    for _, e := range events {
        ApplyEvent(m, e)
    }
    
    // Verifiser slutt-tilstand
    // ...
}
```
