const { chromium, devices } = require('playwright')

async function main() {
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({
    ...devices['iPhone 13'],
  })
  const page = await context.newPage()

  await page.goto('http://36.151.144.227/#/', { waitUntil: 'networkidle' })
  await page.getByRole('spinbutton').fill('15989005706')
  await page.getByRole('textbox').fill('123456')
  await page.getByText('进入工作台').click()
  await page.waitForURL('**/pages/index/index', { timeout: 15000 })

  await page.goto('http://36.151.144.227/#/pages/boarding/dashboard', { waitUntil: 'networkidle' })
  await page.locator('.page .title').first().waitFor({ timeout: 15000 })

  const hasHistoryButton = await page.getByText('历史记录').first().isVisible()
  if (!hasHistoryButton) {
    throw new Error('历史记录按钮未出现')
  }

  await page.getByText('历史记录').first().click()
  await page.waitForURL(/\/pages\/boarding\/history/, { timeout: 15000 })
  await page.locator('.page .title').first().waitFor({ timeout: 15000 })
  await page.waitForFunction(
    () => !document.body.innerText.includes('加载中...'),
    { timeout: 10000 },
  )
  await page.waitForTimeout(800)

  const filterTexts = await page.locator('.filter-chip').allTextContents()
  await page.screenshot({
    path: '/Users/genglsh/workstation/cat/cat/.tmp/boarding-history-list.png',
    fullPage: true,
  })

  const historyCards = page.locator('.history-card')
  const historyCount = await historyCards.count()
  const historyBody = (await page.locator('body').innerText()).slice(0, 1000)

  let detail = null
  if (historyCount > 0) {
    await historyCards.first().click()
    await page.waitForURL(/\/pages\/boarding\/detail/, { timeout: 15000 })
    await page.getByText('家长与猫咪').first().waitFor({ timeout: 15000 })
    detail = await page.evaluate(() => ({
      summaryTitle: document.querySelector('.summary-title')?.textContent?.trim() || '',
      sectionTitles: Array.from(document.querySelectorAll('.section-title'))
        .map((node) => node.textContent?.trim())
        .filter(Boolean),
      bodyText: document.body.innerText.slice(0, 1200),
    }))
    await page.screenshot({
      path: '/Users/genglsh/workstation/cat/cat/.tmp/boarding-history-detail.png',
      fullPage: true,
    })
  } else {
    await page.screenshot({
      path: '/Users/genglsh/workstation/cat/cat/.tmp/boarding-history-empty.png',
      fullPage: true,
    })
  }

  console.log(JSON.stringify({
    hasHistoryButton,
    filterTexts,
    historyCount,
    historyBody,
    detail,
  }, null, 2))

  await browser.close()
}

main().catch((error) => {
  console.error(error)
  process.exit(1)
})
