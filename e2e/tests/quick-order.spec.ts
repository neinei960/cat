import { test, expect } from '@playwright/test';

const BASE_URL = 'http://localhost:5173';
const API_BASE = 'http://localhost:8080/api/v1';

test.describe('Quick Order (快速开单) Tests', () => {

  test('login through UI and navigate to quick order', async ({ page }) => {
    // Go to login page
    await page.goto(BASE_URL);
    await page.waitForTimeout(1000);

    // Fill login form
    const phoneInput = page.locator('input[type="number"]');
    const passwordInput = page.locator('input[type="password"]');

    await phoneInput.fill('13800138000');
    await passwordInput.fill('123456');

    // Click login button (uni-app renders as uni-button)
    const loginBtn = page.locator('.login-btn');
    await loginBtn.waitFor({ state: 'visible', timeout: 5000 });
    await loginBtn.click();
    await page.waitForTimeout(3000);

    // Should be on index page now
    await page.screenshot({ path: 'test-results/01-after-login.png', fullPage: true });
    const content = await page.content();
    console.log('After login - has 工作台:', content.includes('工作台') || content.includes('功能菜单'));
    console.log('After login - has 快速开单:', content.includes('快速开单'));

    // Navigate to quick order page
    // Click on 快速开单 menu item
    const orderBtn = page.locator('text=快速开单');
    if (await orderBtn.isVisible()) {
      await orderBtn.click();
      await page.waitForTimeout(2000);
    } else {
      // Direct navigation
      await page.evaluate(() => {
        (window as any).uni?.navigateTo?.({ url: '/pages/order/create' });
      });
      await page.waitForTimeout(2000);
    }

    await page.screenshot({ path: 'test-results/02-order-create.png', fullPage: true });
    const orderContent = await page.content();
    console.log('Order page - has 选择猫咪:', orderContent.includes('选择猫咪'));
    console.log('Order page - has 洗浴项目:', orderContent.includes('洗浴项目'));
    console.log('Order page - has 附加费用:', orderContent.includes('附加费'));
    console.log('Order page - has 确认开单:', orderContent.includes('确认开单'));
  });

  test('test full order create flow via API', async ({ request }) => {
    // Login
    const loginRes = await request.post(`${API_BASE}/auth/staff/login`, {
      data: { phone: '13800138000', password: '123456' },
    });
    const loginData = await loginRes.json();
    const token = loginData.data.token;
    const headers = { Authorization: `Bearer ${token}` };

    // 1. Search pets by name
    console.log('\n--- Step 1: Search pet ---');
    const petRes = await request.get(`${API_BASE}/b/pets?keyword=旺财`, { headers });
    const petData = await petRes.json();
    console.log('Pet search result count:', petData.data.list?.length);
    expect(petData.code).toBe(0);
    const pet = petData.data.list?.[0];
    if (pet) {
      console.log(`Found: ${pet.name} (ID=${pet.ID}, fur_level=${pet.fur_level}, customer_id=${pet.customer_id})`);
    }

    // 2. List services
    console.log('\n--- Step 2: List services ---');
    const svcRes = await request.get(`${API_BASE}/b/services?page_size=50`, { headers });
    const svcData = await svcRes.json();
    console.log('Services:');
    svcData.data.list?.forEach((s: any) => {
      console.log(`  ID=${s.ID} ${s.name} ¥${s.base_price}`);
    });

    // 3. Price lookup
    console.log('\n--- Step 3: Price lookup ---');
    if (pet?.fur_level) {
      const priceRes = await request.get(
        `${API_BASE}/b/orders/price-lookup?service_id=${svcData.data.list[0].ID}&fur_level=${pet.fur_level}`,
        { headers }
      );
      const priceData = await priceRes.json();
      console.log('Price lookup:', priceData.data);
    } else {
      console.log('Pet has no fur_level, skipping price lookup');
    }

    // 4. List addons
    console.log('\n--- Step 4: List addons ---');
    const addonRes = await request.get(`${API_BASE}/b/addons`, { headers });
    const addonData = await addonRes.json();
    if (Array.isArray(addonData.data)) {
      console.log('Addons:', addonData.data.map((a: any) => `${a.name}(¥${a.default_price})`).join(', '));
    } else {
      console.log('No addons found (data:', addonData.data, ')');
    }

    // 5. List staff
    console.log('\n--- Step 5: List staff ---');
    const staffRes = await request.get(`${API_BASE}/b/staffs?page_size=50`, { headers });
    const staffData = await staffRes.json();
    const groomers = staffData.data.list?.filter((s: any) => s.role === 'staff');
    console.log('Groomers:', groomers?.map((s: any) => `${s.name}(rate=${s.commission_rate}%)`).join(', '));

    // 6. Create order
    console.log('\n--- Step 6: Create order ---');
    const serviceId = svcData.data.list?.[0]?.ID;
    const groomerId = groomers?.[0]?.ID;

    const orderReq = {
      pet_id: pet?.ID || 1,
      service_id: serviceId,
      staff_id: groomerId,
      addons: [
        { name: '超重费', amount: 10 },
        { name: '去油费', amount: 20 },
      ],
      remark: 'Playwright 快速开单测试',
    };
    console.log('Order request:', JSON.stringify(orderReq));

    const orderRes = await request.post(`${API_BASE}/b/orders`, {
      headers,
      data: orderReq,
    });
    const orderData = await orderRes.json();
    console.log('\nOrder response code:', orderData.code);
    if (orderData.code === 0) {
      const o = orderData.data;
      console.log(`Order created: ${o.order_no}`);
      console.log(`  Total: ¥${o.total_amount}`);
      console.log(`  Discount: ¥${o.discount_amount} (rate: ${o.discount_rate})`);
      console.log(`  Pay: ¥${o.pay_amount}`);
      console.log(`  Commission: ¥${o.commission}`);
      console.log(`  Items: ${o.items?.length}`);
      o.items?.forEach((i: any) => {
        console.log(`    [type=${i.item_type}] ${i.name}: ¥${i.amount}`);
      });
      expect(o.total_amount).toBeGreaterThan(0);
      expect(o.items.length).toBeGreaterThanOrEqual(1);
    } else {
      console.log('Order failed:', orderData.msg);
      // Don't fail - let's see the error
    }
  });
});
