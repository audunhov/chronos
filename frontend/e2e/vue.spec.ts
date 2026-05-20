import { test, expect } from '@playwright/test'

test('visits the app root url', async ({ page }) => {
  await page.goto('/')
  // Vår app viser login-skjermen hvis man ikke er innlogget
  await expect(page.locator('h2')).toHaveText('Login')
})
