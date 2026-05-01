"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const createFilePath = node_path_1.default.resolve(__dirname, '../../src/pages/boarding/create.vue');
const detailFilePath = node_path_1.default.resolve(__dirname, '../../src/pages/boarding/detail.vue');
const createSource = node_fs_1.default.readFileSync(createFilePath, 'utf8');
const detailSource = node_fs_1.default.readFileSync(detailFilePath, 'utf8');
(0, strict_1.default)(createSource.includes('暂不填写'), 'boarding create page should expose an explicit optional deworming choice');
(0, strict_1.default)(!createSource.includes("title: '请选择是否已驱虫'"), 'boarding create page should not block submission when deworming is unset');
(0, strict_1.default)(createSource.includes('hasDeworming: null as boolean | null'), 'boarding create page should preserve a nullable deworming state');
(0, strict_1.default)(createSource.includes('has_deworming: form.value.hasDeworming'), 'boarding create submit payload should keep sending nullable deworming data');
(0, strict_1.default)(detailSource.includes("return '未填写'"), 'boarding detail should render a neutral label for unset deworming state');
console.log('boarding deworming optional regression test passed');
