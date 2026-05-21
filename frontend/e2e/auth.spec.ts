import { test, expect } from '@playwright/test'

// Nullstill storageState for autentiseringstester
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Autentisering E2E', () => {
  const adminEmail = 'admin@chronos.no'
  const adminPassword = 'admin123'

  test('skal kunne logge inn som admin', async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('DIN@EPOST.NO').fill(adminEmail)
    await page.getByPlaceholder('********').fill(adminPassword)
    await page.getByRole('button', { name: 'LOGG INN' }).click()

    await expect(page.getByRole('heading', { name: 'CHRONOS' })).toBeVisible()
    await expect(page.getByText(`IDENT: ${adminEmail}`)).toBeVisible()
  })

  test('skal vise feilmelding ved feil passord', async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('DIN@EPOST.NO').fill(adminEmail)
    await page.getByPlaceholder('********').fill('wrongpassword')
    await page.getByRole('button', { name: 'LOGG INN' }).click()

    await expect(page.getByText('Invalid credentials')).toBeVisible()
  })

  test('skal kunne bytte til registreringsskjema', async ({ page }) => {
    await page.goto('/')
    await page.getByRole('button', { name: 'INGEN KONTO? REGISTRER DEG HER' }).click()
    await expect(page.getByRole('heading', { name: 'REGISTRER' })).toBeVisible()
  })
})
