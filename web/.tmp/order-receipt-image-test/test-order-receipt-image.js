"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const filePath = node_path_1.default.resolve(__dirname, '../../src/pages/order/detail.vue');
const source = node_fs_1.default.readFileSync(filePath, 'utf8');
const saveUtilPath = node_path_1.default.resolve(__dirname, '../../src/utils/web-image-save.ts');
const saveUtilSource = node_fs_1.default.readFileSync(saveUtilPath, 'utf8');
(0, strict_1.default)(source.includes('<view class="receipt-wrap" v-if="!receiptImageUrl">'), 'receipt modal should hide the original receipt DOM after image generation to avoid duplicate preview content');
(0, strict_1.default)(source.includes('async function downloadReceiptImage()'), 'receipt save flow should be asynchronous so it can use share/open fallbacks on mobile Safari');
(0, strict_1.default)(saveUtilSource.includes('navigator.share') || saveUtilSource.includes('window.open('), 'receipt save flow should provide a mobile Safari fallback instead of relying only on a.download');
(0, strict_1.default)(source.includes('<view class="modal-mask" v-if="showReceipt" @click="closeReceipt">'), 'receipt modal should clear generated image state when dismissed from the mask');
console.log('order receipt image regression test passed');
