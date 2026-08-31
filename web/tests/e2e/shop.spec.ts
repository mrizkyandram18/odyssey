import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

test.describe('Reward Shop Journeys', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'demo1');
    await expect(page).toHaveURL(/.*#\/$/);
    const shopLink = page.locator('nav a:has-text("Toko Hadiah"), a[href="#/shop"]');
    await shopLink.click();
    await expect(page).toHaveURL(/.*#\/shop$/);
  });

  test('Journey 10 & 11: Member can view reward catalog and switch to claim history', async ({ page }) => {
    // Check shop header and coin balance
    await expect(page.locator('h1:has-text("Toko Hadiah")')).toBeVisible();
    await expect(page.locator('header span:has-text("Koin")')).toBeVisible();
    
    // Check catalog tab
    const catalogTab = page.locator('button:has-text("Katalog Hadiah")');
    await expect(catalogTab).toBeVisible();
    
    // Switch to history tab
    const historyTab = page.locator('button:has-text("Riwayat Penukaran")');
    await historyTab.click();
    await expect(page.locator('p:has-text("Belum Ada Riwayat Penukaran"), .claim-card').first()).toBeVisible();
    
    // Switch back to catalog tab
    await catalogTab.click();
    await expect(catalogTab).toBeVisible();
  });
});
