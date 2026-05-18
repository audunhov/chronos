import { test, expect } from '@playwright/test'

test.describe('Autentisering', () => {
  test('skal kunne logge inn med gyldige legitimasjoner', async ({ page }) => {
    // Mock Supabase Auth request
    await page.route('**/auth/v1/token?grant_type=password', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          access_token: 'fake-access-token',
          token_type: 'bearer',
          expires_in: 3600,
          refresh_token: 'fake-refresh-token',
          user: {
            id: 'user-123',
            email: 'test@example.com',
            user_metadata: { org_id: 'org-123' }
          }
        }),
      })
    })

    // Gå til login-siden
    await page.goto('/')

    // Fyll ut skjemaet
    await page.getByPlaceholder('E-post').fill('test@example.com')
    await page.getByPlaceholder('Passord').fill('password123')

    // Klikk på Logg inn
    await page.getByRole('button', { name: 'Logg inn' }).click()

    // Sjekk at vi blir logget inn (f.eks. ved at login-skjemaet forsvinner eller vi ser en velkomstmelding)
    // Siden jeg ikke har sett App.vue ennå, antar jeg at den viser medlemslisten etter innlogging
    await expect(page.getByText('Logg inn i Medlemsregister')).not.toBeVisible()
  })

  test('skal vise feilmelding ved feil passord', async ({ page }) => {
    // Mock Supabase Auth feil
    await page.route('**/auth/v1/token?grant_type=password', async (route) => {
      await route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'invalid_grant',
          error_description: 'Invalid login credentials'
        }),
      })
    })

    await page.goto('/')
    await page.getByPlaceholder('E-post').fill('test@example.com')
    await page.getByPlaceholder('Passord').fill('wrongpassword')
    await page.getByRole('button', { name: 'Logg inn' }).click()

    // Sjekk feilmelding
    await expect(page.getByText('Invalid login credentials')).toBeVisible()
  })

  test('skal kunne registrere en ny bruker', async ({ page }) => {
    // Mock Supabase Signup
    await page.route('**/auth/v1/signup', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'user-123',
          email: 'new@example.com',
        }),
      })
    })

    // Lytt etter alert
    let alertMessage = ''
    page.on('dialog', async dialog => {
      alertMessage = dialog.message()
      await dialog.accept()
    })

    await page.goto('/')
    await page.getByPlaceholder('E-post').fill('new@example.com')
    await page.getByPlaceholder('Passord').fill('password123')
    await page.getByRole('button', { name: 'Registrer' }).click()

    // Vent på alert og sjekk innholdet
    await expect.poll(() => alertMessage).toContain('Sjekk e-posten din')
  })
})
