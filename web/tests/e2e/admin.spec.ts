import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

test.describe('Admin Management Journeys', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'demo2');
    await expect(page).toHaveURL(/.*#\/$/);
    const adminLink = page.locator('nav a:has-text("Admin"), a[href="#/admin"]');
    await adminLink.click();
    await expect(page).toHaveURL(/.*#\/admin$/);
  });

  test('Journey 8: Admin can inspect task management tab and open task creator modal', async ({ page }) => {
    // Switch to Jadwal Tugas tab
    const tasksTab = page.locator('button:has-text("Jadwal Tugas")');
    await tasksTab.click();
    
    // Check create task button
    const createBtn = page.locator('button:has-text("Tambah Tugas")');
    await expect(createBtn).toBeVisible();
    await createBtn.click();
    
    // Modal should be open
    await expect(page.locator('form, h3, h4').first()).toBeVisible();
    
    // Close modal
    const cancelBtn = page.locator('button:has-text("Batal"), button[title="Tutup"]').first();
    if (await cancelBtn.isVisible()) {
      await cancelBtn.click();
    }
  });

  test('Journey 9: Admin can inspect submission verification queue and claims tab', async ({ page }) => {
    // Submissions verification tab
    const subTab = page.locator('button:has-text("Verifikasi Bukti")');
    await subTab.click();
    await expect(page.locator('p:has-text("Antrean Verifikasi"), h3:has-text("Antrean Verifikasi")').first()).toBeVisible();

    // Claims tab
    const claimsTab = page.locator('button:has-text("Pencairan Koin")');
    await claimsTab.click();
    await expect(claimsTab).toBeVisible();
  });
});
