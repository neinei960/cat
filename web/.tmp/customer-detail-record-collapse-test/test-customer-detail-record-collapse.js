"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const filePath = node_path_1.default.resolve(__dirname, '../../src/pages/customer/detail.vue');
const source = node_fs_1.default.readFileSync(filePath, 'utf8');
(0, strict_1.default)(source.includes('records.value.slice(0, 1)'), 'collapsed customer detail records should keep only the latest record visible');
(0, strict_1.default)(!source.includes('records.value.slice(0, 3)'), 'customer detail records should no longer default to three visible rows');
(0, strict_1.default)(source.includes('records-arrow'), 'customer detail records should render an arrow-style toggle control');
console.log('customer detail record collapse regression test passed');
