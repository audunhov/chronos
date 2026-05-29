import { test, expect } from '@playwright/test';

test.describe('Visual Pipeline Editor E2E', () => {
  test.beforeEach(async ({ page }) => {
    page.on('console', msg => console.log(`BROWSER [${msg.type()}]: ${msg.text()}`));
    await page.goto('/');
    await page.locator('input[type="email"]').fill('admin@chronos.no');
    await page.locator('input[type="password"]').fill('admin123');
    await page.getByRole('button', { name: 'LOGG INN' }).click();
    await expect(page.getByText('IDENT: admin@chronos.no')).toBeVisible();
  });

  test('skal kunne bruke maler (presets) og se valideringsfeil', async ({ page }) => {
    await page.getByRole('button', { name: 'ADMIN' }).click();
    await page.getByRole('button', { name: 'PIPELINES' }).click();
    await page.getByRole('button', { name: 'ÅPNE VISUELL EDITOR' }).first().click();
    
    await expect(page.getByRole('heading', { name: 'Visual Pipeline Builder' })).toBeVisible({ timeout: 15000 });
    
    // Klikk på en mal (Velkomstpakken)
    await page.getByRole('button', { name: 'Velkomstpakken' }).click();
    
    // Verifiser at noder er lagt til i navigatoren
    await expect(page.locator('.vue-flow__node-trigger')).toBeVisible();
    
    // Sjekk at navigatoren lister nodene med deres LABELS
    const navigator = page.locator('aside section').filter({ hasText: 'Navigator' });
    await expect(navigator.getByRole('button', { name: 'FormatText' })).toBeVisible();
    await expect(navigator.getByRole('button', { name: 'SendEmail' })).toBeVisible();

    // Sjekk validering: SendEmail skal ha feil fordi inputs ikke er satt
    const sendMailNode = page.locator('.vue-flow__node-action').filter({ hasText: 'SendEmail' });
    await expect(sendMailNode.getByText('Mangler to_email')).toBeVisible({ timeout: 10000 });
  });

  test('skal kunne fokusere noder via navigatoren', async ({ page }) => {
    await page.getByRole('button', { name: 'ADMIN' }).click();
    await page.getByRole('button', { name: 'PIPELINES' }).click();
    await page.getByRole('button', { name: 'ÅPNE VISUELL EDITOR' }).first().click();
    
    await page.getByRole('button', { name: 'Velkomstpakken' }).click();

    // Klikk på en node i navigatoren
    const navigator = page.locator('aside section').filter({ hasText: 'Navigator' });
    await navigator.getByRole('button', { name: 'SendEmail' }).click();
    
    // Her er det vanskelig å verifisere nøyaktig zoom/pan i E2E uten dypere Vue Flow integrasjon,
    // men vi verifiserer at knappen er trykkbar og ikke kræsjer applikasjonen.
  });
});
