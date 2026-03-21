import { test, expect } from '@playwright/test';

const BASE_URL = 'http://localhost:5173';
const API_BASE = 'http://localhost:8080/api/v1';

test.describe.serial('UI Calendar + Quick Order Tests', () => {
  let token = '';

  test('login via API and get token', async ({ request }) => {
    const res = await request.post(`${API_BASE}/auth/staff/login`, {
      data: { phone: '13800138000', password: '123456' },
    });
    const data = await res.json();
    expect(data.code).toBe(0);
    token = data.data.token;
  });

  test('open calendar page and check appointments visible', async ({ page }) => {
    // Set token in storage before navigating
    await page.goto(BASE_URL);
    await page.evaluate((t) => {
      localStorage.setItem('token', t);
      localStorage.setItem('staff', JSON.stringify({ ID: 1, name: '张店长', role: 'admin', shop_id: 1 }));
    }, token);

    // Navigate to calendar
    await page.goto(`${BASE_URL}/#/pages/appointment/calendar`);
    await page.waitForTimeout(2000);

    // Take screenshot
    await page.screenshot({ path: 'test-results/calendar-page.png', fullPage: true });
    console.log('📸 Calendar screenshot saved to test-results/calendar-page.png');

    // Check page content
    const content = await page.content();
    console.log('Page title visible:', content.includes('时间'));
    console.log('Has 王技师 column:', content.includes('王技师'));
    console.log('Has 李技师 column:', content.includes('李技师'));
  });

  test('open order create page and check it loads', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.evaluate((t) => {
      localStorage.setItem('token', t);
      localStorage.setItem('staff', JSON.stringify({ ID: 1, name: '张店长', role: 'admin', shop_id: 1 }));
    }, token);

    // Navigate to quick order page
    await page.goto(`${BASE_URL}/#/pages/order/create`);
    await page.waitForTimeout(2000);

    // Take screenshot
    await page.screenshot({ path: 'test-results/order-create-page.png', fullPage: true });
    console.log('📸 Order create screenshot saved to test-results/order-create-page.png');

    // Check key elements exist
    const content = await page.content();
    console.log('Has 选择猫咪:', content.includes('选择猫咪') || content.includes('猫咪'));
    console.log('Has 洗浴项目:', content.includes('洗浴项目') || content.includes('服务'));
    console.log('Has 附加费用:', content.includes('附加费'));
    console.log('Has 确认开单:', content.includes('确认开单') || content.includes('开单'));
  });

  test('open appointment list page', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.evaluate((t) => {
      localStorage.setItem('token', t);
      localStorage.setItem('staff', JSON.stringify({ ID: 1, name: '张店长', role: 'admin', shop_id: 1 }));
    }, token);

    await page.goto(`${BASE_URL}/#/pages/appointment/list`);
    await page.waitForTimeout(2000);

    await page.screenshot({ path: 'test-results/appointment-list.png', fullPage: true });
    console.log('📸 Appointment list screenshot saved to test-results/appointment-list.png');

    const content = await page.content();
    console.log('Has 王技师:', content.includes('王技师'));
    console.log('Has 旺财:', content.includes('旺财'));
  });
});
