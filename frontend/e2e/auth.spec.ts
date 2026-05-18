import { test, expect } from '@playwright/test'

// Nullstill storageState for autentiseringstester
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Autentisering E2E', () => {
  const randomId = Math.random().toString(36).substring(7)
  const testEmail = `auth-test-${randomId}@example.com`
  const testPassword = 'Password123!'

  test('skal kunne registrere en ny bruker og logge inn', async ({ page }) => {
    await page.goto('/')

    // 1. Registrer
    await page.getByPlaceholder('E-post').fill(testEmail)
    await page.getByPlaceholder('Passord').fill(testPassword)
    
    const dialogPromise = page.waitForEvent('dialog')
    await page.getByRole('button', { name: 'Registrer' }).click()
    const dialog = await dialogPromise
    await dialog.accept()

    // 2. Logg inn
    await page.getByPlaceholder('E-post').fill(testEmail)
    await page.getByPlaceholder('Passord').fill(testPassword)
    await page.getByRole('button', { name: 'Logg inn' }).click()

    // Bekreft dashbord
    await expect(page.getByText('Medlemsregister')).toBeVisible()
    await expect(page.getByText(`Innlogget som: ${testEmail}`)).toBeVisible()
  })

  test('skal vise feilmelding ved feil passord', async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('E-post').fill('nonexistent@example.com')
    await page.getByPlaceholder('Passord').fill('wrongpassword')
    await page.getByRole('button', { name: 'Logg inn' }).click()

    await expect(page.getByText('Invalid credentials')).toBeVisible()
  })
})
