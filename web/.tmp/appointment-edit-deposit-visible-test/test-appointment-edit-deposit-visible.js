"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = require("fs");
const path_1 = require("path");
const assert_1 = __importDefault(require("assert"));
const source = (0, fs_1.readFileSync)((0, path_1.resolve)(process.cwd(), 'src/pages/appointment/create.vue'), 'utf8');
const stepsIndex = source.indexOf('<!-- Step indicator -->');
const editDepositIndex = source.indexOf('class="card edit-deposit-card"');
const stepOneIndex = source.indexOf('<!-- Step 1: Customer & Pet -->');
(0, assert_1.default)(editDepositIndex > stepsIndex, 'edit deposit card should render after the step indicator');
(0, assert_1.default)(editDepositIndex < stepOneIndex, 'edit deposit card should render before step content so it is visible while editing');
(0, assert_1.default)(source.includes('添加/修改预约金'), 'edit deposit card should clearly say it can add or modify the deposit');
(0, assert_1.default)(source.includes('deposit: form.value.deposit'), 'appointment update payload should include deposit');
console.log('appointment edit deposit visibility checks passed');
