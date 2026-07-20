"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const pagePath = node_path_1.default.resolve(process.cwd(), 'src/pages/product/list.vue');
const source = node_fs_1.default.readFileSync(pagePath, 'utf8');
(0, strict_1.default)(source.includes('<text v-if="!keyword" class="search-placeholder">搜索商品名 / 品牌</text>'), 'product list search should render placeholder only when keyword is empty');
(0, strict_1.default)(!source.includes('placeholder="搜索商品名 / 品牌"'), 'product list search should not use native input placeholder that overlaps on iOS H5');
(0, strict_1.default)(source.includes('pointer-events: none'), 'custom placeholder should not block input focus');
console.log('product list search placeholder checks passed');
