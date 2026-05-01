"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = __importDefault(require("fs"));
const path_1 = __importDefault(require("path"));
const source = fs_1.default.readFileSync(path_1.default.resolve(process.cwd(), 'src/pages/appointment/calendar.vue'), 'utf8');
function assertContains(pattern, message) {
    if (!pattern.test(source)) {
        throw new Error(message);
    }
}
assertContains(/\.header-cell\s*\{[^}]*box-sizing:\s*border-box/s, 'header cells must include padding inside the fixed column width');
assertContains(/\.time-col\s*\{[^}]*flex:\s*0 0 var\(--time-col-width,\s*96rpx\)/s, 'time column must not flex-shrink or grow');
assertContains(/\.time-col\s*\{[^}]*box-sizing:\s*border-box/s, 'time column border/padding must not change its fixed width');
assertContains(/\.staff-col\s*\{[^}]*flex:\s*0 0 var\(--staff-col-width,\s*327rpx\)/s, 'staff columns must not flex-shrink or grow');
assertContains(/\.staff-col\s*\{[^}]*box-sizing:\s*border-box/s, 'staff column border/padding must not change its fixed width');
