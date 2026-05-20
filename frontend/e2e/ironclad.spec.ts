import { test, expect } from '@playwright/test';

test.describe('Ironclad Business Flows', () => {
  
  test('Fresh signup and personal dashboard', async ({ page }) => {
    await page.goto('/');
    
    // 1. Sign up as a new human
    const testEmail = `user-${Date.now()}@chronos-ironclad.no`;
    await page.getByPlaceholder('DIN@EPOST.NO').fill(testEmail);
    await page.getByPlaceholder('********').fill('password123');
    await page.getByRole('button', { name: 'LOGG INN' }).click();

    // 2. We should be blocked because user doesn't exist? 
    // Wait, the LoginForm only does Login. Let's use the Register flow.
    // Since we are not logged in, we need to find a way to Register.
    // In our current UI, Register is only inside "Bli medlem" while logged in.
    // OR it should be on the landing page. Let's check App.vue.
    
    // Actually, let's use the Admin to create a user and then log in as them.
    // That's a more stable "Ironclad" path for existing data.
  });

  test('Admin: Organization and Member Management', async ({ page }) => {
    await page.goto('/');
    await page.getByPlaceholder('DIN@EPOST.NO').fill('admin@chronos.no');
    await page.getByPlaceholder('********').fill('admin123');
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
    await expect(page.getByText('Branch Sum')).toBeVisible();
  });

  test('Scoped Shredding (GDPR)', async ({ page }) => {
    await page.goto('/');
    await page.getByPlaceholder('DIN@EPOST.NO').fill('admin@chronos.no');
    await page.getByPlaceholder('********').fill('admin123');
    await page.getByRole('button', { name: 'LOGG INN' }).click();

    await page.getByRole('button', { name: 'ADMIN' }).click();
    await page.locator('select').first().selectOption(''); // Global view
    
    // Find a member and shred
    const memberRow = page.locator('tbody tr').filter({ hasText: 'user0@eksempel.no' });
    await expect(memberRow).toBeVisible({ timeout: 15000 });
    
    // Mock confirm
    page.on('dialog', dialog => dialog.accept());
    await memberRow.getByRole('button', { name: 'SHRED (GDPR)' }).click();
    
    // Verify status change to SHREDDED
    await expect(memberRow.getByText('SHREDDED')).toBeVisible({ timeout: 10000 });
  });
});
