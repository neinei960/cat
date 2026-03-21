import { test, expect } from '@playwright/test';

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

async function login(page: any) {
  await page.goto('/');
  await page.waitForTimeout(2000);
  const inputs = page.locator('input');
  await inputs.nth(0).fill('13800138000');
  await inputs.nth(1).fill('123456');
  await page.locator('uni-button, button').filter({ hasText: /登/ }).first().click({ force: true });
  await page.waitForTimeout(5000);
}

test.describe('New Customer Template Parse', () => {
  test('navigate to appointment create and parse template', async ({ page }) => {
    // Step 1: Login
    await login(page);

    // Step 2: Navigate to appointment create page
    await page.goto('/#/pages/appointment/create');
    await page.waitForTimeout(3000);

    // Take screenshot to see current state
    await page.screenshot({ path: 'test-results/01-appointment-create.png' });

    // Step 3: Look for the page content
    const content = await page.content();
    console.log('Page has 新客:', content.includes('新客'));
    console.log('Page has 熟客:', content.includes('熟客'));

    // Step 4: Click "新客" tab
    const newCustomerTab = page.locator('text=新客').first();
    const tabExists = await newCustomerTab.isVisible().catch(() => false);
    console.log('新客 tab visible:', tabExists);

    if (tabExists) {
      await newCustomerTab.click();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: 'test-results/02-new-customer-tab.png' });
    } else {
      // Try alternative selectors
      console.log('Trying alternative selectors...');
      const allText = await page.locator('view, text, div, span').allTextContents();
      console.log('All visible texts:', allText.filter(t => t.trim()).slice(0, 30));
      await page.screenshot({ path: 'test-results/02-no-tab-found.png' });
      // Fail with helpful info
      expect(tabExists, 'Could not find 新客 tab').toBeTruthy();
      return;
    }

    // Step 5: Find textarea and paste template
    const textarea = page.locator('textarea').first();
    const textareaVisible = await textarea.isVisible().catch(() => false);
    console.log('Textarea visible:', textareaVisible);

    if (textareaVisible) {
      await textarea.fill(TEMPLATE_TEXT);
      await page.waitForTimeout(500);
      await page.screenshot({ path: 'test-results/03-template-pasted.png' });
    } else {
      console.log('No textarea found');
      await page.screenshot({ path: 'test-results/03-no-textarea.png' });
      expect(textareaVisible, 'Could not find textarea').toBeTruthy();
      return;
    }

    // Step 6: Click parse button
    const parseBtn = page.locator('button, uni-button').filter({ hasText: /解析/ }).first();
    const parseBtnVisible = await parseBtn.isVisible().catch(() => false);
    console.log('Parse button visible:', parseBtnVisible);

    if (parseBtnVisible) {
      await parseBtn.click({ force: true });
      await page.waitForTimeout(2000);
      await page.screenshot({ path: 'test-results/04-after-parse.png' });
    } else {
      console.log('No parse button found');
      await page.screenshot({ path: 'test-results/04-no-parse-btn.png' });
      expect(parseBtnVisible, 'Could not find parse button').toBeTruthy();
      return;
    }

    // Step 7: Verify parsed results
    const pageContent = await page.content();
    console.log('Has 黑凤:', pageContent.includes('黑凤'));
    console.log('Has 混合:', pageContent.includes('混合'));
    console.log('Has e猫:', pageContent.includes('e猫'));
    console.log('Has 三针已打完:', pageContent.includes('三针已打完'));

    // Check that parsed fields appear in the page
    expect(pageContent).toContain('黑凤');
    expect(pageContent).toContain('混合');

    // Check input values for parsed data
    const allInputs = page.locator('input');
    const inputCount = await allInputs.count();
    console.log('Total inputs after parse:', inputCount);
    for (let i = 0; i < inputCount; i++) {
      const val = await allInputs.nth(i).inputValue().catch(() => '');
      if (val) console.log(`Input ${i}: "${val}"`);
    }

    await page.screenshot({ path: 'test-results/05-parse-results.png', fullPage: true });

    console.log('Parse test completed successfully!');
  });
});
