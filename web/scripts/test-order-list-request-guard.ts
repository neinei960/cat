import fs from 'node:fs'
import path from 'node:path'

function main() {
  const root = process.cwd()
  const source = fs.readFileSync(path.join(root, 'src/pages/order/list.vue'), 'utf8')

  assert(source.includes('latestLoadRequestId'), 'order list should track the latest load request id')
  assert(source.includes('const requestId = ++latestLoadRequestId'), 'loadData should allocate a monotonically increasing request id')
  assert(source.includes('if (requestId !== latestLoadRequestId) return'), 'stale order list responses should be ignored')
  assert(source.includes('if (requestId === latestLoadRequestId) {\n      loading.value = false\n    }'), 'only the latest request should clear the loading state')
}

main()

function assert(condition: unknown, label: string) {
  if (!condition) {
    throw new Error(label)
  }
}
