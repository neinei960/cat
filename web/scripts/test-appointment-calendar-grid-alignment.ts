import fs from 'fs'
import path from 'path'

const source = fs.readFileSync(path.resolve(process.cwd(), 'src/pages/appointment/calendar.vue'), 'utf8')

function assertContains(pattern: RegExp, message: string) {
  if (!pattern.test(source)) {
    throw new Error(message)
  }
}

assertContains(/\.header-cell\s*\{[^}]*box-sizing:\s*border-box/s, 'header cells must include padding inside the fixed column width')
assertContains(/\.time-col\s*\{[^}]*flex:\s*0 0 var\(--time-col-width,\s*96rpx\)/s, 'time column must not flex-shrink or grow')
assertContains(/\.time-col\s*\{[^}]*box-sizing:\s*border-box/s, 'time column border/padding must not change its fixed width')
assertContains(/\.staff-col\s*\{[^}]*flex:\s*0 0 var\(--staff-col-width,\s*327rpx\)/s, 'staff columns must not flex-shrink or grow')
assertContains(/\.staff-col\s*\{[^}]*box-sizing:\s*border-box/s, 'staff column border/padding must not change its fixed width')
