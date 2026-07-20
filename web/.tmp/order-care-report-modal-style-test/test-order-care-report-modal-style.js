"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const node_fs_1 = require("node:fs");
const node_path_1 = require("node:path");
const filePath = (0, node_path_1.resolve)(new URL('.', import.meta.url).pathname, '../src/components/order/OrderCareReportModal.vue');
const source = (0, node_fs_1.readFileSync)(filePath, 'utf8');
assertNotContains(source, '导出图片会以这张预览为准', 'stage preview tip should be removed');
assertMatches(source, /\.care-report-input\s*\{[\s\S]*?display:\s*flex;[\s\S]*?min-height:\s*\d+rpx;/, 'input shell should define a real flex height');
assertContains(source, '.care-report-input :deep(.uni-input-wrapper)', 'missing deep wrapper selector');
assertContains(source, '.care-report-input :deep(.uni-input-input)', 'missing deep input selector');
assertContains(source, '.care-report-input :deep(.uni-input-placeholder)', 'missing deep placeholder selector');
assertMatches(source, /\.care-report-input\s*:deep\(\.uni-input-wrapper\)[\s\S]*?(min-height:\s*100%|height:\s*100%)/, 'wrapper selector should fill the input shell height');
assertMatches(source, /\.care-report-input\s*:deep\(\.uni-input-input\)[\s\S]*?(min-height|height):\s*\d+rpx;/, 'input selector should define height');
assertMatches(source, /\.care-report-input\s*:deep\(\.uni-input-input\)[\s\S]*?line-height:\s*\d+rpx;/, 'input selector should define line-height');
assertMatches(source, /\.care-report-form-row\s*\{[\s\S]*?grid-template-columns:\s*1fr;/, 'mobile form rows should use one column');
assertMatches(source, /\.care-report-choice\s*\{[\s\S]*?min-height:\s*88rpx;/, 'form choices should have a mobile-friendly touch target');
assertContains(source, 'env(safe-area-inset-bottom)', 'bottom actions should respect the mobile safe area');
assertMatches(source, /\.care-report-draft-preview\s*\{[\s\S]*?position:\s*absolute;[\s\S]*?inset:\s*0;/, 'read-only preview should cover the modal without changing form layout');
assertNotContains(source, '.care-report-editor-dock', 'legacy canvas editor dock should stay removed');
function assertContains(content, token, message) {
    if (!content.includes(token)) {
        throw new Error(message);
    }
}
function assertNotContains(content, token, message) {
    if (content.includes(token)) {
        throw new Error(message);
    }
}
function assertMatches(content, pattern, message) {
    if (!pattern.test(content)) {
        throw new Error(message);
    }
}
