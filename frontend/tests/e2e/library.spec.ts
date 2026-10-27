import { expect, test } from '@playwright/test';
test('add, find, update, and retain a title across reloads', async ({ page }, testInfo) => {
  await page.goto('/library');
  await expect(page.getByRole('heading', { name: 'Your library.' })).toBeVisible();
  await page.getByRole('button', { name: 'Add a title', exact: true }).click();
  const title = `The patient reader ${testInfo.project.name}`;
  await page.getByLabel('Title', { exact: true }).fill(title);
  await page.getByLabel('Author or creator').fill('A. Reader');
  await page.getByLabel('Total (0 if unknown)').fill('10');
  await page.getByLabel('Collection', { exact: true }).fill('Evening reads');
  await page.getByRole('button', { name: 'Add to library', exact: true }).click();
  await expect(page.getByRole('dialog')).not.toBeVisible();
  await page
    .getByRole('link')
    .filter({ has: page.getByRole('heading', { name: title, exact: true }) })
    .click();
  await page.getByRole('button', { name: 'Log a chapter' }).click();
  await expect(page.getByText('1 / 10 chapters')).toBeVisible();
  await page.reload();
  await expect(page.getByText('1 / 10 chapters')).toBeVisible();
  await page.getByRole('link', { name: 'Activity', exact: true }).click();
  await expect(
    page
      .locator('.activity-list li')
      .filter({ hasText: title })
      .getByText('In progress · Progress: 1'),
  ).toBeVisible();
  await page.goto('/?q=patient');
  await expect(page.getByRole('heading', { name: title, exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'List view' }).click();
  await expect(page.locator('.item-list')).toBeVisible();
  await page.goto('/library');
  await page.screenshot({
    path: `test-results/library-${testInfo.project.name}.png`,
    fullPage: true,
  });
});
test('import a local text file, read it, and save completion', async ({ page }, testInfo) => {
  await page.goto('/library');
  await page.getByRole('button', { name: 'Import', exact: true }).click();
  await page.locator('input[type=file]').setInputFiles({
    name: `A short story ${testInfo.project.name}.txt`,
    mimeType: 'text/plain',
    buffer: Buffer.from('A small beginning.\n\nA satisfying ending.'),
  });
  await page.getByRole('button', { name: 'Import file', exact: true }).click();
  await expect(page.getByRole('dialog')).not.toBeVisible();
  await page
    .getByRole('link')
    .filter({
      has: page.getByRole('heading', {
        name: `A short story ${testInfo.project.name}`,
        exact: true,
      }),
    })
    .click();
  await page.getByRole('link', { name: 'Open reader' }).click();
  await expect(page.getByText('A small beginning.')).toBeVisible();
  await page.getByRole('button', { name: 'Finish reading' }).click();
  await page.getByRole('link', { name: 'Library', exact: true }).click();
  await expect(page.locator('.detail-copy>.status-badge')).toHaveText('Completed');
});
test('mobile layout stays within the viewport and navigation works', async ({ page }, testInfo) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/library');
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
  await page.screenshot({
    path: `test-results/library-mobile-${testInfo.project.name}.png`,
    fullPage: true,
  });
  await page.getByRole('button', { name: 'Open navigation' }).click();
  await page.getByRole('link', { name: 'Extensions', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'A little more possibility.' })).toBeVisible();
  await page.screenshot({ path: 'test-results/extensions-mobile.png', fullPage: true });
});
