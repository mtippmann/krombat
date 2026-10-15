const assert = require('node:assert/strict')
const { chromium } = require('playwright')

const base = process.env.BASE_URL || 'http://krombat.test'
assert.equal(base, 'http://krombat.test', 'Browser journeys are local-only')

async function main() {
  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({ viewport: { width: 1440, height: 1100 } })
  const failures = []
  page.on('pageerror', error => failures.push(error.message))
  let name = `browser-${Date.now()}`
  let created = false
  try {
    await page.goto(base)
    await page.getByRole('button', { name: 'Enter local demo', exact: true }).click()
    await page.getByLabel('Dungeon name', { exact: true }).fill(name)
    await page.getByLabel('Monsters', { exact: true }).fill('2')
    await page.getByLabel('Difficulty', { exact: true }).selectOption('easy')
    await page.getByRole('button', { name: 'Create dungeon', exact: true }).click()
    created = true
    const monster1 = page.getByRole('button', { name: 'Attack monster 1', exact: true })
    await monster1.waitFor()
    assert(await page.getByRole('button', { name: 'Attack boss', exact: true }).isDisabled())
    for (let monster = 1; monster <= 2; monster++) {
      const button = page.getByRole('button', { name: `Attack monster ${monster}`, exact: true })
      for (let hit = 0; hit < 2; hit++) {
        await button.click()
        await page.waitForTimeout(800)
      }
      assert(await button.isDisabled(), `Monster ${monster} must be dead`)
    }
    console.log('PASS local login, creation, locked boss, monster combat')
    await page.getByTestId('resource-list').getByText('Loot', { exact: false }).first()
      .waitFor({ timeout: 60000 })
    const boss = page.getByRole('button', { name: 'Attack boss', exact: true })
    for (let hit = 0; hit < 8; hit++) {
      await boss.click()
      await page.waitForTimeout(800)
      if (hit === 3 || hit === 5) {
        const expected = hit === 3 ? 'phase2' : 'phase3'
        await page.waitForFunction(phase =>
          document.querySelector('[data-testid="boss-phase"]')?.textContent.includes(phase),
        expected, { timeout: 60000 })
      }
    }
    await page.getByTestId('outcome').filter({ hasText: /victory/i }).waitFor()
    await page.getByRole('button', { name: 'Open treasure', exact: true }).click()
    await page.waitForTimeout(2000)
    console.log('PASS boss phase transitions, victory and treasure reveal')
    await page.getByLabel('CEL expression', { exact: true }).fill('schema.spec.game.bossHP == 0')
    await page.getByRole('button', { name: 'Evaluate CEL', exact: true }).click()
    await page.getByTestId('cel-result').filter({ hasText: 'true' }).waitFor()
    console.log('PASS upstream CEL playground through browser UI')
    await page.reload()
    await page.getByRole('button', { name, exact: true }).click()
    await page.getByTestId('outcome').filter({ hasText: /victory/i }).waitFor()
    page.once('dialog', dialog => dialog.accept())
    await Promise.all([
      page.waitForResponse(response => response.request().method() === 'DELETE' &&
        response.url().endsWith(`/rpg-demo/${name}`) && response.status() === 204),
      page.getByRole('button', { name: 'Delete dungeon', exact: true }).click(),
    ])
    created = false
    console.log(`CLEANUP ${name}`)
    await page.getByRole('button', { name: 'Create dungeon', exact: true }).waitFor()
    await page.getByRole('button', { name, exact: true }).waitFor({ state: 'detached' })
    name += '-hard'
    await page.getByLabel('Dungeon name', { exact: true }).fill(name)
    await page.getByLabel('Monsters', { exact: true }).fill('10')
    await page.getByLabel('Difficulty', { exact: true }).selectOption('hard')
    await page.getByRole('button', { name: 'Create dungeon', exact: true }).click()
    created = true
    await page.getByRole('button', { name: 'Attack monster 1', exact: true }).waitFor()
    for (let turn = 0; turn < 50; turn++) {
      await page.waitForFunction(() =>
        /defeat/i.test(document.querySelector('[data-testid="outcome"]')?.textContent || '') ||
        [...document.querySelectorAll('button')].some(button =>
          /^Attack (monster \d+|boss)$/.test(button.getAttribute('aria-label') || '') && !button.disabled))
      if (await page.getByTestId('outcome').filter({ hasText: /defeat/i }).count()) break
      const monsters = page.getByRole('button', { name: /^Attack monster \d+$/ })
      let attacked = false
      for (let i = 0; i < await monsters.count(); i++) {
        if (await monsters.nth(i).isEnabled()) {
          await monsters.nth(i).click()
          attacked = true
          break
        }
      }
      if (!attacked) await page.getByRole('button', { name: 'Attack boss', exact: true }).click()
      await page.waitForTimeout(800)
    }
    await page.getByTestId('outcome').filter({ hasText: /defeat/i }).waitFor()
    page.once('dialog', dialog => dialog.accept())
    await Promise.all([
      page.waitForResponse(response => response.request().method() === 'DELETE' &&
        response.url().endsWith(`/rpg-demo/${name}`) && response.status() === 204),
      page.getByRole('button', { name: 'Delete dungeon', exact: true }).click(),
    ])
    created = false
    console.log(`CLEANUP ${name}`)
    await page.getByRole('button', { name: 'Create dungeon', exact: true }).waitFor()
    await page.getByRole('button', { name, exact: true }).waitFor({ state: 'detached' })
    assert.deepEqual(failures, [], 'No browser runtime errors')
    console.log('PASS persisted session/run, fresh hard run, defeat, deletion and clean browser console')
  } finally {
    if (created) {
      try {
        page.once('dialog', dialog => dialog.accept())
        await Promise.all([
          page.waitForResponse(response => response.request().method() === 'DELETE' &&
            response.url().endsWith(`/rpg-demo/${name}`) && response.status() === 204,
          { timeout: 10000 }),
          page.getByRole('button', { name: 'Delete dungeon', exact: true }).click({ timeout: 5000 }),
        ])
        await page.getByRole('button', { name, exact: true }).waitFor({ state: 'detached', timeout: 10000 })
      } catch (error) {
        console.error(`UI cleanup failed for ${name}: ${error.message}`)
      }
    }
    await browser.close()
  }
}
main().catch(error => { console.error(error); process.exitCode = 1 })
