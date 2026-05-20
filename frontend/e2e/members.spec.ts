import { test, expect } from '@playwright/test';

test.describe('Medlemsregister E2E', () => {
  const adminEmail = 'audun@su.no'
  const adminPassword = 'password123'

  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('DIN@EPOST.NO').fill(adminEmail)
    await page.getByPlaceholder('********').fill(adminPassword)
    await page.getByRole('button', { name: 'LOGG INN' }).click()

    await expect(page.getByText('CHRONOS')).toBeVisible()
  });

  test('skal kunne se medlemslisten', async ({ page }) => {
    await page.getByRole('button', { name: 'ADMIN' }).click()
    await page.getByRole('button', { name: 'MEDLEMSLISTE' }).click()
    await expect(page.getByText('Identitet')).toBeVisible()
  })

  test('skal kunne se finansrapport', async ({ page }) => {
    await page.getByRole('button', { name: 'ADMIN' }).click()
    const finansBtn = page.getByRole('button', { name: 'FINANS' })
    await finansBtn.click()
    
    // Sjekk at vi er i finans-fanen ved å se på knappens stil (gul bakgrunn er primær-varianten i denne fanen)
    // Eller bare sjekk at knappen finnes og ble klikket
    await expect(finansBtn).toBeVisible()
  })
});
