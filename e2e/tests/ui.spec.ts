import { test, expect } from '@playwright/test';

test.describe('B-End UI Tests', () => {
  test('login page loads and displays correctly', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(2000);
    const body = await page.content();
    expect(body).toContain('宠物洗护管理系统');
    expect(body).toContain('员工登录');
    expect(body).toContain('手机号');
    expect(body).toContain('密码');
  });

  test('login and navigate to workstation', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(2000);

    // Fill in the form
    const inputs = page.locator('input');
    await inputs.nth(0).fill('13800138000');
    await inputs.nth(1).fill('123456');

    // Click the login button using force click on the text
    await page.locator('uni-button, button').filter({ hasText: /登/ }).first().click({ force: true });

    // Wait for API call to complete and page to navigate
    await page.waitForTimeout(5000);

    // Take screenshot for debugging
    const content = await page.content();
    const hasWorkstation = content.includes('你好') || content.includes('工作台') || content.includes('猫咪洗护') || content.includes('预约日历');

    if (!hasWorkstation) {
      // Maybe the page hasn't navigated yet, try waiting for URL change
      await page.waitForTimeout(3000);
      const content2 = await page.content();
      const hasWorkstation2 = content2.includes('你好') || content2.includes('猫咪洗护') || content2.includes('预约日历');
      expect(hasWorkstation2).toBeTruthy();
    } else {
      expect(hasWorkstation).toBeTruthy();
    }
  });

  test('workstation displays all menu items', async ({ page }) => {
    // Directly inject auth token to skip login
    await page.goto('/');
    await page.waitForTimeout(1000);

    // Use evaluate to set storage and reload
    await page.evaluate(() => {
      localStorage.setItem('token', 'test');
      localStorage.setItem('staffInfo', JSON.stringify({ id: 1, name: '张店长', role: 'admin', shopId: 1 }));
    });

    // First login properly via API
    const loginRes = await page.request.post('http://localhost:8080/api/v1/auth/staff/login', {
      data: { phone: '13800138000', password: '123456' },
    });
    const loginData = await loginRes.json();
    const token = loginData.data.token;

    // Set real token
    await page.evaluate((t) => {
      localStorage.setItem('token', t);
      localStorage.setItem('staffInfo', JSON.stringify({ id: 1, name: '张店长', role: 'admin', shopId: 1, phone: '13800138000' }));
    }, token);

    // Navigate to workstation directly
    await page.goto('/#/pages/index/index');
    await page.waitForTimeout(3000);

    const content = await page.content();
    // Sidebar navigation labels
    expect(content).toContain('预约日历');
    expect(content).toContain('客户管理');
    expect(content).toContain('猫咪管理');
    expect(content).toContain('服务管理');
    expect(content).toContain('员工管理');
    expect(content).toContain('订单管理');
    expect(content).toContain('数据看板');
    expect(content).toContain('店铺设置');
  });
});
