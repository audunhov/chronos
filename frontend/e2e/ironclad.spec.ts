import { test, expect } from '@playwright/test';

test.describe('Ironclad Business Flows', () => {

  test('Admin: Organization and Member Management', async ({ page }) => {
    await page.goto('/');
    await page.getByPlaceholder('DIN@EPOST.NO').fill('admin@chronos.no');
    await page.getByPlaceholder('********').fill('admin123');
    await page.getByRole('button', { name: 'LOGG INN' }).click();

    await expect(page.getByRole('heading', { name: 'CHRONOS' })).toBeVisible();

    // 1. Verify Structure
    await page.getByRole('button', { name: 'ADMIN' }).click();

    await page.getByRole('button', { name: 'STRUKTUR' }).click();
    await expect(page.getByText('Norges Spillforbund').first()).toBeVisible();

    // 2. Verify Stats
    await page.getByRole('button', { name: 'STATISTIKK' }).click();
    await expect(page.getByText('Statistikk & Vekst')).toBeVisible();
    
    // 3. Verify Treasury
    await page.getByRole('button', { name: 'FINANS' }).click();
    await expect(page.getByRole('button', { name: 'FINANS' })).toBeVisible();
  });
});
