import { test, expect } from '@playwright/test';

const PROD_URL = 'http://36.151.144.227';

test('receipt image generation works', async ({ page }) => {
  test.setTimeout(90000);

  // 1. Login - use direct token approach via localStorage
  await page.goto(`${PROD_URL}/#/pages/login/index`);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);

  // Debug: dump the HTML to find correct input selectors
  const inputs = await page.locator('input').all();
  console.log(`Found ${inputs.length} input elements`);
  for (const input of inputs) {
    const placeholder = await input.getAttribute('placeholder');
    const type = await input.getAttribute('type');
    const cls = await input.getAttribute('class');
    console.log(`  input: type=${type}, placeholder=${placeholder}, class=${cls}`);
  }

  // Try clicking directly on the input area and typing
  const phoneInput = page.locator('input').first();
  await phoneInput.click();
  await phoneInput.fill('13800138000');

  const pwdInput = page.locator('input').nth(1);
  await pwdInput.click();
  await pwdInput.fill('123456');

  // Click login button
  const loginBtn = page.locator('button, .login-btn, [class*="btn"]').filter({ hasText: /登/ });
  await loginBtn.first().click();
  await page.waitForTimeout(3000);

  // Check if we're logged in - should redirect
  await page.screenshot({ path: 'tests/screenshots/after-login.png' });
  console.log('Current URL after login:', page.url());

  // 2. Navigate to orders
  await page.goto(`${PROD_URL}/#/pages/order/list`);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(2000);
  await page.screenshot({ path: 'tests/screenshots/order-list.png' });

  // 3. Click first order
  const orderCards = page.locator('.card, .order-card, [class*="order-item"]');
  const count = await orderCards.count();
  console.log(`Found ${count} order cards`);

  if (count === 0) {
    // Maybe we need to look for a different selector
    const html = await page.locator('body').innerHTML();
    console.log('Page body (first 2000 chars):', html.substring(0, 2000));
    throw new Error('No order cards found');
  }

  await orderCards.first().click();
  await page.waitForTimeout(2000);
  await page.screenshot({ path: 'tests/screenshots/order-detail.png' });

  // 4. Click "生成小票"
  const receiptBtn = page.getByText('生成小票');
  await expect(receiptBtn).toBeVisible({ timeout: 5000 });
  await receiptBtn.click();
  await page.waitForTimeout(1000);

  // 5. Verify receipt modal
  await page.screenshot({ path: 'tests/screenshots/receipt-modal.png' });
  const receiptContent = page.locator('#receiptContent');
  await expect(receiptContent).toBeVisible({ timeout: 5000 });

  // 6. Click "生成图片"
  const generateBtn = page.locator('.btn-receipt-save');
  await expect(generateBtn).toBeVisible({ timeout: 5000 });

  const consoleErrors: string[] = [];
  const consoleLogs: string[] = [];
  page.on('console', msg => {
    if (msg.type() === 'error') consoleErrors.push(msg.text());
    consoleLogs.push(`[${msg.type()}] ${msg.text()}`);
  });

  await generateBtn.click();
  console.log('Clicked 生成图片');

  // 7. Wait for result
  // Either the image appears, or check for "生成中..." text
  await page.waitForTimeout(2000);
  await page.screenshot({ path: 'tests/screenshots/after-generate-click.png' });

  const receiptImage = page.locator('.receipt-image');
  try {
    await expect(receiptImage).toBeVisible({ timeout: 20000 });
    console.log('SUCCESS: Receipt image generated!');

    // uni-app <image> compiles to <uni-image> with <img> inside
    const imgEl = receiptImage.locator('img');
    const src = await imgEl.getAttribute('src').catch(() => null)
      || await receiptImage.getAttribute('src');
    expect(src).toBeTruthy();
    expect(src!.startsWith('blob:') || src!.startsWith('data:image/png;base64,')).toBe(true);
    console.log(`Image src: ${src!.substring(0, 30)}... (${src!.length} chars)`);

    // Verify download button exists
    const downloadBtn = page.locator('.btn-receipt-save');
    await expect(downloadBtn).toBeVisible({ timeout: 3000 });
    const dlText = await downloadBtn.textContent();
    console.log(`Download button text: ${dlText}`);

    await page.screenshot({ path: 'tests/screenshots/receipt-image-success.png' });
  } catch (e) {
    await page.screenshot({ path: 'tests/screenshots/receipt-image-failure.png' });
    console.log('Console errors:', consoleErrors);
    console.log('All console logs:', consoleLogs);

    const btnText = await generateBtn.textContent().catch(() => 'N/A');
    console.log('Button text:', btnText);

    throw e;
  }
});
