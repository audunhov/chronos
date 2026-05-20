import { test, expect } from '@playwright/test';

test.describe('Ironclad Business Flows', () => {

  test('Admin: Organization and Member Management', async ({ page }) => {
    await page.goto('/');
    await page.getByPlaceholder('DIN@EPOST.NO').fill('audun@su.no');
    await page.getByPlaceholder('********').fill('password123');
    await page.getByRole('button', { name: 'LOGG INN' }).click();

    // 1. Verify Structure
    await page.getByRole('button', { name: 'ADMIN' }).click();
    await page.getByRole('button', { name: 'STRUKTUR' }).click();
    await expect(page.getByText('Norges Spillforbund')).toBeVisible();

    // 2. Verify Stats
    await page.getByRole('button', { name: 'STATISTIKK' }).click();
    await expect(page.getByText('Statistikk & Vekst')).toBeVisible();
    
    // 3. Verify Treasury
    await page.getByRole('button', { name: 'FINANS' }).click();
    await expect(page.getByRole('button', { name: 'FINANS' })).toBeVisible();
  });
});
