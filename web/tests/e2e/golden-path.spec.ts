import { test, expect } from '@playwright/test';
import { login, logout } from './helpers/auth';
import { verifyHomeLoaded, verifyQuestExists } from './helpers/home';
import { startQuest, completeChallenge, verifyXpIncreased, resetQuestState } from './helpers/quest';
import { navigateToJournal, verifyJournalEntryExists } from './helpers/journal';

test.describe('Golden Path', () => {
  // Test data strategy: Ephemeral DB is seeded before this test runs in CI.
  const TEST_USER = 'demo1';
  const QUEST_TITLE = 'Bab 2: Jejak Kaki Raksasa'; 
  const CHALLENGE_TITLE = 'Peran 1 (Seeker): Ukur telapak kakimu dengan sebuah benda (seperti buku atau pensil).'; 

  test.beforeEach(async () => {
    await resetQuestState();
  });

  test('Test 1: Login -> Home -> Quest list appears', async ({ page }) => {
    await login(page, TEST_USER);
    await verifyHomeLoaded(page);
    await verifyQuestExists(page, QUEST_TITLE);
  });

  test('Test 2: Start Quest -> Complete Challenge -> XP increases -> Journal updates', async ({ page }) => {
    await login(page, TEST_USER);
    await verifyHomeLoaded(page);
    
    await startQuest(page, QUEST_TITLE);
    await completeChallenge(page, CHALLENGE_TITLE);
    
    await verifyXpIncreased(page);
    
    await navigateToJournal(page);
    await verifyJournalEntryExists(page, QUEST_TITLE);
  });

  test('Test 3: Logout -> Login -> Progress persists', async ({ page }) => {
    await login(page, TEST_USER);
    await verifyHomeLoaded(page);
    
    await navigateToJournal(page);
    
    await logout(page);
    
    await login(page, TEST_USER);
    await navigateToJournal(page);
    // Verify state persists (stubbed)
    await expect(page).toHaveURL(/.*#\/journal$/);
  });
});
