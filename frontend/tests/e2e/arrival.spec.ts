import { expect, test } from '@playwright/test';

test('warms the page under the opening screen and opens the library', async ({
  page,
}, testInfo) => {
  let release!: () => void;
  const pending = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route('**/api/stats', async (route) => {
    await pending;
    await route.continue();
  });
  await page.goto('/');
  await expect(page.getByLabel('Preparing your space')).toBeVisible();
  await expect(page.locator('.arrival-content')).toHaveAttribute('inert', '');
  await expect(page.locator('#welcome-title')).toBeAttached();
  await page.screenshot({ path: `test-results/opening-${testInfo.project.name}.png` });
  release();
  await expect(page.getByLabel('Preparing your space')).toBeHidden();
  await expect(page.getByRole('heading', { name: 'Come in. Stay for a chapter.' })).toBeVisible();
  await page.screenshot({
    path: `test-results/welcome-${testInfo.project.name}.png`,
    animations: 'disabled',
    fullPage: true,
  });
  await page.getByRole('link', { name: 'Open your library' }).click();
  await expect(page.getByRole('heading', { name: 'Your library.' })).toBeVisible();
  await expect(page.getByLabel('Preparing your space')).toBeHidden();
});

test('mobile welcome supports reduced motion and keyboard navigation', async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.goto('/');
  await expect(page.getByLabel('Preparing your space')).toBeHidden();
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
  expect(
    await page.locator('.reading-still-life').evaluate((el) => getComputedStyle(el).animationName),
  ).toBe('none');
  await page.screenshot({
    path: `test-results/welcome-mobile-${testInfo.project.name}.png`,
    fullPage: true,
  });
  await page.getByRole('link', { name: 'Open your library' }).focus();
  await page.keyboard.press('Enter');
  await expect(page.getByRole('heading', { name: 'Your library.' })).toBeVisible();
});

test('startup failure offers recovery without trapping the reader', async ({ page }) => {
  await page.route('**/api/stats', (route) =>
    route.fulfill({
      status: 503,
      body: '{"error":"Unavailable"}',
      contentType: 'application/json',
    }),
  );
  await page.goto('/');
  await expect(page.getByRole('button', { name: 'Try again' })).toBeVisible();
  await page.unroute('**/api/stats');
  await page.getByRole('button', { name: 'Try again' }).click();
  await expect(page.getByLabel('Preparing your space')).toBeHidden();
  await expect(page.getByRole('link', { name: 'Open your library' })).toBeVisible();
});
