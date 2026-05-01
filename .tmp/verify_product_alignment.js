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

  await page.goto('http://36.151.144.227/#/pages/product/list', { waitUntil: 'networkidle' })
  await page.locator('.page .title').first().waitFor({ timeout: 15000 })

  const searchInput = page.locator('input[placeholder="搜索商品名 / 品牌"]').first()
  if (await searchInput.count()) {
    await searchInput.fill('3M5酸奶兔排冻干')
    await page.keyboard.press('Enter')
    await page.waitForTimeout(1000)
  }

  await page.getByText('3M5酸奶兔排冻干').first().click()
  await page.waitForURL(/\/pages\/product\/edit/, { timeout: 15000 })
  await page.waitForLoadState('networkidle')
  await page.waitForTimeout(1500)

  const metrics = await page.evaluate(() => {
    const inputs = Array.from(document.querySelectorAll('input'))
    const inspect = (input) => {
      const style = window.getComputedStyle(input)
      return {
        className: input.className,
        placeholder: input.getAttribute('placeholder'),
        value: input.value,
        textAlign: style.textAlign,
        width: style.width,
        paddingLeft: style.paddingLeft,
        boxSizing: style.boxSizing,
      }
    }
    return {
      url: location.href,
      title: document.title,
      inputCount: inputs.length,
      inputs: inputs.map(inspect),
      bodyText: document.body.innerText.slice(0, 500),
    }
  })

  await page.screenshot({ path: '/Users/genglsh/workstation/cat/cat/.tmp/product-edit-alignment.png', fullPage: true })

  console.log(JSON.stringify({ metrics }, null, 2))
  await browser.close()
}

main().catch((error) => {
  console.error(error)
  process.exit(1)
})
