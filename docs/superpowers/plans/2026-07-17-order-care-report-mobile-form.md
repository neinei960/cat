# 护理报告手机端独立表单 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将护理报告从“点击缩小图片编辑”改为手机友好的独立分组表单，并把报告图片保留为默认收起的只读预览。

**Architecture:** `OrderCareReportModal.vue` 继续持有唯一的 `OrderCareReportDraft`，基本字段、体型和检查分组直接更新这份 draft；`OrderCareReportStage.vue` 只接收 draft 做只读映射。生成接口、payload、图片保存和裁剪上传协议保持不变。

**Tech Stack:** Vue 3 `<script setup lang="ts">`、uni-app H5、TypeScript 源码回归测试、现有 Go 护理报告接口。

---

## 文件结构

- Modify: `web/src/components/order/OrderCareReportModal.vue`
  - 独立表单、折叠检查分组、只读预览层、验证定位和移动端样式。
- Modify: `web/scripts/test-order-care-report-wysiwyg.ts`
  - 将旧的图片热点编辑断言改为独立表单和只读预览断言。
- Modify: `web/scripts/test-order-care-report-modal-style.ts`
  - 覆盖触控尺寸、单列布局、固定操作栏和安全区。
- Verify: `web/src/utils/order-care-report.ts`
  - 继续提供 draft、体型和检查项定义；不改变 payload 合约。
- Verify: `web/src/components/order/OrderCareReportStage.vue`
  - 保留当前模板绘制能力；弹窗不再启用其 `editable` 模式。

### Task 1: 建立独立表单行为回归测试

**Files:**
- Modify: `web/scripts/test-order-care-report-wysiwyg.ts`
- Test: `web/scripts/test-order-care-report-wysiwyg.ts`

- [ ] **Step 1: 用独立表单预期替换旧热点编辑断言**

将测试主体改为验证以下结构：

```ts
assert(modalSource.includes('care-report-basic-section'), 'modal should render a standalone basic-information form')
assert(modalSource.includes('orderCareReportBodyShapeOptions'), 'modal should reuse shared body-shape options')
assert(modalSource.includes('v-for="section in sectionDefinitions"'), 'modal should render all inspection sections from shared definitions')
assert(modalSource.includes('expandedSectionKey'), 'inspection sections should use a single accordion state')
assert(modalSource.includes('care-report-section-note'), 'each inspection section should expose its own note input')
assert(modalSource.includes('previewExpanded'), 'draft preview should be controlled independently')
assert(modalSource.includes('<OrderCareReportStage') && !modalSource.includes('@edit-target="openEditor"'), 'draft preview should be read only')
assert(!modalSource.includes('care-report-editor-dock'), 'contextual image editor should be removed')
assert(!modalSource.includes(':active-editor-key='), 'modal should not pass image editing state')
```

- [ ] **Step 2: 编译并运行测试，确认 RED**

Run:

```bash
cd web
rm -rf .tmp/order-care-report-wysiwyg-test
pnpm exec tsc scripts/test-order-care-report-wysiwyg.ts --outDir .tmp/order-care-report-wysiwyg-test --target ES2020 --module commonjs --moduleResolution node --esModuleInterop --skipLibCheck
node .tmp/order-care-report-wysiwyg-test/test-order-care-report-wysiwyg.js
```

Expected: FAIL，首个错误为缺少 `care-report-basic-section`。

### Task 2: 将图片热点编辑改为独立基本信息表单

**Files:**
- Modify: `web/src/components/order/OrderCareReportModal.vue`
- Test: `web/scripts/test-order-care-report-wysiwyg.ts`

- [ ] **Step 1: 移除弹窗内的热点编辑状态**

删除 `EditableFieldKey`、`StageEditTarget`、`ActiveEditor`、`activeEditor`、`activeEditorMarker`、`editorDescriptor`、`openEditor`、`closeEditor` 和只为编辑底栏服务的字段读写函数。保留上传、裁剪、提交和保存流程。

- [ ] **Step 2: 添加表单与预览状态**

```ts
import { computed, nextTick, ref, watch } from 'vue'
import {
  orderCareReportBodyShapeOptions,
  orderCareReportSectionDefinitions,
  type OrderCareReportSectionKey,
} from '@/utils/order-care-report'

const previewExpanded = ref(false)
const expandedSectionKey = ref<OrderCareReportSectionKey | null>('skin')

function togglePreview() {
  previewExpanded.value = !previewExpanded.value
}

function toggleSectionPanel(sectionKey: OrderCareReportSectionKey) {
  expandedSectionKey.value = expandedSectionKey.value === sectionKey ? null : sectionKey
}
```

在 `initializeDraft`、`resetState` 和 `selectPet` 中将预览恢复为关闭、检查分组恢复为 `skin`。

- [ ] **Step 3: 用单列基本信息表单替换可编辑 Stage**

表单包含照片、宝贝名字、品种、性别、年龄、护理内容、体重和两个日期选择器。文本输入统一使用现有事件更新方式：

```vue
<input
  :value="draft.petName"
  class="care-report-input"
  maxlength="20"
  placeholder="填写宝贝名字"
  @input="updateDraftTextField('petName', $event)"
/>
```

```ts
type TextDraftField = 'petName' | 'breed' | 'age' | 'careContent' | 'weight'

function updateDraftTextField(key: TextDraftField, event: any) {
  if (!draft.value) return
  draft.value[key] = String(event?.detail?.value || '')
}

function updateDraftDate(field: 'careDate' | 'nextCareDate', event: any) {
  if (!draft.value) return
  draft.value[field] = String(event?.detail?.value || '').replace(/-/g, '.')
}
```

照片区域点击继续调用 `choosePortrait`，已上传时显示圆形缩略图和“更换照片”，未上传时显示“上传照片”。

- [ ] **Step 4: 运行结构测试，确认基本表单相关断言继续推进**

Run Task 1 的测试命令。

Expected: 基本信息断言通过，检查分组或预览断言仍可能失败。

### Task 3: 实现体型选择和检查项折叠表单

**Files:**
- Modify: `web/src/components/order/OrderCareReportModal.vue`
- Test: `web/scripts/test-order-care-report-wysiwyg.ts`

- [ ] **Step 1: 渲染共享体型选项**

```vue
<view class="care-report-choice-grid body-shape">
  <view
    v-for="option in bodyShapeOptions"
    :key="option.value"
    :class="['care-report-choice', draft.bodyShape === option.value ? 'active' : '']"
    @click="setBodyShape(option.value)"
  >
    {{ option.label }}
  </view>
</view>
```

`setBodyShape` 为单选，不再点击已选项后清空，避免误触导致必填项丢失。

```ts
const bodyShapeOptions = orderCareReportBodyShapeOptions

function setBodyShape(value: string) {
  if (!draft.value) return
  draft.value.bodyShape = value
}
```

- [ ] **Step 2: 渲染七个折叠检查分组**

```vue
<view v-for="section in sectionDefinitions" :key="section.key" class="care-report-inspection-section">
  <view class="care-report-section-head" @click="toggleSectionPanel(section.key)">
    <text>{{ section.label }}</text>
    <text>{{ getSectionSelectedCount(section.key) }}项</text>
  </view>
  <view v-if="expandedSectionKey === section.key" class="care-report-section-body">
    <view class="care-report-choice-grid">
      <view
        v-for="option in section.options"
        :key="option.value"
        :class="['care-report-choice', isSectionCheckSelected(section.key, option.value) ? 'active' : '']"
        @click="toggleSectionCheck(section.key, option.value)"
      >
        {{ option.label }}
      </view>
    </view>
    <textarea
      :value="getSectionNote(section.key)"
      class="care-report-textarea care-report-section-note"
      maxlength="120"
      auto-height
      placeholder="备注（选填）"
      @input="updateSectionNote(section.key, $event)"
    />
  </view>
</view>
```

补齐折叠和选择辅助函数：

```ts
function getSectionSelectedCount(sectionKey: OrderCareReportSectionKey) {
  return draft.value?.[sectionKey]?.checks.length || 0
}

function isSectionCheckSelected(sectionKey: OrderCareReportSectionKey, value: string) {
  return Boolean(draft.value?.[sectionKey]?.checks.includes(value))
}

function toggleSectionPanel(sectionKey: OrderCareReportSectionKey) {
  expandedSectionKey.value = expandedSectionKey.value === sectionKey ? null : sectionKey
}
```

- [ ] **Step 3: 运行结构测试，确认 GREEN**

Run Task 1 的测试命令。

Expected: PASS。

### Task 4: 建立并实现只读预览与验证定位

**Files:**
- Modify: `web/scripts/test-order-care-report-wysiwyg.ts`
- Modify: `web/src/components/order/OrderCareReportModal.vue`

- [ ] **Step 1: 添加预览默认关闭和验证定位断言**

```ts
assert(modalSource.includes('const previewExpanded = ref(false)'), 'draft preview should be closed by default')
assert(modalSource.includes('v-if="previewExpanded"'), 'draft preview should render only on demand')
assert(modalSource.includes(':draft="draft"') && !modalSource.includes('<OrderCareReportStage\n              :draft="draft"\n              editable'), 'preview stage should not be editable')
assert(modalSource.includes('scrollIntoView'), 'validation should bring the first missing field into view')
```

- [ ] **Step 2: 运行测试，确认 RED**

Run Task 1 的测试命令。

Expected: FAIL，缺少预览层或验证滚动实现。

- [ ] **Step 3: 添加只读预览层**

```vue
<view v-if="previewExpanded" class="care-report-draft-preview">
  <view class="care-report-draft-preview-head">
    <text class="care-report-title">报告预览</text>
    <text class="care-report-close" @click="togglePreview">返回填写</text>
  </view>
  <view class="care-report-draft-preview-scroll">
    <OrderCareReportStage :draft="draft" />
  </view>
</view>
```

底部操作栏改为“预览报告”和“生成报告”两个按钮；预览只切换 `previewExpanded`，不调用后端。

- [ ] **Step 4: 为必填项添加锚点和定位函数**

```ts
function focusValidationTarget(anchorId: string) {
  previewExpanded.value = false
  nextTick(() => {
    if (typeof document === 'undefined') return
    document.getElementById(anchorId)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}
```

`validateDraft` 使用明确的顺序和锚点，并保留当前 Toast 文案：

```ts
const requiredChecks = [
  { valid: () => Boolean(draft.value?.portraitUrl), anchor: 'care-report-anchor-portrait', message: '请先上传护理照片' },
  { valid: () => Boolean(draft.value?.careContent.trim()), anchor: 'care-report-anchor-care-content', message: '请填写护理内容' },
  { valid: () => Boolean(draft.value?.weight.trim()), anchor: 'care-report-anchor-weight', message: '请填写体重' },
  { valid: () => Boolean(draft.value?.careDate), anchor: 'care-report-anchor-care-date', message: '请填写护理日期' },
  { valid: () => Boolean(draft.value?.nextCareDate), anchor: 'care-report-anchor-next-care-date', message: '请填写建议下次护理日期' },
  { valid: () => Boolean(draft.value?.bodyShape), anchor: 'care-report-anchor-body-shape', message: '请选择体型' },
]

for (const check of requiredChecks) {
  if (check.valid()) continue
  focusValidationTarget(check.anchor)
  uni.showToast({ title: check.message, icon: 'none' })
  return false
}
return true
```

- [ ] **Step 5: 运行结构测试，确认 GREEN**

Run Task 1 的测试命令。

Expected: PASS。

### Task 5: 建立并实现移动端样式约束

**Files:**
- Modify: `web/scripts/test-order-care-report-modal-style.ts`
- Modify: `web/src/components/order/OrderCareReportModal.vue`

- [ ] **Step 1: 添加移动端样式失败断言**

```ts
assertMatches(source, /\.care-report-form-row[\s\S]*?display:\s*grid;[\s\S]*?grid-template-columns:\s*1fr;/, 'form rows should stay single-column on mobile')
assertMatches(source, /\.care-report-choice[\s\S]*?min-height:\s*88rpx;/, 'choices should provide a 44px touch target')
assertContains(source, 'env(safe-area-inset-bottom)', 'bottom actions should reserve the mobile safe area')
assertMatches(source, /\.care-report-draft-preview[\s\S]*?position:\s*absolute;/, 'draft preview should be a dedicated overlay')
assertNotContains(source, '.care-report-editor-dock', 'old contextual editor styles should be removed')
```

- [ ] **Step 2: 运行样式测试，确认 RED**

```bash
cd web
rm -rf .tmp/order-care-report-modal-style-test
pnpm exec tsc scripts/test-order-care-report-modal-style.ts --outDir .tmp/order-care-report-modal-style-test --target ES2020 --module commonjs --moduleResolution node --esModuleInterop --skipLibCheck
node .tmp/order-care-report-modal-style-test/test-order-care-report-modal-style.js
```

Expected: FAIL，缺少单列表单或预览层样式。

- [ ] **Step 3: 实现移动端样式**

样式要求：

```css
.care-report-form-row {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16rpx;
}

.care-report-choice {
  min-height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.care-report-actions {
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
}

.care-report-draft-preview {
  position: absolute;
  inset: 0;
  z-index: 4;
  display: flex;
  flex-direction: column;
  background: #fff9f0;
}
```

检查分组使用全宽分隔带而非卡片嵌套；选项网格使用 `repeat(3, minmax(0, 1fr))`，长文字允许换行。

- [ ] **Step 4: 运行样式和结构测试，确认 GREEN**

Run Task 1 和 Task 5 的测试命令。

Expected: 两个测试均 PASS。

### Task 6: 回归、部署和手机浏览器验收

**Files:**
- Verify: `web/src/components/order/OrderCareReportModal.vue`
- Verify: `web/src/components/order/OrderCareReportStage.vue`
- Verify: `web/src/utils/order-care-report.ts`

- [ ] **Step 1: 运行护理报告全部前端测试**

```bash
cd web
rm -rf .tmp/order-care-report-test
pnpm exec tsc src/utils/order-care-report.ts src/utils/web-image-save.ts scripts/test-order-care-report.ts --outDir .tmp/order-care-report-test --target ES2020 --module commonjs --moduleResolution node --esModuleInterop --types node --skipLibCheck
node .tmp/order-care-report-test/scripts/test-order-care-report.js

rm -rf .tmp/order-care-report-frontend-test
pnpm exec tsc scripts/test-order-care-report-frontend.ts --outDir .tmp/order-care-report-frontend-test --target ES2020 --module commonjs --moduleResolution node --esModuleInterop --skipLibCheck
node .tmp/order-care-report-frontend-test/test-order-care-report-frontend.js
```

Expected: 全部 PASS，payload 合约无变化。

- [ ] **Step 2: 构建 H5**

```bash
cd web
pnpm build:h5
```

Expected: `DONE Build complete.`

- [ ] **Step 3: 部署前端**

```bash
printf '{"tool_input":{"file_path":"/Users/genglsh/workstation/cat/cat/web/src/components/order/OrderCareReportModal.vue"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

Expected: 构建和远端同步成功。

- [ ] **Step 4: 使用 Playwright 验收手机视口**

在 `390x844` 视口打开订单详情并进入“生成护理报告”，确认：

- 页面默认显示独立表单，没有图片热点。
- 预览默认不出现。
- 基本字段、体型和检查选项的点击区域可正常操作。
- 同一时间只展开一个检查分组。
- 点击“预览报告”后出现只读报告，点击“返回填写”数据仍保留。
- 页面宽度没有横向溢出，底部操作栏不遮挡最后一个输入区。

- [ ] **Step 5: 检查远端资源和控制台**

确认订单详情 JS 的本地/远端哈希一致，浏览器控制台没有本次改动引入的运行时错误。
