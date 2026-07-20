"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = require("fs");
const path_1 = require("path");
const assert_1 = __importDefault(require("assert"));
const source = (0, fs_1.readFileSync)((0, path_1.resolve)(process.cwd(), 'src/pages/appointment/create.vue'), 'utf8');
(0, assert_1.default)(source.includes('function compareServicesByMonthlyUsage'), 'appointment service picker should use a named monthly usage comparator');
(0, assert_1.default)(source.includes('monthly_usage_count'), 'appointment service picker should sort services by monthly_usage_count from /b/services?order_by=monthly_usage');
(0, assert_1.default)(!source.includes('serviceRankingMap.value[b.name]'), 'appointment service picker should not sort by service name ranking, which can mix duplicate service names');
console.log('appointment service monthly usage sort checks passed');
