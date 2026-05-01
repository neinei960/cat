"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const filePath = node_path_1.default.resolve(__dirname, '../../src/pages/customer/edit.vue');
const source = node_fs_1.default.readFileSync(filePath, 'utf8');
(0, strict_1.default)(!source.includes('contenteditable="plaintext-only"'), 'customer edit remark field should not rely on H5 contenteditable="plaintext-only"; use native textarea for mobile browser editing');
(0, strict_1.default)(source.includes('<textarea v-model="form.remark" placeholder="添加文字" class="field-textarea" :auto-height="false" />'), 'customer edit remark field should render the shared textarea control');
console.log('customer edit remark regression test passed');
