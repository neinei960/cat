"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = require("fs");
const path_1 = require("path");
const assert_1 = __importDefault(require("assert"));
const repoRoot = (0, path_1.resolve)(process.cwd(), '..');
const pageSource = (0, fs_1.readFileSync)((0, path_1.resolve)(repoRoot, 'web/src/pages/appointment/detail.vue'), 'utf8');
const routerSource = (0, fs_1.readFileSync)((0, path_1.resolve)(repoRoot, 'server/internal/router/router.go'), 'utf8');
(0, assert_1.default)(pageSource.includes('editingRecordId'), 'appointment detail should keep editingRecordId state for service record editing');
(0, assert_1.default)(pageSource.includes('function openRecordEdit'), 'appointment detail should allow opening an existing service record for edit');
(0, assert_1.default)(pageSource.includes("method: editingRecordId.value ? 'PUT' : 'POST'"), 'appointment detail should submit edits with PUT and new records with POST');
(0, assert_1.default)(routerSource.includes('b.PUT("/service-records/:id"'), 'server should expose PUT /service-records/:id for editing service records');
console.log('appointment detail service record edit checks passed');
