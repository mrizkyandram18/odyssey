import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

test.describe('Member Task Stepper Journeys', () => {
  test.beforeEach(async ({ page }) => {
    await login(page, 'demo1');
    await expect(page).toHaveURL(/.*#\/$/);
  });

  test('Journey 3: Member loads daily task stepper and stats header', async ({ page }) => {
    // Assert daily title banner
    await expect(page.locator('h2:has-text("Petualangan Harian")')).toBeVisible();
    
    // Assert stats header (level and coins)
    await expect(page.locator('header').getByText('Lv.')).toBeVisible();
    
    // Check if task nodes or empty state is rendered
    const taskContent = page.locator('h4:has-text("Belum Ada Tugas"), .step-node, div:has-text("Puncak Petualangan")').first();
    await expect(taskContent).toBeVisible({ timeout: 10000 });
  });

  test('Journey 4 & 5/6: Member can interact with task node and inspect modal', async ({ page }) => {
    // If interactive step nodes are present, click the active one
    const activeStep = page.locator('button[data-status="ACTIVE"], button:has-text("Mulai"), div[role="button"]').first();
    if (await activeStep.isVisible()) {
      await activeStep.click();
      // Modal should appear
      const modal = page.locator('.fixed.inset-0, [role="dialog"]').first();
      await expect(modal).toBeVisible();
      
      // Close button should dismiss modal
      const closeBtn = page.locator('button:has-text("Batal"), button[title="Tutup"], button:has-text("Tutup")').first();
      if (await closeBtn.isVisible()) {
        await closeBtn.click();
      }
    } else {
      // Stepper correctly displayed empty or completed state
      await expect(page.locator('h2:has-text("Petualangan Harian")')).toBeVisible();
    }
  });
});
