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

  test('skal kunne åpne den visuelle editoren', async ({ page }) => {
    await page.getByRole('button', { name: 'ADMIN' }).click();
    await page.getByRole('button', { name: 'PIPELINES' }).click();
    await page.getByRole('button', { name: 'ÅPNE VISUELL EDITOR' }).first().click();
    
    await expect(page.getByRole('heading', { name: 'Visual Pipeline Builder' })).toBeVisible({ timeout: 15000 });
    
    // Vent på Vue Flow
    await page.waitForSelector('.vue-flow', { timeout: 15000 });
    
    // Legg til en node
    await page.locator('button:has-text("IF / THEN")').click();
    
    // Vi sjekker bare at vi ikke får en hvit skjerm og at headeren er der
    await expect(page.getByRole('button', { name: 'PUBLISER ENDRINGER' })).toBeVisible();
  });
});
