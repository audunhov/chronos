import { test, expect } from '@playwright/test'

test.describe('Forms Management E2E', () => {
  const adminEmail = 'audun@su.no'
  const adminPassword = 'password123'

  test.beforeEach(async ({ page }) => {
    // Logg inn som admin
    await page.goto('/')
    await page.getByPlaceholder('DIN@EPOST.NO').fill(adminEmail)
    await page.getByPlaceholder('********').fill(adminPassword)
    await page.getByRole('button', { name: 'LOGG INN' }).click()
    await expect(page.getByText('CHRONOS')).toBeVisible()
  })

  test('skal kunne opprette et nytt skjema og svare på det', async ({ page }) => {
    const randomId = Math.random().toString(36).substring(7)
    const testFormTitle = `Test-Skjema-${randomId}`

    // 1. Gå til Admin -> Skjemaer
    await page.getByRole('button', { name: 'ADMIN' }).click()
    await expect(page.getByRole('button', { name: 'MEDLEMSLISTE' })).toBeVisible()
    
    await page.getByRole('button', { name: 'SKJEMAER' }).click()

    // 2. Opprett nytt skjema
    await page.getByRole('button', { name: '+ NYTT SKJEMA' }).click()
    await expect(page.getByText('Konfigurer nytt skjema')).toBeVisible()
    
    await page.getByPlaceholder('F.eks. Medlemsundersøkelse 2026').fill(testFormTitle)
    
    // Velg organisasjon (vent til opsjoner er lastet)
    const orgSelect = page.locator('select').first()
    await expect(orgSelect.locator('option').nth(1)).toBeAttached()
    await orgSelect.selectOption({ index: 1 }) 

    await page.getByRole('button', { name: '+ LEGG TIL' }).click()
    await page.getByPlaceholder('f_navn').fill('navn')
    await page.getByPlaceholder('Hva heter du?').fill('Navn')
    
    await page.getByRole('button', { name: 'LAGRE OG PUBLISER' }).click()
    
    // Vent på at skjemaet dukker opp i listen
    await expect(page.getByRole('heading', { name: testFormTitle })).toBeVisible()

    // 3. Svar på skjemaet (Mine sider -> Undersøkelser)
    await page.getByRole('button', { name: 'MINE SIDER' }).click()
    await page.getByRole('button', { name: 'UNDERSØKELSER' }).click()
    
    await page.getByRole('button', { name: 'SVAR PÅ SKJEMA' }).last().click()
    await page.locator('input[type="text"]').fill('Playwright Tester')
    await page.getByRole('button', { name: 'SEND INN SVAR' }).click()

    await expect(page.getByText('SUKSESS!')).toBeVisible()
  })
})
