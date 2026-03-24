import { test, expect } from '@playwright/test';

const PROD_URL = 'http://36.151.144.227';

// Reusable login helper
async function login(page: any) {
  await page.goto(`${PROD_URL}/#/pages/login/index`);
  await page.waitForLoadState('networkidle');
  const inputs = page.locator('input');
  await inputs.first().click();
  await inputs.first().fill('13800138000');
  await inputs.nth(1).click();
  await inputs.nth(1).fill('123456');
  const loginBtn = page.locator('button, [class*="btn"]').filter({ hasText: /登/ });
  await loginBtn.first().click();
  await page.waitForTimeout(2000);
}

test.describe('Smoke tests', () => {
  test.setTimeout(60000);

  test('login and home page loads', async ({ page }) => {
    await login(page);
    // Should redirect to index after login
    expect(page.url()).toContain('/pages/index/index');
  });

  test('pet list page loads', async ({ page }) => {
    await login(page);
    await page.goto(`${PROD_URL}/#/pages/pet/list`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    // Should see owner groups or empty state
    const hasContent = await page.locator('.owner-group, .empty').first().isVisible();
    expect(hasContent).toBe(true);
  });

  test('order list page loads', async ({ page }) => {
    await login(page);
    await page.goto(`${PROD_URL}/#/pages/order/list`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    const hasContent = await page.locator('.card, .order-card, .empty').first().isVisible();
    expect(hasContent).toBe(true);
  });

  test('customer list page loads', async ({ page }) => {
    await login(page);
    await page.goto(`${PROD_URL}/#/pages/customer/list`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    const hasContent = await page.locator('.card, .empty').first().isVisible();
    expect(hasContent).toBe(true);
  });

  test('member card settings page loads', async ({ page }) => {
    await login(page);
    await page.goto(`${PROD_URL}/#/pages/setting/member-card`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    const hasContent = await page.locator('.card, .empty').first().isVisible();
    expect(hasContent).toBe(true);
  });

  test('no console errors on key pages', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error' && !msg.text().includes('favicon')) {
        errors.push(msg.text());
      }
    });

    await login(page);

    const pages = [
      '/pages/pet/list',
      '/pages/order/list',
      '/pages/customer/list',
    ];

    for (const p of pages) {
      await page.goto(`${PROD_URL}/#${p}`);
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(800);
    }

    // Filter out non-critical errors
    const critical = errors.filter(e => !e.includes('net::') && !e.includes('404'));
    expect(critical).toEqual([]);
  });
});
