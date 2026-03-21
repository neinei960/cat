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
  viewport: { width: 375, height: 812 }, // iPhone X size
  userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)',
});

test('mobile: parse template end-to-end', async ({ page }) => {
  page.on('console', msg => console.log('BROWSER:', msg.type(), msg.text()));
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));

  // Login
  await page.goto(PROD_URL);
  await page.waitForTimeout(3000);
  const inputs = page.locator('input');
  await inputs.nth(0).fill('13800138000');
  await inputs.nth(1).fill('123456');
  await page.locator('uni-button, button').filter({ hasText: /登/ }).first().click({ force: true });
  await page.waitForTimeout(5000);

  // Navigate
  await page.goto(`${PROD_URL}/#/pages/appointment/create`);
  await page.waitForTimeout(3000);
  await page.screenshot({ path: 'test-results/mobile-01-page.png' });

  // Click 新客
  await page.locator('text=新客').first().click({ force: true });
  await page.waitForTimeout(1000);
  await page.screenshot({ path: 'test-results/mobile-02-new-tab.png' });

  // Paste template text - simulate real typing by using evaluate to set value
  const textarea = page.locator('textarea').first();
  await textarea.click();
  await textarea.fill(TEMPLATE_TEXT);
  await page.waitForTimeout(500);

  // Click parse - try scrolling to it first
  await page.evaluate(() => {
    const btn = document.querySelector('button');
    const allBtns = document.querySelectorAll('button, uni-button');
    for (const b of allBtns) {
      if (b.textContent && b.textContent.includes('解析')) {
        (b as HTMLElement).scrollIntoView();
        (b as HTMLElement).click();
        break;
      }
    }
  });
  await page.waitForTimeout(2000);

  // Scroll down to see results
  await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
  await page.waitForTimeout(500);
  await page.screenshot({ path: 'test-results/mobile-03-after-parse.png', fullPage: true });

  // Verify all parsed values in inputs
  const allInputs = page.locator('input');
  const count = await allInputs.count();
  console.log('--- Parsed field values ---');
  const values: Record<string, string> = {};
  for (let i = 0; i < count; i++) {
    const val = await allInputs.nth(i).inputValue().catch(() => '');
    const ph = await allInputs.nth(i).getAttribute('placeholder').catch(() => '');
    console.log(`  Input[${i}] placeholder="${ph}" value="${val}"`);
    if (val) values[`input_${i}`] = val;
  }

  // Check textarea values too
  const allTA = page.locator('textarea');
  const taCount = await allTA.count();
  for (let i = 0; i < taCount; i++) {
    const val = await allTA.nth(i).inputValue().catch(() => '');
    const ph = await allTA.nth(i).getAttribute('placeholder').catch(() => '');
    console.log(`  Textarea[${i}] placeholder="${ph}" value="${val?.substring(0, 50)}..."`);
  }

  // Assertions
  const pageContent = await page.content();

  // The key test: does "黑凤" appear in an input value (not just in the textarea)?
  const nameInput = allInputs.nth(2); // based on previous test, index 2 is name
  const nameVal = await nameInput.inputValue().catch(() => '');
  console.log(`\nName field value: "${nameVal}"`);
  expect(nameVal).toBe('黑凤');

  const breedVal = await allInputs.nth(3).inputValue().catch(() => '');
  console.log(`Breed field value: "${breedVal}"`);
  expect(breedVal).toBe('混合');

  const ageVal = await allInputs.nth(4).inputValue().catch(() => '');
  console.log(`Age field value: "${ageVal}"`);
  expect(ageVal).toBe('8个月');

  console.log('\n✅ All parse assertions passed!');
});
