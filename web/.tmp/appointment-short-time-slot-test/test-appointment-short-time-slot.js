"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const node_fs_1 = require("node:fs");
const node_path_1 = require("node:path");
function assert(condition, message) {
    if (!condition) {
        throw new Error(message);
    }
}
const pagePath = (0, node_path_1.resolve)(process.cwd(), 'src/pages/appointment/create.vue');
const source = (0, node_fs_1.readFileSync)(pagePath, 'utf8');
assert(source.includes('const shortCareDurations'), 'appointment create should define short care duration shortcuts');
assert(source.includes('刷牙 / 剪指甲'), 'appointment create should label short care as brushing / nail trim');
assert(source.includes('selectShortCareEndTime'), 'appointment create should provide a quick short-care end-time setter');
assert(source.includes('isShortCareEndTimeValid'), 'appointment create should accept selected short-care end times as valid');
assert(source.includes('@click="selectShortCareEndTime(option.minutes)"'), 'short-care buttons should set the end time');
assert(source.includes('appointment-shortcuts'), 'short-care shortcut controls should be rendered near end-time selection');
console.log('appointment short time slot checks passed');
