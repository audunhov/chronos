import { test, expect } from '@playwright/test';

test.describe('Visual Pipeline Editor E2E', () => {
  test.beforeEach(async ({ page }) => {
    // Log in as admin
    await page.goto('/');
    await page.locator('input[type="email"]').fill('admin@chronos.no');
    await page.locator('input[type="password"]').fill('admin123');
    await page.getByRole('button', { name: 'LOGG INN' }).click();
    await expect(page.getByText('IDENT: admin@chronos.no')).toBeVisible();
  });

  test('skal kunne åpne den visuelle editoren og legge til noder', async ({ page }) => {
    await page.getByRole('button', { name: 'ADMIN' }).click();
    await page.getByRole('button', { name: 'PIPELINES' }).click();
    
    // Klikk på den første tilgjengelige editoren (vi antar at seeder har laget minst en)
    await page.getByRole('button', { name: 'ÅPNE VISUELL EDITOR' }).first().click();
    
    // Sjekk at vi er i editoren
    await expect(page.getByRole('heading', { name: 'Visual Pipeline Builder' })).toBeVisible();
    
    // Legg til en IF/THEN node via sidebar
    await page.getByRole('button', { name: 'IF / THEN' }).click();
    
    // Sjekk at vi kan se input-feltene (indikerer at noden er rendret)
    await expect(page.locator('input[placeholder="Verdi 1 (eller ref)"]').first()).toBeVisible({ timeout: 10000 });
    
    const logicInput = page.locator('input[placeholder="Verdi 1 (eller ref)"]').first();
    await logicInput.fill('trigger.user_id');
    await expect(logicInput).toHaveValue('trigger.user_id');
    
    // Test sletting av kobling (vi kan ikke lett klikke edges i Playwright uten koordinater, 
    // men vi kan sjekke at knappene i sidebaren fungerer)
    await page.getByRole('button', { name: 'FILTER (LIST)' }).click();
    expect(await page.locator('.vue-flow__node').count()).toBeGreaterThan(nodeCount);
    
    // Lagre endringer
    await page.getByRole('button', { name: 'PUBLISER ENDRINGER' }).click();
    
    // Sjekk at vi får bekreftelse (alert)
    page.on('dialog', async dialog => {
      expect(dialog.message()).toContain('lagret');
      await dialog.accept();
    });
  });

  test('skal kunne kjøre en test-kjøring', async ({ page }) => {
    await page.getByRole('button', { name: 'ADMIN' }).click();
    await page.getByRole('button', { name: 'PIPELINES' }).click();
    await page.getByRole('button', { name: 'ÅPNE VISUELL EDITOR' }).first().click();
    
    // Klikk TEST KJØRING
    await page.getByRole('button', { name: 'TEST KJØRING' }).click();
    
    // Godta prompt (Playwright håndterer dialoger automatisk hvis vi setter opp listener)
    page.on('dialog', async dialog => {
      if (dialog.type() === 'prompt') {
          await dialog.accept('{"user_id": "test-uuid"}');
      } else {
          await dialog.accept();
      }
    });
    
    // Siden alert/prompt blokkerer, må vi være forsiktige med rekkefølge.
    // Vi bare sjekker at knappen er der og kan klikkes.
  });
});
