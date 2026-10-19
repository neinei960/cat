import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

function assert(condition: unknown, message: string) {
  if (!condition) throw new Error(message)
}

const page = readFileSync(resolve(process.cwd(), 'src/pages/appointment/create.vue'), 'utf8')

assert(
  page.includes('const customerMatches =') && page.includes('if (!customerMatches) return []'),
  'customer suggestions must locally filter customer fields against the current keyword',
)

assert(
  page.includes('let customerSearchRequestVersion = 0'),
  'customer search must track the latest request version',
)

assert(
  page.includes('requestVersion !== customerSearchRequestVersion'),
  'stale customer search responses must not overwrite the latest result',
)

assert(
  /function onCustomerKeywordInput\(\) \{\s*customerSuggestionOpen\.value = true\s*customerSearchRequestVersion \+= 1/.test(page),
  'typing a new keyword must immediately invalidate in-flight searches',
)

assert(
  /<img[\s\S]*?class="suggestion-pet-avatar"[\s\S]*?:src="item\.pet\.avatar"[\s\S]*?@error="onSuggestionAvatarError\(item\)"/.test(page),
  'pet suggestions must render their own pet avatar and handle broken images',
)
assert(
  page.includes('failedSuggestionAvatars[item.key] !== item.pet.avatar'),
  'missing or failed avatars must not render a broken image',
)
assert(
  /\.suggestion-pet-avatar\s*\{[^}]*width: 64px;[^}]*height: 64px;[^}]*flex-shrink: 0;/.test(page),
  'avatar dimensions must be stable and must not shrink beside text',
)

console.log('appointment customer search checks passed')
