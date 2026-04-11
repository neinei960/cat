import { test, expect } from '@playwright/test';

const PROD_URL = 'http://36.151.144.227';

async function login(page: any) {
  await page.goto(`${PROD_URL}/#/pages/login/index`);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);
  const inputs = page.locator('input');
  await inputs.first().fill('15989005706');
  await inputs.nth(1).fill('123456');
  const loginBtn = page.locator('button, [class*="btn"]').filter({ hasText: /进入|登/ });
  await loginBtn.first().click();
  await page.waitForTimeout(3000);
}

test('delete pending order - UI flow', async ({ page }) => {
  test.setTimeout(60000);
  await login(page);

  await page.goto(`${PROD_URL}/#/pages/order/list`);
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(2000);
  await page.screenshot({ path: 'test-results/01-order-list.png' });

  // Find pending order
  const pendingCard = page.locator('.card').filter({ hasText: '待付款' }).first();
  const cardVisible = await pendingCard.isVisible().catch(() => false);
  console.log('Pending card visible:', cardVisible);
  if (!cardVisible) {
    console.log('No pending orders to test');
    return;
  }

  // Long press
  const box = await pendingCard.boundingBox();
  if (!box) return;

  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  await page.mouse.down();
  await page.waitForTimeout(600);
  await page.mouse.up();
  await page.waitForTimeout(1500);
  await page.screenshot({ path: 'test-results/02-after-longpress.png' });

  // Check action sheet items
  const allText = await page.locator('.uni-actionsheet-cell').allTextContents().catch(() => []);
  console.log('Action sheet items:', allText);

  const deleteOption = page.locator('.uni-actionsheet-cell').filter({ hasText: '删除' });
  const deleteVisible = await deleteOption.isVisible().catch(() => false);
  console.log('Delete button visible:', deleteVisible);

  // Also check the whole page for any action sheet
  const pageText = await page.textContent('body');
  console.log('Page contains 删除订单:', pageText?.includes('删除订单'));
  console.log('Page contains 修改订单:', pageText?.includes('修改订单'));

  await page.screenshot({ path: 'test-results/03-final.png' });
});

test('delete pending order via API', async ({ request }) => {
  const loginRes = await request.post(`${PROD_URL}/api/v1/auth/staff/login`, {
    data: { phone: '15989005706', password: '123456' }
  });
  const loginData = await loginRes.json();
  const token = loginData?.data?.token;
  expect(token).toBeTruthy();

  const listRes = await request.get(`${PROD_URL}/api/v1/b/orders?status=0&page=1&page_size=5`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  const listData = await listRes.json();
  const orders = listData?.data?.list || [];
  console.log('Pending orders:', orders.length);

  if (orders.length === 0) {
    console.log('No pending orders');
    return;
  }

  const orderId = orders[0].ID;
  const deleteRes = await request.delete(`${PROD_URL}/api/v1/b/orders/${orderId}`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  const deleteData = await deleteRes.json();
  console.log('Delete result:', JSON.stringify(deleteData));
  expect(deleteRes.status()).toBe(200);
  expect(deleteData.code).toBe(0);
});
