# Playwright Integrasjonstesting

Vi bruker Playwright for å teste frontend-funksjonalitet isolert fra backend ved å mocke API-responser.

## Best Practices
- **Mocking**: Bruk `page.route` for å simulere alle API-endepunkter.
- **Data-attributes**: Foretrekk `getByRole` eller `getByText` for å finne elementer, som en ekte bruker ville gjort.
- **Isolasjon**: Hver test bør sette opp sin egen tilstand (f.eks. via `localStorage`).

## Eksempel: Mocking av API
```typescript
test('skal vise liste', async ({ page }) => {
  await page.route('**/api/members', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([{ ID: '1', Name: 'Test' }]),
    });
  });

  await page.goto('/');
  await expect(page.getByText('Test')).toBeVisible();
});
```

## Eksempel: Simulering av autentisering
```typescript
test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    window.localStorage.setItem('supabase_token', 'fake-token');
  });
});
```
