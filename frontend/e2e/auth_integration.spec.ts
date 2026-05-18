import { test, expect } from '@playwright/test'

test.describe('E2E Autentisering', () => {
  const randomId = Math.random().toString(36).substring(7)
  const testEmail = `test-${randomId}@example.com`
  const testPassword = 'Password123!'
  const testOrg = `org-${randomId}`

  test('skal kunne registrere seg, logge inn og se dashbord', async ({ page }) => {
    // Lytt på console logs
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()))

    // 1. Gå til siden
    await page.goto('/')

    // 2. Registrering
    await page.getByPlaceholder('E-post').fill(testEmail)
    await page.getByPlaceholder('Passord').fill(testPassword)
    await page.getByPlaceholder('Organisasjons-ID (valgfritt)').fill(testOrg)
    
    // Sett opp promise for å vente på alert
    const dialogPromise = page.waitForEvent('dialog')
    
    await page.getByRole('button', { name: 'Registrer' }).click()

    // Vent på at alerten dukker opp og aksepter den
    const dialog = await dialogPromise
    console.log('Dialog:', dialog.message())
    await dialog.accept()

    // 3. Logg inn
    await page.getByPlaceholder('E-post').fill(testEmail)
    await page.getByPlaceholder('Passord').fill(testPassword)
    
    console.log('Klikker Logg inn for:', testEmail)
    await page.getByRole('button', { name: 'Logg inn' }).click()

    // 4. Verifiser dashbord
    try {
      await expect(page.getByText('Medlemsregister')).toBeVisible({ timeout: 15000 })
      await expect(page.getByText(`Innlogget som: ${testEmail}`)).toBeVisible({ timeout: 15000 })
    } catch (e) {
      console.log('E2E Debug: Dashboard ikke synlig. Innhold på siden:')
      console.log(await page.content())
      await page.screenshot({ path: 'test-failure.png' })
      throw e
    }

    // 5. Registrer et nytt medlem for å opprette organisasjonen i projeksjonen
    await page.getByRole('button', { name: 'Nytt medlem' }).click()
    await page.locator('#name').fill('E2E Test Medlem')
    await page.locator('#email').fill(`member-${randomId}@example.com`)
    await page.locator('#birth_year').fill('1990')
    // Bruker vår egen org
    await page.locator('#org_id').fill(testOrg)
    
    // Vent på at API-kallet blir ferdig
    const registerPromise = page.waitForResponse(resp => resp.url().includes('/commands/register-member') && resp.status() === 201)
    await page.getByRole('button', { name: 'Registrer medlem' }).click()
    await registerPromise
    
    // NÅ som vi har synkron projeksjon, bør medlemmet dukke opp med en gang ved neste fetch
    await page.getByTitle('Oppdater liste').click()
    await expect(page.getByText('E2E Test Medlem').first()).toBeVisible()
    
    // Sjekk at org-dropdown inneholder vår nye org
    const orgDropdown = page.locator('#org-select')
    await expect(orgDropdown).toContainText(testOrg)
  })
})
