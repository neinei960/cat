import { test, expect } from '@playwright/test';

const PROD_URL = 'http://36.151.144.227';

const TEMPLATE_TEXT = `树街の猫|宝贝洗护小调查

📇基本信息
大名：黑凤
年龄：8个月
品种：混合
性别：公
是否绝育：否
是否毛发打结：是
上次洗澡时间：去年10月

🧑🏻‍⚕️健康情况
疫苗是否齐全：三针已打完
疾病史和健康情况：无疾病

💭性格小秘密
i猫e猫：e猫
对水 声音 风 陌生环境反应：不能进风箱，要有人在旁边`;

test.use({
  viewport: { width: 375, height: 812 },
  userAgent: 'Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36',
  hasTouch: true,
});

test('debug: real mobile click parse', async ({ page }) => {
  const errors: string[] = [];
  const logs: string[] = [];
  page.on('console', msg => {
    logs.push(`[${msg.type()}] ${msg.text()}`);
    if (msg.type() === 'error') errors.push(msg.text());
  });
  page.on('pageerror', err => errors.push(`PAGE ERROR: ${err.message}`));

  // Login
  await page.goto(PROD_URL);
  await page.waitForTimeout(3000);
  await page.locator('input').nth(0).fill('13800138000');
  await page.locator('input').nth(1).fill('123456');
  await page.locator('uni-button, button').filter({ hasText: /登/ }).first().click({ force: true });
  await page.waitForTimeout(5000);

  // Navigate
  await page.goto(`${PROD_URL}/#/pages/appointment/create`);
  await page.waitForTimeout(3000);

  // Click 新客 tab - use real tap
  const newTab = page.locator('text=新客').first();
  await newTab.tap();
  await page.waitForTimeout(1000);

  // Fill textarea - simulate real paste behavior
  const textarea = page.locator('textarea').first();
  await textarea.tap();
  await page.waitForTimeout(300);
  // Use keyboard type to simulate real input
  await textarea.fill('');
  await textarea.type(TEMPLATE_TEXT.substring(0, 50), { delay: 10 });
  // Now set full text via fill (simulating paste)
  await textarea.fill(TEMPLATE_TEXT);
  await page.waitForTimeout(500);

  // Check textarea value is set
  const taValue = await textarea.inputValue();
  console.log('Textarea value length:', taValue.length);
  console.log('Textarea has 大名：黑凤:', taValue.includes('大名：黑凤'));

  // Take screenshot before click
  await page.screenshot({ path: 'test-results/debug-01-before-parse.png', fullPage: true });

  // Find parse button and click it with real tap
  const parseBtn = page.locator('button, uni-button').filter({ hasText: /解析/ }).first();
  console.log('Parse button text:', await parseBtn.innerText());
  console.log('Parse button visible:', await parseBtn.isVisible());

  // Scroll into view and tap
  await parseBtn.scrollIntoViewIfNeeded();
  await page.waitForTimeout(300);
  await parseBtn.tap();
  console.log('Tapped parse button');

  // Wait and check for errors
  await page.waitForTimeout(3000);

  console.log('\n=== Console errors ===');
  errors.forEach(e => console.log('ERROR:', e));
  console.log('Total errors:', errors.length);

  // Take screenshot after parse
  await page.screenshot({ path: 'test-results/debug-02-after-parse.png', fullPage: true });

  // Check if parsed section appeared
  const content = await page.content();
  const hasParsedSection = content.includes('解析结果');
  console.log('\nHas 解析结果 section:', hasParsedSection);
  console.log('Has 黑凤 in page:', content.includes('黑凤'));

  // Check rendered text (not input values, but what's actually visible)
  const visibleText = await page.locator('body').innerText();
  console.log('\nVisible text contains 黑凤:', visibleText.includes('黑凤'));
  console.log('Visible text contains 混合:', visibleText.includes('混合'));
  console.log('Visible text contains 8个月:', visibleText.includes('8个月'));
  console.log('Visible text contains 公:', visibleText.includes('公'));

  // Check input values
  const allInputs = page.locator('input');
  const inputCount = await allInputs.count();
  console.log('\nInput count:', inputCount);
  for (let i = 0; i < inputCount; i++) {
    const val = await allInputs.nth(i).inputValue().catch(() => '');
    if (val) console.log(`Input[${i}]: "${val}"`);
  }

  // Final assertions
  expect(hasParsedSection, 'Parsed section should appear').toBeTruthy();
  expect(errors.length, `Should have no JS errors. Errors: ${errors.join(', ')}`).toBe(0);
});
