import { test, expect } from '@playwright/test';

test.describe('Medlemsregister E2E', () => {
  const workerId = Math.random().toString(36).substring(7)
  const testEmail = `admin-${workerId}@example.com`
  const testPassword = 'Password123!'
  const testOrg = `org-${workerId}`

  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('E-post').fill(testEmail)
    await page.getByPlaceholder('Passord').fill(testPassword)
    await page.getByPlaceholder('Organisasjons-ID (valgfritt)').fill(testOrg)
    
    const dialogPromise = page.waitForEvent('dialog')
    await page.getByRole('button', { name: 'Registrer' }).click()
    const dialog = await dialogPromise
    await dialog.accept()

    await page.getByPlaceholder('E-post').fill(testEmail)
    await page.getByPlaceholder('Passord').fill(testPassword)
    await page.getByRole('button', { name: 'Logg inn' }).click()

    await expect(page.getByText('Medlemsregister')).toBeVisible()
  });

  test('skal kunne registrere et nytt medlem og se det i listen', async ({ page }) => {
    await page.getByRole('button', { name: 'Nytt medlem' }).click();
    
    const memberName = `Berit Bø ${Math.random().toString(36).substring(7)}`
    const memberEmail = `berit-${workerId}@example.com`

    await page.locator('#name').fill(memberName);
    await page.locator('#email').fill(memberEmail);
    await page.locator('#birth_year').fill('1995');
    await page.locator('#org_id').fill(testOrg);
    
    const registerPromise = page.waitForResponse(resp => resp.url().includes('/commands/register-member') && resp.status() === 201)
    await page.getByRole('button', { name: 'Registrer medlem' }).click();
    await registerPromise

    await page.getByTitle('Oppdater liste').click()
    await expect(page.getByText(memberName).first()).toBeVisible({ timeout: 10000 });
  });

  test('skal kunne utføre crypto-shredding (GDPR glem)', async ({ page }) => {
    const deleteName = `Slette-meg ${Math.random().toString(36).substring(7)}`
    
    await page.getByRole('button', { name: 'Nytt medlem' }).click();
    await page.locator('#name').fill(deleteName);
    await page.locator('#email').fill(`delete-${workerId}@example.com`);
    await page.locator('#birth_year').fill('1980');
    await page.locator('#org_id').fill(testOrg);
    
    const registerPromise = page.waitForResponse(resp => resp.url().includes('/commands/register-member') && resp.status() === 201)
    await page.getByRole('button', { name: 'Registrer medlem' }).click();
    await registerPromise

    await page.getByTitle('Oppdater liste').click()
    await expect(page.getByText(deleteName).first()).toBeVisible();

    page.on('dialog', d => d.accept());
    await page.locator('tr').filter({ hasText: deleteName }).getByText('Glem (GDPR)').click();

    await page.getByTitle('Oppdater liste').click()
    await expect(page.getByText('REDACTED').first()).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('SHREDDED').first()).toBeVisible();
  });
});
