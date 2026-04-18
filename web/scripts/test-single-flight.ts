import { singleFlight } from '../src/utils/single-flight'

function assertEqual(actual: unknown, expected: unknown, label: string) {
  if (actual !== expected) {
    throw new Error(`${label}: expected "${String(expected)}", got "${String(actual)}"`)
  }
}

function assert(condition: boolean, label: string) {
  if (!condition) {
    throw new Error(label)
  }
}

function wait(ms: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms)
  })
}

async function testBlocksConcurrentCalls() {
  let callCount = 0
  const run = singleFlight(async () => {
    callCount += 1
    await wait(30)
    return callCount
  })

  const [first, second] = await Promise.all([run(), run()])

  assertEqual(callCount, 1, 'concurrent.callCount')
  assertEqual(first, 1, 'concurrent.firstResult')
  assertEqual(second, undefined, 'concurrent.secondResult')
}

async function testUnlocksAfterResolve() {
  let callCount = 0
  const run = singleFlight(async () => {
    callCount += 1
    await wait(5)
    return callCount
  })

  await run()
  const second = await run()

  assertEqual(callCount, 2, 'resolve.callCount')
  assertEqual(second, 2, 'resolve.secondResult')
}

async function testUnlocksAfterReject() {
  let callCount = 0
  const run = singleFlight(async () => {
    callCount += 1
    await wait(5)
    if (callCount === 1) {
      throw new Error('first call fails')
    }
    return 'ok'
  })

  let failed = false
  try {
    await run()
  } catch (error) {
    failed = error instanceof Error && error.message === 'first call fails'
  }

  const second = await run()

  assert(failed, 'reject.firstCallShouldFail')
  assertEqual(callCount, 2, 'reject.callCount')
  assertEqual(second, 'ok', 'reject.secondResult')
}

async function main() {
  await testBlocksConcurrentCalls()
  await testUnlocksAfterResolve()
  await testUnlocksAfterReject()
  console.log('single-flight tests passed')
}

main().catch((error) => {
  console.error(error)
  process.exit(1)
})
