import { test, expect } from '@playwright/test';

test.describe('Medlemsregister Integrasjon', () => {
  test.beforeEach(async ({ page }) => {
    // Mock Supabase session for å hoppe over innlogging
    await page.route('**/auth/v1/session', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          access_token: 'fake-token',
          user: { email: 'test@example.com' }
        }),
      });
    });

    // Mock API-kall for medlemmer
    await page.route('**/api/members', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          {
            ID: 'uuid-1',
            Name: 'Ola Nordmann',
            Email: 'ola@example.com',
            Status: 'ACTIVE',
            Balance: 0,
            Metadata: {}
          },
          {
            ID: 'uuid-2',
            Name: 'Kari Nordmann',
            Email: 'kari@example.com',
            Status: 'ACTIVE',
            Balance: 50000,
            Metadata: {}
          }
        ]),
      });
    });
  });

  test('skal vise medlemslisten ved oppstart (innlogget)', async ({ page }) => {
    // Vi må simulere at brukeren er innlogget i localStorage for supabase-js
    await page.addInitScript(() => {
      const session = {
        access_token: 'fake-token',
        user: { email: 'test@example.com' }
      };
      window.localStorage.setItem('sb-localhost-auth-token', JSON.stringify(session));
    });

    await page.goto('/');

    await expect(page.locator('h1')).toContainText('Medlemsregister');
    await expect(page.getByText('Ola Nordmann')).toBeVisible();
    await expect(page.getByText('500.00 kr')).toBeVisible(); // Kari sin saldo
  });

  test('skal kunne registrere et nytt medlem', async ({ page }) => {
    await page.addInitScript(() => {
      const session = { access_token: 'fake-token', user: { email: 'test@example.com' } };
      window.localStorage.setItem('sb-localhost-auth-token', JSON.stringify(session));
    });
    
    await page.goto('/');

    await page.route('**/api/commands/register-member', async (route) => {
      await route.fulfill({ status: 201, body: JSON.stringify({ id: 'uuid-3' }) });
    });

    await page.getByRole('button', { name: 'Nytt medlem' }).click();
    
    await page.locator('#name').fill('Berit Bø');
    await page.locator('#email').fill('berit@example.com');
    await page.locator('#birth_year').fill('1995');
    
    await page.getByRole('button', { name: 'Registrer medlem' }).click();

    // Verifiser at API-kall ble gjort (kan gjøres mer eksplisitt, men her sjekker vi UI-respons)
    await expect(page.getByText('Berit Bø')).toBeVisible();
  });

  test('skal kunne utføre crypto-shredding', async ({ page }) => {
    await page.addInitScript(() => {
      const session = { access_token: 'fake-token', user: { email: 'test@example.com' } };
      window.localStorage.setItem('sb-localhost-auth-token', JSON.stringify(session));
    });
    
    await page.goto('/');

    await page.route('**/api/commands/shred-member', async (route) => {
      await route.fulfill({ status: 200 });
    });

    // Mock oppdatert liste etter shredding
    await page.route('**/api/members', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { ID: 'uuid-1', Name: 'REDACTED', Email: 'redacted@example.com', Status: 'SHREDDED', Balance: 0 }
        ]),
      });
    });

    page.on('dialog', dialog => dialog.accept());
    await page.getByText('Glem (GDPR)').first().click();

    await expect(page.getByText('REDACTED')).toBeVisible();
    await expect(page.getByText('SHREDDED')).toBeVisible();
  });
});
