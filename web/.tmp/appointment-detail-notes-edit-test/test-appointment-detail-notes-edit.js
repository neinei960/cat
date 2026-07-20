"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const strict_1 = __importDefault(require("node:assert/strict"));
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const pagePath = node_path_1.default.resolve(__dirname, '../../src/pages/appointment/detail.vue');
const source = node_fs_1.default.readFileSync(pagePath, 'utf8');
(0, strict_1.default)(source.includes('updateAppointmentNotes'), 'appointment detail should use the dedicated notes update API');
(0, strict_1.default)(source.includes('v-model="notesDraft"') && source.includes('class="notes-textarea"'), 'appointment detail notes card should render an editable textarea bound to notesDraft');
(0, strict_1.default)(source.includes('async function saveNotesEdit()'), 'appointment detail should have an inline save handler for notes edits');
(0, strict_1.default)(source.includes('保存备注'), 'appointment detail should expose a direct save action for notes');
console.log('appointment detail notes edit checks passed');
