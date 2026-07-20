"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = require("fs");
const path_1 = require("path");
function assertIncludes(source, expected, label) {
    if (!source.includes(expected)) {
        throw new Error(`${label}: expected to find "${expected}"`);
    }
}
function main() {
    const file = (0, fs_1.readFileSync)((0, path_1.resolve)(process.cwd(), 'src/pages/order/detail.vue'), 'utf8');
    const saveUtil = (0, fs_1.readFileSync)((0, path_1.resolve)(process.cwd(), 'src/utils/web-image-save.ts'), 'utf8');
    assertIncludes(file, 'const isAppleSafari = computed(() => isAppleSafariBrowser())', 'receipt.ios.detect');
    assertIncludes(file, 'const showNativeReceiptImage = computed(() => true)', 'receipt.native.all-mobile');
    assertIncludes(file, 'const receiptPreviewSrc = computed(() => {', 'receipt.preview.src');
    assertIncludes(file, 'return isAppleSafari.value ? receiptImageUrl.value : receiptBlobUrl.value || receiptImageUrl.value', 'receipt.preview.prefers.data-url');
    assertIncludes(file, 'const receiptImageHint = computed(() => {', 'receipt.image.hint');
    assertIncludes(file, '如未出现菜单，点击「保存图片」后在新页面长按', 'receipt.image.hint.ios');
    assertIncludes(file, '<img', 'receipt.image.native-tag');
    assertIncludes(file, 'class="receipt-image receipt-image-native"', 'receipt.image.native-class');
    assertIncludes(file, '@click.stop', 'receipt.image.click.stop');
    assertIncludes(file, '@touchstart.stop', 'receipt.image.touch.stop');
    assertIncludes(saveUtil, 'function isMobileImageSaveFallbackBrowser()', 'receipt.mobile.detect');
    assertIncludes(saveUtil, "isMobileImageSaveFallbackBrowser() && openImagePreviewWindow(src, options.title || fileName)", 'receipt.mobile.preview.fallback');
    assertIncludes(file, "uni.showToast({ title: '新页面已打开，请长按图片保存', icon: 'none' })", 'receipt.preview.window.toast');
    assertIncludes(file, '-webkit-touch-callout: default;', 'receipt.image.native.touch-callout');
}
main();
console.log('order receipt ios save checks passed');
