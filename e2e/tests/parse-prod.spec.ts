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

test('prod: parse template on appointment create page', async ({ page }) => {
  // Enable console logging
  page.on('console', msg => console.log('BROWSER:', msg.type(), msg.text()));
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));

  // Step 1: Login
  console.log('=== Step 1: Login ===');
  await page.goto(PROD_URL);
  await page.waitForTimeout(3000);
  await page.screenshot({ path: 'test-results/prod-01-login.png' });

  const inputs = page.locator('input');
  const inputCount = await inputs.count();
  console.log('Login page input count:', inputCount);

  if (inputCount >= 2) {
    await inputs.nth(0).fill('13800138000');
    await inputs.nth(1).fill('123456');
    await page.locator('uni-button, button').filter({ hasText: /登/ }).first().click({ force: true });
    await page.waitForTimeout(5000);
  }

  await page.screenshot({ path: 'test-results/prod-02-after-login.png' });
  console.log('Current URL after login:', page.url());

  // Step 2: Navigate to appointment create
  console.log('=== Step 2: Navigate ===');
  await page.goto(`${PROD_URL}/#/pages/appointment/create`);
  await page.waitForTimeout(3000);
  await page.screenshot({ path: 'test-results/prod-03-create-page.png' });

  const pageContent = await page.content();
  console.log('Has 熟客:', pageContent.includes('熟客'));
  console.log('Has 新客:', pageContent.includes('新客'));
  console.log('Has 选客户:', pageContent.includes('选客户'));

  // Dump all visible text for debugging
  const bodyText = await page.locator('body').innerText().catch(() => '');
  const uniqueTexts = bodyText.split('\n').map(s => s.trim()).filter(s => s.length > 0 && s.length < 50);
  console.log('Visible texts on page:', JSON.stringify(uniqueTexts.slice(0, 40)));

  // Step 3: Find and click 新客 tab
  console.log('=== Step 3: Click 新客 tab ===');

  // Try multiple selectors
  let clicked = false;
  const selectors = [
    'text=新客',
    '.tab:has-text("新客")',
    '.tab-active:has-text("新客")',
    'view:has-text("新客")',
  ];

  for (const sel of selectors) {
    const el = page.locator(sel).first();
    const visible = await el.isVisible().catch(() => false);
    console.log(`Selector "${sel}" visible: ${visible}`);
    if (visible && !clicked) {
      await el.click({ force: true });
      clicked = true;
      console.log(`Clicked using: ${sel}`);
    }
  }

  await page.waitForTimeout(2000);
  await page.screenshot({ path: 'test-results/prod-04-new-customer-tab.png' });

  // Step 4: Check for textarea
  console.log('=== Step 4: Find textarea ===');
  const textareas = page.locator('textarea');
  const taCount = await textareas.count();
  console.log('Textarea count:', taCount);

  if (taCount === 0) {
    // Maybe the page didn't update - dump HTML
    const html = await page.content();
    console.log('Page HTML snippet:', html.substring(0, 3000));
    await page.screenshot({ path: 'test-results/prod-05-no-textarea.png', fullPage: true });
    expect(taCount, 'No textarea found on page').toBeGreaterThan(0);
    return;
  }

  // Find the template textarea (not the notes one)
  const templateTA = textareas.first();
  await templateTA.fill(TEMPLATE_TEXT);
  await page.waitForTimeout(1000);
  await page.screenshot({ path: 'test-results/prod-05-template-filled.png' });

  // Step 5: Click parse button
  console.log('=== Step 5: Click parse button ===');
  const parseBtn = page.locator('button, uni-button').filter({ hasText: /解析/ }).first();
  const parseBtnVisible = await parseBtn.isVisible().catch(() => false);
  console.log('Parse button visible:', parseBtnVisible);

  if (!parseBtnVisible) {
    // Try finding any buttons
    const allBtns = page.locator('button, uni-button');
    const btnCount = await allBtns.count();
    console.log('Total buttons:', btnCount);
    for (let i = 0; i < btnCount; i++) {
      const txt = await allBtns.nth(i).innerText().catch(() => '');
      console.log(`Button ${i}: "${txt}"`);
    }
    expect(parseBtnVisible, 'Parse button not found').toBeTruthy();
    return;
  }

  await parseBtn.click({ force: true });
  await page.waitForTimeout(2000);
  await page.screenshot({ path: 'test-results/prod-06-after-parse.png', fullPage: true });

  // Step 6: Verify parsed results
  console.log('=== Step 6: Verify results ===');
  const afterParseContent = await page.content();
  console.log('Has 黑凤:', afterParseContent.includes('黑凤'));
  console.log('Has 混合:', afterParseContent.includes('混合'));
  console.log('Has e猫:', afterParseContent.includes('e猫'));

  // Check all input values
  const allInputs = page.locator('input');
  const allInputCount = await allInputs.count();
  console.log('Total inputs after parse:', allInputCount);
  for (let i = 0; i < allInputCount; i++) {
    const val = await allInputs.nth(i).inputValue().catch(() => '');
    const placeholder = await allInputs.nth(i).getAttribute('placeholder').catch(() => '');
    console.log(`Input ${i}: value="${val}" placeholder="${placeholder}"`);
  }

  // Also check textarea values
  const allTA = page.locator('textarea');
  const allTACount = await allTA.count();
  for (let i = 0; i < allTACount; i++) {
    const val = await allTA.nth(i).inputValue().catch(() => '');
    const placeholder = await allTA.nth(i).getAttribute('placeholder').catch(() => '');
    console.log(`Textarea ${i}: value="${val}" placeholder="${placeholder}"`);
  }

  expect(afterParseContent).toContain('黑凤');
});
